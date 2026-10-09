//go:build !windows

package sdkman

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/tucnak/store"
	appconfig "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

func TestRefreshAvailablePreservesCacheOnFailure(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "offline response",
			script: `printf '%s\n' 'INTERNET NOT REACHABLE' 'Offline: only showing installed java versions' ' > 21.0.8-tem'`,
		},
		{
			name:   "command failure",
			script: `echo 'API request failed' >&2; return 1`,
		},
		{
			name:   "empty catalog",
			script: `echo 'No Java versions available'`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := catalogTestService(t, test.script)
			want := []jdk.Version{
				{
					Version: "17.0.16-tem",
					Url:     "Temurin",
				},
			}

			store.Init(appconfig.ProjectConfigDir)
			if err := store.Save("jdk_versions.json", &jdk.JdkVersionCache{
				Versions:    want,
				LastUpdated: time.Now().Add(-48 * time.Hour).Unix(),
				Source:      sdkmanCacheSource,
			}); err != nil {
				t.Fatal(err)
			}

			cachePath := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), appconfig.ProjectConfigDir, "jdk_versions.json")
			before, err := os.ReadFile(cachePath)
			if err != nil {
				t.Fatal(err)
			}

			got, err := manager.RefreshAvailable()
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("RefreshAvailable() = %#v, %v; want %#v", got, err, want)
			}

			after, err := os.ReadFile(cachePath)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(before, after) {
				t.Fatal("failed refresh changed the saved catalog or its timestamp")
			}
		})
	}
}

func TestRefreshAvailableReplacesCacheAfterSuccess(t *testing.T) {
	manager := catalogTestService(t, `printf '%s\n' 'Temurin | | 21.0.8 | tem | | 21.0.8-tem'`)
	previous := []jdk.Version{
		{
			Version: "17.0.16-tem",
			Url:     "Temurin",
		},
	}
	if err := jdk.CacheVersions(manager.config, sdkmanCacheSource, previous); err != nil {
		t.Fatal(err)
	}

	got, err := manager.Available()
	if err != nil || !reflect.DeepEqual(got, previous) {
		t.Fatalf("Available() = %#v, %v; want cached %#v", got, err, previous)
	}

	want := []jdk.Version{
		{
			Version: "21.0.8-tem",
			Url:     "Temurin",
		},
	}

	got, err = manager.RefreshAvailable()
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("RefreshAvailable() = %#v, %v; want fresh %#v", got, err, want)
	}

	got, err = jdk.LoadCachedVersions(manager.config, sdkmanCacheSource, false)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("saved catalog = %#v, %v; want %#v", got, err, want)
	}
}

func catalogTestService(t *testing.T, body string) *service {
	t.Helper()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
		t.Fatal(err)
	}

	script := "sdk() {\n" + body + "\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "bin", "sdkman-init.sh"), []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	return &service{
		config: appconfig.NewConfig(),
		dir:    dir,
	}
}
