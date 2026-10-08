//go:build windows

package native

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ystyle/jvms/internal/jdk"
)

func (m *service) SwitchPath(path string) error {
	if !m.isAdmin() {
		return errors.New("this command requires administrator privileges.")
	}

	home, err := jdk.ValidateJavaHome(path, "javac.exe")
	if err != nil {
		return err
	}

	if err := m.replaceJavaHomeTarget(home); err != nil {
		return err
	}

	m.config.CurrentJDKVersion = ""
	return nil
}

func (m *service) Switch(v string) error {
	config := m.config
	if !m.isAdmin() {
		return errors.New("this command requires administrator privileges.")
	}

	if config.JavaHomeNotSet() {
		if err := m.Ensure(); err != nil {
			return err
		}
	}

	if !jdk.IsVersionInstalled(config.Store, v) {
		return fmt.Errorf("JDK %s is not installed", v)
	}

	if err := m.replaceJavaHomeTarget(filepath.Join(config.Store, v)); err != nil {
		return err
	}

	config.CurrentJDKVersion = v
	return nil
}
