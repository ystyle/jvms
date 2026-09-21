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

func install(config *appcfg.Config, manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "install",
		ShortName: "i",
		Usage:     "Install a JDK by version or available-catalog index.",
		Action:    installFunc(config, manager),
	}
}

func installFunc(config *appcfg.Config, manager platform.Provider) func(*cli.Context) error {
	return func(c *cli.Context) error {
		v := strings.TrimSpace(c.Args().Get(0))
		if v == "" {
			return errors.New("a JDK version or available index is required; run `jvms rls` to see available versions")
		}
		v, err := platform.ResolveAvailableVersion(manager, v, config.ResolutionPriority)
		if err != nil {
			return err
		}

		return installVersion(manager, v)
	}
}

// installVersion installs an exact identifier already selected or resolved.
func installVersion(manager platform.Provider, v string) error {
	return installVersionWithOutput(manager, v, os.Stdout)
}

func installVersionWithOutput(manager platform.Provider, v string, output io.Writer) error {
	isInstalled, err := platform.IsInstalled(manager, v)
	if err != nil {
		return err
	}
	if isInstalled {
		fmt.Fprintln(output, "Version "+v+" is already installed.")
		return nil
	}
	fmt.Fprintf(output, "Installing JDK %s with %s ...\n", v, manager.Name())
	if err := manager.Install(v); err != nil {
		return err
	}
	fmt.Fprintf(output, "Installed JDK %s. Use `jvms switch %s` to activate it.\n", v, v)
	return nil
}
