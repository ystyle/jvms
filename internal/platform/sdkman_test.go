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
