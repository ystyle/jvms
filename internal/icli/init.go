package icli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

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
				Value: models.DefaultJavaHome,
			},
			cli.StringFlag{
				Name:  "originalpath",
				Usage: "the jdk download index file url.",
				Value: models.DefaultOriginalPath,
			},
		},
		Action: initFunc(config),
	}

}
func initPrerequisites(c *cli.Context, config *models.Config) error {
	if !admin.IsAdmin() {
		return errors.New("jvms init requires administrator privileges. Please run as administrator")
	}

	if c.IsSet("java_home") || config.JavaHome == "" {
		config.SetJavaHome(c.String("java_home"))
	}

	if c.IsSet("originalpath") || config.OriginalPath == "" {
		config.SetOriginalPath(c.String("originalpath"))
	}

	return nil
}

func initFunc(config *models.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := initPrerequisites(c, config); err != nil {
			return err
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
	}
}
