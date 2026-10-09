package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/platform"
)

func install(manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "install",
		ShortName: "i",
		Usage:     "Install a JDK by version or explicit available index (for example, \"#17\").",
		Action:    installFunc(manager),
	}
}

func installFunc(manager platform.Provider) func(*cli.Context) error {
	return func(c *cli.Context) error {
		v := strings.TrimSpace(c.Args().Get(0))
		if v == "" {
			return errors.New("a JDK version or available index is required; run `jvms rls` to see available versions")
		}

		version, err := platform.ResolveAvailableVersion(manager, v)
		if err != nil {
			return err
		}

		if version != v {
			fmt.Fprintf(c.App.Writer, "Using available index %s to select JDK %s\n", strings.TrimPrefix(v, "#"), version)
		}

		return installVersionWithOutput(manager, version, c.App.Writer)
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
