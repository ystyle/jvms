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
	if jdk.IsVersionInstalled(config.Store, v) {
		fmt.Printf("Remove JDK %s ...\n", v)
		if config.CurrentJDKVersion == v {
			if err := os.Remove(config.JavaHome); err != nil {
				fmt.Printf("Warning: failed to remove JavaHome symlink: %v\n", err)
			}
		}
		dir := filepath.Join(config.Store, v)
		e := os.RemoveAll(dir)
		if e != nil {
			fmt.Println("Error removing jdk " + v)
			fmt.Println("Manually remove " + dir + ".")
			return e
		} else {
			if config.CurrentJDKVersion == v {
				config.CurrentJDKVersion = ""
			}
			fmt.Printf(" done")
		}
	} else {
		return fmt.Errorf("JDK %s is not installed", v)
	}
	return nil
}
