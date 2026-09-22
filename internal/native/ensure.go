//go:build windows

package native

import (
	"errors"
	"fmt"
	"os"

	"github.com/ystyle/jvms/internal/fsutil"
)

func (m *service) Ensure() error {
	config := m.config
	if !m.isAdmin() {
		return errors.New("jvms init requires administrator privileges. Please run as administrator")
	}

	err := m.run("cmd", "/C", "setx", "JAVA_HOME", config.JavaHome, "/M")
	if err != nil {
		return fmt.Errorf("failed to set JAVA_HOME environment variable to %s: %w\n\n"+
			"Possible reasons:\n"+
			"- Insufficient permissions (try running as administrator)\n"+
			"- Command execution failed\n"+
			"- Invalid path format\n"+
			"Please run Command Prompt as Administrator and try again", config.JavaHome, err)
	}
	fmt.Println("set `JAVA_HOME` Environment variable to ", config.JavaHome)
	path := fmt.Sprintf(`%s/bin;%s;%s`, config.JavaHome, os.Getenv("PATH"), fsutil.GetCurrentPath())
	err = m.run("cmd", "/C", "setx", "path", path, "/m")
	if err != nil {
		return fmt.Errorf("failed to add jvms.exe to PATH environment variable: %w\n\n"+
			"Possible reasons:\n"+
			"- Insufficient permissions (try running as administrator)\n"+
			"- PATH variable is too long (Windows has a 2048 character limit)\n"+
			"- Command execution failed\n"+
			"Please run Command Prompt as Administrator and try again", err)
	}
	fmt.Println("add jvms.exe to `path` Environment variable")
	return nil
}
