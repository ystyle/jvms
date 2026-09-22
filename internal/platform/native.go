//go:build windows

package platform

import (
	"fmt"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
	"github.com/ystyle/jvms/internal/native"
)

type nativeManager struct {
	config  *appcfg.Config
	service native.Service
}

var _ Provider = (*nativeManager)(nil)

func NewProvider(config *appcfg.Config) Provider {
	return &nativeManager{config: config, service: native.NewService(config)}
}

func (m *nativeManager) Name() string                         { return "JVMS native Windows" }
func (m *nativeManager) Available() ([]jdk.Version, error)    { return jdk.GetJdkVersions(m.config, true) }
func (m *nativeManager) StreamAvailable() <-chan VersionEvent { return jdk.StreamJdkVersions(m.config) }

// RefreshAvailable discards the native catalog cache before fetching versions.
func (m *nativeManager) RefreshAvailable() ([]jdk.Version, error) {
	if err := jdk.InvalidateCache(); err != nil {
		return nil, fmt.Errorf("refresh JDK catalog: %w", err)
	}
	return m.Available()
}

func (m *nativeManager) Ensure() error                      { return m.service.Ensure() }
func (m *nativeManager) Installed() ([]Installation, error) { return m.service.Installed() }
func (m *nativeManager) Install(version string) error       { return m.service.Install(version) }
func (m *nativeManager) Remove(version string) error        { return m.service.Remove(version) }
func (m *nativeManager) Switch(version string) error        { return m.service.Switch(version) }
func (m *nativeManager) SwitchPath(path string) error       { return m.service.SwitchPath(path) }
