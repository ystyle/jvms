package icli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/file"
	"github.com/ystyle/jvms/utils/jdk"
)

var asPathUsage = "Interpret the argument as a direct path rather than a version or index number."

// Shared flags between switch and use commands
var switchFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "as_path",
		Usage: asPathUsage,
	},
	cli.BoolFlag{
		Name:  "p",
		Usage: asPathUsage,
	},
}

func switch_(config *models.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "switch",
		ShortName: "s",
		Usage:     "Switch to use the specified version or index number.",
		Flags:     switchFlags,
		Action:    switchFunc(config),
	}
	return cmd
}

// switchPrerequisites checks if the prerequisites for switching JDK are met
// SwitchFunc is used by both switch and use commands
func switchFunc(config *models.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		// Ensure the config is initialized with java_home & isAdmin
		if err := switchPrerequisites(config)(c); err != nil {
			return err
		}
		v := strings.TrimSpace(c.Args().Get(0))
		if v == "" {
			return errors.New("you should input a version or index number, Type \"jvms list\" to see what is installed")
		}

		// Try to resolve the version from context, which may be a version string,
		// an index number, or a direct path
		v, err := jdk.ResolveJdkVersion(c, config, v)
		if err != nil {
			return err
		}

		if !jdk.IsVersionInstalled(config.Store, v) {
			fmt.Printf("jdk %s is not installed. ", v)
			return nil
		}
		// Create or update the symlink
		if file.Exists(config.JavaHome) {
			err := os.Remove(config.JavaHome)
			if err != nil {
				return fmt.Errorf("failed to remove existing JavaHome symlink at %s: %w\n\nPossible reasons:\n"+
					"- Insufficient permissions (try running as administrator)\n"+
					"- File is in use by another process\n"+
					"- Path points to a directory instead of a symlink\n"+
					"Please manually remove it and try again", config.JavaHome, err)
			}
		}
		cmd := exec.Command("cmd", "/C", "setx", "JAVA_HOME", config.JavaHome, "/M")
		err = cmd.Run()
		if err != nil {
			return errors.New("set Environment variable `JAVA_HOME` failure: Please run as admin user")
		}
		err = os.Symlink(filepath.Join(config.Store, v), config.JavaHome)
		if err != nil {
			return errors.New("Switch jdk failed, " + err.Error())
		}
		fmt.Println("\nSwitch success.\nNow using JDK " + v)
		config.CurrentJDKVersion = v
		return nil
	}
}
