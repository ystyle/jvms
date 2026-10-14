//go:build !windows

package platform

import (
	"testing"

	appcfg "github.com/ystyle/jvms/internal/config"
)

func TestUnixProviderUsesSDKMAN(t *testing.T) {
	provider := NewProvider(appcfg.NewConfig())
	if provider.Name() != "SDKMAN!" {
		t.Fatalf("provider = %q, want SDKMAN!", provider.Name())
	}
}

func TestSDKMANInstallWarningCanBeDismissed(t *testing.T) {
	config := appcfg.NewConfig()
	provider := NewProvider(config).(*sdkmanManager)
	if provider.InstallWarning() == "" {
		t.Fatal("SDKMAN install warning should be visible by default")
	}

	config.DismissSDKInstallWarning = true
	if provider.InstallWarning() != "" {
		t.Fatal("SDKMAN install warning was not dismissed")
	}

	config.DismissSDKInstallWarning = false
	if provider.InstallWarning() == "" {
		t.Fatal("SDKMAN install warning was not restored")
	}
}
