//go:build !windows

package platform

import (
	"fmt"
	"runtime"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

type unsupportedProvider struct{}

var _ Provider = (*unsupportedProvider)(nil)

func NewProvider(_ *appcfg.Config) Provider { return &unsupportedProvider{} }

func unsupportedPlatformError() error {
	return fmt.Errorf("JVMS is not yet supported on %s yet; the native provider requires Windows", runtime.GOOS)
}

func (*unsupportedProvider) Name() string  { return "Unsupported platform: " + runtime.GOOS }
func (*unsupportedProvider) Ensure() error { return unsupportedPlatformError() }
func (*unsupportedProvider) Available() ([]jdk.Version, error) {
	return nil, unsupportedPlatformError()
}
func (*unsupportedProvider) StreamAvailable() <-chan VersionEvent {
	events := make(chan VersionEvent, 1)
	events <- VersionEvent{Err: unsupportedPlatformError(), Done: true}
	close(events)
	return events
}
func (*unsupportedProvider) Installed() ([]Installation, error) {
	return nil, unsupportedPlatformError()
}
func (*unsupportedProvider) Install(string) error    { return unsupportedPlatformError() }
func (*unsupportedProvider) Remove(string) error     { return unsupportedPlatformError() }
func (*unsupportedProvider) Switch(string) error     { return unsupportedPlatformError() }
func (*unsupportedProvider) SwitchPath(string) error { return unsupportedPlatformError() }
