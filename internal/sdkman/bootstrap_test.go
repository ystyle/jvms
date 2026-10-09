//go:build !windows

package sdkman

import (
	"os"
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
