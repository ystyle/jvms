package cli

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/platform"
)

func CommandNotFound(c *cli.Context, command string) {
	log.Fatal("Command Not Found")
}

// Commands registers all CLI commands.
func Commands(c *appcfg.Config, m platform.Provider) []cli.Command {
	return []cli.Command{
		*tui(c, m), *switch_(c, m), *install(m), *remove(m),
		*init_(c, m), *proxy(c), *configCommand(c), *list(m), *use(c, m), *rls(m),
	}
}

func proxy(config *appcfg.Config) *cli.Command {
	return &cli.Command{
		Name:  "proxy",
		Usage: "Set a proxy to use for downloads.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "show",
				Usage: "show proxy.",
			},
			cli.StringFlag{
				Name:  "set",
				Usage: "set proxy.",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("show") {
				fmt.Fprintf(c.App.Writer, "Current proxy: %s\n", config.Proxy)
				return nil
			}

			if c.IsSet("set") {
				config.Proxy = c.String("set")
			}

			return nil
		},
	}
}

func list(manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "list",
		ShortName: "ls",
		Usage:     "List current JDK installations.",
		Action: func(c *cli.Context) error {
			fmt.Fprintln(c.App.Writer, "Installed jdk (* marks in use):")
			v, err := manager.Installed()
			if err != nil {
				return err
			}

			for i, version := range v {
				str := ""
				if version.Current {
					str = fmt.Sprintf("%s  * %d) %s", str, i+1, version.Version)
				} else {
					str = fmt.Sprintf("%s    %d) %s", str, i+1, version.Version)
				}

				fmt.Fprintln(c.App.Writer, str)
			}

			if len(v) == 0 {
				fmt.Fprintln(c.App.Writer, "No installations recognized.")
			} else {
				fmt.Fprintln(c.App.Writer, "\nUse `jvms switch \"#<index>\"` to select a numbered row.")
			}

			return nil
		},
	}
}

func rls(manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:  "rls",
		Usage: "Show versions available for installation.",
		Flags: []cli.Flag{cli.BoolFlag{Name: "a", Usage: "list all versions"}},
		Action: func(c *cli.Context) error {
			versions, err := manager.RefreshAvailable()
			if err != nil && len(versions) == 0 {
				return err
			}

			for i, version := range versions {
				fmt.Fprintf(c.App.Writer, "    %d) %s\n", i+1, version.Version)
				if !c.Bool("a") && i >= 9 {
					fmt.Fprintln(c.App.Writer, "\nUse `jvms rls -a` to show all versions.")
					break
				}
			}

			if len(versions) == 0 {
				fmt.Fprintln(c.App.Writer, "No JDK versions are available for installation.")
			} else {
				fmt.Fprintln(c.App.Writer, "\nUse `jvms install \"#<index>\"` or `jvms use \"#<index>\"` to select a numbered row.")
			}
			if err != nil {
				return fmt.Errorf("showing versions recognized from partial provider output: %w", err)
			}

			fmt.Fprintf(c.App.Writer, "\nVersions supplied by the %s provider.\n", manager.Name())
			return nil
		},
	}
}

func remove(manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "remove",
		ShortName: "rm",
		Usage:     "Remove a specific version.",
		Action: func(c *cli.Context) error {
			v := strings.TrimSpace(c.Args().Get(0))
			if v == "" {
				return errors.New("a JDK version is required; run `jvms list` to see installed versions")
			}

			installed, err := platform.IsInstalled(manager, v)
			if err != nil {
				return err
			}

			if installed {
				fmt.Fprintf(c.App.Writer, "Remove JDK %s ...\n", v)
				if err := manager.Remove(v); err != nil {
					return err
				}

				fmt.Fprintln(c.App.Writer, "done")
			} else {
				fmt.Fprintln(c.App.Writer, "JDK "+v+" is not installed; run `jvms list` to see installed versions.")
			}

			return nil
		},
	}
}
