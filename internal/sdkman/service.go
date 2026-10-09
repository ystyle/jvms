//go:build !windows

package sdkman

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	appconfig "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/httpclient"
	"github.com/ystyle/jvms/internal/jdk"
)

// Service manages SDKMAN Java installations and the available version catalog.
type Service interface {
	Ensure() error
	Health() error
	Available() ([]jdk.Version, error)
	RefreshAvailable() ([]jdk.Version, error)
	StreamAvailable() <-chan jdk.VersionEvent
	Installed() ([]jdk.Installation, error)
	Install(version string) error
	Remove(version string) error
	Switch(version string) error
	SwitchPath(path string) error
}

type service struct {
	config   *appconfig.Config
	dir      string
	download func(string) ([]byte, error)
}

var _ Service = (*service)(nil)

func NewService(config *appconfig.Config) Service {
	dir := strings.TrimSpace(os.Getenv("SDKMAN_DIR"))
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".sdkman")
	}

	if config == nil {
		config = appconfig.NewConfig()
	}

	manager := &service{
		config: config,
		dir:    dir,
	}

	manager.download = func(url string) ([]byte, error) {
		httpclient.SetProxy(config.Proxy)
		return httpclient.GetBytes(url, 2*time.Minute)
	}

	return manager
}

func (m *service) initScript() string {
	return filepath.Join(m.dir, "bin", "sdkman-init.sh")
}

func (m *service) candidatesDir() string {
	return filepath.Join(m.dir, "candidates", "java")
}

func (m *service) Health() error {
	output, err := m.run(false, "version")
	if err != nil {
		return fmt.Errorf("SDKMAN health check failed: %w", err)
	}

	if !strings.Contains(output, "SDKMAN!") {
		return errors.New("SDKMAN health check returned an unexpected response")
	}

	return nil
}

func (m *service) Install(version string) error {
	if err := requireSDKMANUser(); err != nil {
		return err
	}

	if err := jdk.ValidateVersionIdentifier(version); err != nil {
		return err
	}

	_, err := m.run(true, "install", "java", version)
	return err
}

func (m *service) Remove(version string) error {
	if err := requireSDKMANUser(); err != nil {
		return err
	}

	if err := jdk.ValidateVersionIdentifier(version); err != nil {
		return err
	}

	_, err := m.run(true, "uninstall", "java", version)
	return err
}

func (m *service) Switch(version string) error {
	if err := requireSDKMANUser(); err != nil {
		return err
	}

	if err := jdk.ValidateVersionIdentifier(version); err != nil {
		return err
	}

	_, err := m.run(true, "default", "java", version)
	return err
}

func (m *service) SwitchPath(path string) error {
	if err := requireSDKMANUser(); err != nil {
		return err
	}

	if err := m.requireInstalled(); err != nil {
		return err
	}

	home, err := jdk.ValidateJavaHome(path, "java")
	if err != nil {
		return err
	}

	current, err := filepath.Abs(filepath.Join(m.candidatesDir(), "current"))
	if err != nil {
		return fmt.Errorf("resolve SDKMAN current JDK path: %w", err)
	}

	if home == current {
		return nil
	}

	// Resolve aliases before replacing current so indirect links cannot become recursive.
	home, err = filepath.EvalSymlinks(home)
	if err != nil {
		return fmt.Errorf("resolve JDK home symlinks: %w", err)
	}

	if err := os.MkdirAll(m.candidatesDir(), 0755); err != nil {
		return fmt.Errorf("create SDKMAN Java candidates directory: %w", err)
	}

	// SDKMAN exposes the active JDK through this link; custom paths have no candidate identifier.
	if info, err := os.Lstat(current); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("refusing to replace non-symlink SDKMAN current JDK at %s", current)
		}

		if err := os.Remove(current); err != nil {
			return fmt.Errorf("remove current SDKMAN JDK link: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect current SDKMAN JDK link: %w", err)
	}

	if err := os.Symlink(home, current); err != nil {
		return fmt.Errorf("activate JDK path: %w", err)
	}

	return nil
}

func (m *service) requireInstalled() error {
	if _, err := os.Stat(m.initScript()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("SDKMAN is not installed; run `jvms init`")
		}

		return fmt.Errorf("inspect SDKMAN init script: %w", err)
	}

	return nil
}
