//go:build !windows

package platform

import (
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
	"github.com/ystyle/jvms/internal/sdkman"
)

type sdkmanManager struct {
	service sdkman.Service
	config  *appcfg.Config
}

var _ Provider = (*sdkmanManager)(nil)

func NewProvider(config *appcfg.Config) Provider {
	return &sdkmanManager{
		service: sdkman.NewService(config),
		config:  config,
	}
}

func (m *sdkmanManager) Name() string {
	return "SDKMAN!"
}

func (m *sdkmanManager) InstallWarning() string {
	if m.config != nil && m.config.DismissSDKInstallWarning {
		return ""
	}

	return "Warning: SDKMAN installation may update both your current Java version and the default."
}

func (m *sdkmanManager) Available() ([]jdk.Version, error) {
	return m.service.Available()
}

func (m *sdkmanManager) StreamAvailable() <-chan VersionEvent {
	return m.service.StreamAvailable()
}

func (m *sdkmanManager) RefreshAvailable() ([]jdk.Version, error) {
	return m.service.RefreshAvailable()
}

func (m *sdkmanManager) Ensure() error {
	return m.service.Ensure()
}

func (m *sdkmanManager) Installed() ([]Installation, error) {
	return m.service.Installed()
}

func (m *sdkmanManager) Install(version string) error {
	return m.service.Install(version)
}

func (m *sdkmanManager) Remove(version string) error {
	return m.service.Remove(version)
}

func (m *sdkmanManager) Switch(version string) error {
	return m.service.Switch(version)
}

func (m *sdkmanManager) SwitchPath(path string) error {
	return m.service.SwitchPath(path)
}
