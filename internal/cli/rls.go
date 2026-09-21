package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/platform"
)

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
			}
			if err != nil {
				return fmt.Errorf("showing versions recognized from partial provider output: %w", err)
			}

			fmt.Fprintf(c.App.Writer, "\nVersions supplied by the %s provider.\n", manager.Name())
			return nil
		},
	}
}
