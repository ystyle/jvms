package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/platform"
)

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
