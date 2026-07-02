package icli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/admin"
	"github.com/ystyle/jvms/utils/file"
)

func init_(config *models.Config) *cli.Command {
	return &cli.Command{
		Name:        "init",
		Usage:       "Initialize config file",
		Description: `before init you should clear JAVA_HOME, PATH Environment variable。`,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "java_home",
				Usage: "the JAVA_HOME location",
				Value: filepath.Join(os.Getenv("ProgramFiles"), "jdk"),
			},
			cli.StringFlag{
				Name:  "originalpath",
				Usage: "the jdk download index file url.",
				Value: DefaultOriginalpath,
			},
		},
		Action: func(c *cli.Context) error {
			if !admin.IsAdmin() {
				return errors.New("jvms init requires administrator privileges. Please run as administrator")
			}
			if c.IsSet("java_home") || config.JavaHome == "" {
				config.JavaHome = c.String("java_home")
			}
			cmd := exec.Command("cmd", "/C", "setx", "JAVA_HOME", config.JavaHome, "/M")
			err := cmd.Run()
			if err != nil {
				return fmt.Errorf("failed to set JAVA_HOME environment variable to %s: %w\n\n"+
					"Possible reasons:\n"+
					"- Insufficient permissions (try running as administrator)\n"+
					"- Command execution failed\n"+
					"- Invalid path format\n"+
					"Please run Command Prompt as Administrator and try again", config.JavaHome, err)
			}
			fmt.Println("set `JAVA_HOME` Environment variable to ", config.JavaHome)

			if c.IsSet("originalpath") || config.Originalpath == "" {
				config.Originalpath = c.String("originalpath")
			}
			path := fmt.Sprintf(`%s/bin;%s;%s`, config.JavaHome, os.Getenv("PATH"), file.GetCurrentPath())
			cmd = exec.Command("cmd", "/C", "setx", "path", path, "/m")
			err = cmd.Run()
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
		},
	}

}
