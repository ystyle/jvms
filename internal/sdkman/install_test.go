//go:build !windows

package sdkman

import (
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/ystyle/jvms/internal/config"
)

func TestInstallRecordsSDKMANCurrentSelection(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("SDKMAN installation requires a regular user")
	}

	for _, test := range []struct {
		name        string
		mode        string
		wantCurrent string
		wantError   bool
	}{
		{
			name:        "new default",
			mode:        "select",
			wantCurrent: "21.0.8-tem",
		},
		{
			name:        "existing default retained",
			mode:        "retain",
			wantCurrent: "17.0.16-tem",
		},
		{
			name: "no default",
			mode: "clear",
		},
		{
			name:        "default changed before failure",
			mode:        "fail",
			wantCurrent: "21.0.8-tem",
			wantError:   true,
		},
		{
			name:        "current selection cannot be read",
			mode:        "unreadable",
			wantCurrent: "stale-config",
			wantError:   true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"bin", "candidates/java/17.0.16-tem/bin"} {
				if err := os.MkdirAll(filepath.Join(dir, name), 0755); err != nil {
					t.Fatal(err)
				}
			}

			if err := os.Symlink("17.0.16-tem", filepath.Join(dir, "candidates/java/current")); err != nil {
				t.Fatal(err)
			}

			script := `sdk() {
    [ "$1" = install ] && [ "$2" = java ] && [ "$3" = 21.0.8-tem ] || return 9
    local candidates="$SDKMAN_DIR/candidates/java"
    mkdir -p "$candidates/$3/bin"
    touch "$candidates/$3/bin/java"
    case "$JVMS_SDKMAN_TEST_INSTALL_MODE" in
        retain) return 0 ;;
        clear) unlink "$candidates/current"; return 0 ;;
        unreadable) unlink "$candidates/current"; mkdir "$candidates/current"; return 0 ;;
    esac
    ln -sfn "$3" "$candidates/current"
    [ "$JVMS_SDKMAN_TEST_INSTALL_MODE" != fail ] || return 7
}`
			if err := os.WriteFile(filepath.Join(dir, "bin/sdkman-init.sh"), []byte(script), 0644); err != nil {
				t.Fatal(err)
			}

			t.Setenv("SDKMAN_DIR", dir)
			t.Setenv("BASH_ENV", "")
			t.Setenv("JVMS_SDKMAN_TEST_INSTALL_MODE", test.mode)
			config := appconfig.NewConfig()
			config.CurrentJDKVersion = "stale-config"
			manager := NewService(config)

			err := manager.Install("21.0.8-tem")
			if (err != nil) != test.wantError {
				t.Fatalf("Install() error = %v, want error %t", err, test.wantError)
			}

			if config.CurrentJDKVersion != test.wantCurrent {
				t.Fatalf("current version = %q, want %q", config.CurrentJDKVersion, test.wantCurrent)
			}
		})
	}
}
