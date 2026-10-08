//go:build windows

package native

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

type Installation = jdk.Installation

// Service manages native JDK installations and the Windows Java environment.
type Service interface {
	Ensure() error
	Installed() ([]Installation, error)
	Install(version string) error
	Remove(version string) error
	Switch(version string) error
	SwitchPath(path string) error
}

type service struct {
	config  *appcfg.Config
	isAdmin func() bool
	run     func(string, ...string) error
}

var _ Service = (*service)(nil)

// NewService retains the shared config so changes to the active JDK are visible to callers.
func NewService(config *appcfg.Config) Service {
	return &service{config: config, isAdmin: isAdmin,
		run: func(name string, args ...string) error { return exec.Command(name, args...).Run() }}
}

func (m *service) Installed() ([]Installation, error) {
	var installed []Installation

	for _, v := range jdk.GetInstalled(m.config.Store) {
		if !jdk.IsVersionInstalled(m.config.Store, v) {
			continue
		}

		installed = append(installed, Installation{Version: v, Current: v == m.config.CurrentJDKVersion})
	}

	return installed, nil
}

func (m *service) replaceJavaHomeTarget(target string) error {
	config := m.config
	// Create or update the symlink
	if info, err := os.Lstat(config.JavaHome); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("refusing to replace non-symlink JAVA_HOME at %s", config.JavaHome)
		}

		err := os.Remove(config.JavaHome)
		if err != nil {
			return fmt.Errorf("failed to remove existing JavaHome symlink at %s: %w\n\nPossible reasons:\n"+
				"- Insufficient permissions (try running as administrator)\n"+
				"- File is in use by another process\n"+
				"- Path points to a directory instead of a symlink\n"+
				"Please manually remove it and try again", config.JavaHome, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	err := m.run("cmd", "/C", "setx", "JAVA_HOME", config.JavaHome, "/M")
	if err != nil {
		return errors.New("set Environment variable `JAVA_HOME` failure: Please run as admin user")
	}

	err = os.Symlink(target, config.JavaHome)
	if err != nil {
		return errors.New("Switch jdk failed, " + err.Error())
	}

	return nil
}
