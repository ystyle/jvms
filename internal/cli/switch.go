package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/platform"
)

var asPathUsage = "Interpret the argument as a direct path rather than a version or explicit index (#N)."

// Shared flags for the switch and use commands.
var switchFlags = []cli.Flag{
	cli.BoolFlag{Name: "as_path, p", Usage: asPathUsage},
}

func switch_(config *appcfg.Config, manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "switch",
		ShortName: "s",
		Usage:     "Activate an installed JDK by version or explicit installed index (for example, \"#1\").",
		Flags:     switchFlags,
		Action:    switchFunc(config, manager),
	}
}

func switchFunc(config *appcfg.Config, manager platform.Provider) func(*cli.Context) error {
	return func(c *cli.Context) error {
		v := strings.TrimSpace(c.Args().Get(0))
		if v == "" {
			return errors.New("a JDK version, index, or path is required; run `jvms list` to see installed versions")
		}
		return switchVersionWithOutput(config, manager, v, c.Bool("as_path") || c.Bool("p"), false, c.App.Writer)
	}
}

func switchVersionWithOutput(config *appcfg.Config, manager platform.Provider, value string, asPath, installMissing bool, output io.Writer) error {
	if asPath {
		if err := manager.SwitchPath(value); err != nil {
			return err
		}
		config.CurrentJDKVersion = ""
		fmt.Fprintln(output, "Switch success.\nNow using JDK at "+value)
		return nil
	}

	var version string
	var err error
	if installMissing {
		version, err = platform.ResolveAvailableVersion(manager, value)
	} else {
		version, err = platform.ResolveVersion(manager, value)
	}
	if err != nil {
		return err
	}
	return switchExactVersionWithOutput(config, manager, version, installMissing, output)
}

// switchExactVersion activates an exact identifier without interpreting it as an index.
func switchExactVersion(config *appcfg.Config, manager platform.Provider, version string, installMissing bool) error {
	return switchExactVersionWithOutput(config, manager, version, installMissing, os.Stdout)
}

func switchExactVersionWithOutput(config *appcfg.Config, manager platform.Provider, version string, installMissing bool, output io.Writer) error {
	installed, err := platform.IsInstalled(manager, version)
	if err != nil {
		return err
	}
	if !installed {
		if !installMissing {
			return fmt.Errorf("JDK %s is not installed", version)
		}
		fmt.Fprintf(output, "Version %s is not installed. Installing now...\n", version)
		fmt.Fprintf(output, "Installing JDK %s with %s ...\n", version, manager.Name())
		if err := manager.Install(version); err != nil {
			return err
		}
		fmt.Fprintf(output, "Installed JDK %s. Use `jvms switch %s` to activate it.\n", version, version)
	}
	if err := manager.Switch(version); err != nil {
		return err
	}
	config.CurrentJDKVersion = version
	fmt.Fprintln(output, "Switch success.\nNow using JDK "+version)
	return nil
}
