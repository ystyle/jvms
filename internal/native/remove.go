//go:build windows

package native

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ystyle/jvms/internal/jdk"
)

func (m *service) Remove(v string) error {
	config := m.config
	if v == "" {
		return errors.New("you should input a version, Type \"jvms list\" to see what is installed")
	}

	if !jdk.IsVersionInstalled(config.Store, v) {
		return fmt.Errorf("JDK %s is not installed", v)
	}

	if config.CurrentJDKVersion == v {
		if err := os.Remove(config.JavaHome); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove JAVA_HOME symlink: %w", err)
		}
	}

	dir := filepath.Join(config.Store, v)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove JDK %s at %s: %w", v, dir, err)
	}

	if config.CurrentJDKVersion == v {
		config.CurrentJDKVersion = ""
	}

	return nil
}
