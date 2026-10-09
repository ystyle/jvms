//go:build !windows

package sdkman

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAvailableRequiresSDKMAN(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	manager := &service{dir: t.TempDir()}
	_, err := manager.Available()
	if err == nil || !strings.Contains(err.Error(), "run `jvms init`") {
		t.Fatalf("Available() error = %v, want init guidance", err)
	}
}

func TestReplaceEnvRemovesOlderValues(t *testing.T) {
	got := replaceEnv(
		[]string{"PATH=/bin", "SDKMAN_DIR=/wrong", "SDKMAN_DIR=/also-wrong"},
		map[string]string{"SDKMAN_DIR": "/right"},
	)

	wantCount := 0
	for _, entry := range got {
		if entry == "SDKMAN_DIR=/right" {
			wantCount++
		}

		if strings.HasPrefix(entry, "SDKMAN_DIR=") && entry != "SDKMAN_DIR=/right" {
			t.Fatalf("replaceEnv retained stale value %q", entry)
		}
	}

	if wantCount != 1 {
		t.Fatalf("replacement count = %d, want 1", wantCount)
	}
}

func TestAvailableSupportsSDKMANDirectoryWithSpaces(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := filepath.Join(t.TempDir(), "sdk man")
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
		t.Fatal(err)
	}

	script := `sdk() {
  printf '%s\n' 'Temurin | | 21.0.8 | tem | | 21.0.8-tem'
}`
	if err := os.WriteFile(filepath.Join(dir, "bin", "sdkman-init.sh"), []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	versions, err := (&service{dir: dir}).Available()
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 1 || versions[0].Version != "21.0.8-tem" {
		t.Fatalf("Available() = %#v", versions)
	}
}
