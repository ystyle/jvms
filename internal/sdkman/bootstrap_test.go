//go:build !windows

package sdkman

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/ystyle/jvms/internal/config"
)

func TestEnsureDoesNotAutoInstallSDKMANWithoutPermission(t *testing.T) {
	downloaded := false
	manager := &service{
		config: &appconfig.Config{SDKMANInstallURL: appconfig.DefaultSDKMANInstallURL},
		dir:    t.TempDir(),
		download: func(string) ([]byte, error) {
			downloaded = true
			return nil, nil
		},
	}

	err := manager.Ensure()
	if err == nil || !strings.Contains(err.Error(), "enable automatic installation") {
		t.Fatalf("Ensure() error = %v", err)
	}

	if downloaded {
		t.Fatal("Ensure downloaded the SDKMAN installer while automatic installation was disabled")
	}
}

func TestDownloadInstallerUsesConfiguredURL(t *testing.T) {
	const wantURL = "https://example.com/sdkman.sh"

	manager := &service{
		config: &appconfig.Config{SDKMANInstallURL: wantURL},
		download: func(gotURL string) ([]byte, error) {
			if gotURL != wantURL {
				t.Fatalf("download URL = %q, want %q", gotURL, wantURL)
			}

			return []byte("#!/usr/bin/env bash\n"), nil
		},
	}

	installer, err := manager.downloadInstaller()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(installer)
}

func TestEnsureBootstrapsAndReusesSDKMAN(t *testing.T) {
	installer := `mkdir -p "$SDKMAN_DIR/bin"
cat > "$SDKMAN_DIR/bin/sdkman-init.sh" <<'INIT'
sdk() {
    [ "$1" = version ] || return 1
    printf '%s\n' "$1" >> "$SDKMAN_DIR/health-runs"
    printf '%s\n' 'SDKMAN! test installation'
}
INIT
`
	manager := bootstrapTestService(t, installer)
	downloads := 0
	manager.download = func(string) ([]byte, error) {
		downloads++
		return []byte(installer), nil
	}

	if err := manager.Ensure(); err != nil {
		t.Fatal(err)
	}

	// An existing installation must work even after bootstrap permission is revoked.
	manager.config.AllowAutoInstallSDKMAN = false
	if err := manager.Ensure(); err != nil {
		t.Fatal(err)
	}

	if downloads != 1 {
		t.Fatalf("download count = %d, want 1", downloads)
	}

	entries, err := os.ReadDir(os.Getenv("TMPDIR"))
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Fatalf("bootstrap left temporary files: %v", entries)
	}

	healthRuns, err := os.ReadFile(filepath.Join(manager.dir, "health-runs"))
	if err != nil {
		t.Fatal(err)
	}

	if string(healthRuns) != "version\nversion\n" {
		t.Fatalf("health checks = %q, want one per Ensure call", healthRuns)
	}
}

func TestEnsureReportsBootstrapFailuresAndCleansUp(t *testing.T) {
	tests := []struct {
		name      string
		installer string
		wantError string
	}{
		{
			name:      "installer exits unsuccessfully",
			installer: "exit 7\n",
			wantError: "SDKMAN installer failed",
		},
		{
			name:      "installer leaves no init script",
			installer: "exit 0\n",
			wantError: "SDKMAN was installed but is not working",
		},
		{
			name: "health command fails",
			installer: `mkdir -p "$SDKMAN_DIR/bin"
cat > "$SDKMAN_DIR/bin/sdkman-init.sh" <<'INIT'
sdk() {
    printf '%s\n' 'health probe failed' >&2
    return 9
}
INIT
`,
			wantError: "health probe failed",
		},
		{
			name: "health response is unexpected",
			installer: `mkdir -p "$SDKMAN_DIR/bin"
cat > "$SDKMAN_DIR/bin/sdkman-init.sh" <<'INIT'
sdk() {
    printf '%s\n' 'unexpected version response'
}
INIT
`,
			wantError: "health check returned an unexpected response",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := bootstrapTestService(t, test.installer)
			err := manager.Ensure()
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Ensure() error = %v, want %q", err, test.wantError)
			}

			entries, err := os.ReadDir(os.Getenv("TMPDIR"))
			if err != nil {
				t.Fatal(err)
			}

			if len(entries) != 0 {
				t.Fatalf("bootstrap left temporary files: %v", entries)
			}
		})
	}
}

func TestEnsureDoesNotReinstallAnUnhealthySDKMAN(t *testing.T) {
	manager := bootstrapTestService(t, "")
	if err := os.MkdirAll(filepath.Dir(manager.initScript()), 0755); err != nil {
		t.Fatal(err)
	}

	script := `sdk() {
    printf '%s\n' 'broken SDKMAN installation' >&2
    return 1
}`
	if err := os.WriteFile(manager.initScript(), []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	manager.download = func(string) ([]byte, error) {
		t.Fatal("Ensure tried to overwrite an existing unhealthy installation")
		return nil, nil
	}

	err := manager.Ensure()
	if err == nil || !strings.Contains(err.Error(), "broken SDKMAN installation") {
		t.Fatalf("Ensure() error = %v, want the health-check failure", err)
	}
}

func bootstrapTestService(t *testing.T, installer string) *service {
	t.Helper()

	if os.Geteuid() == 0 {
		t.Skip("SDKMAN bootstrap requires a regular user")
	}

	// Real Bash executes local fixtures; download tools only satisfy prerequisite checks.
	toolsDir := t.TempDir()
	for _, name := range []string{"bash", "mkdir", "cat"} {
		path, err := exec.LookPath(name)
		if err != nil {
			t.Fatal(err)
		}

		if err := os.Symlink(path, filepath.Join(toolsDir, name)); err != nil {
			t.Fatal(err)
		}
	}

	for _, name := range []string{"curl", "zip", "unzip"} {
		if err := os.WriteFile(filepath.Join(toolsDir, name), []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("PATH", toolsDir)
	t.Setenv("TMPDIR", t.TempDir())
	t.Setenv("BASH_ENV", "")
	t.Setenv("SDKMAN_DIR", filepath.Join(t.TempDir(), "wrong-sdkman"))

	config := appconfig.NewConfig()
	config.AllowAutoInstallSDKMAN = true

	return &service{
		config: config,
		dir:    filepath.Join(t.TempDir(), "sdk man"),
		download: func(string) ([]byte, error) {
			return []byte(installer), nil
		},
	}
}
