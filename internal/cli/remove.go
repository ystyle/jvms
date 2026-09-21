package cli

import (
	"errors"
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/platform"
)

func remove(manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "remove",
		ShortName: "rm",
		Usage:     "Remove a specific version.",
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			if v == "" {
				return errors.New("you should input a version, Type \"jvms list\" to see what is installed")
			}
			installed, err := platform.IsInstalled(manager, v)
			if err != nil {
				return err
			}
			if installed {
				fmt.Printf("Remove JDK %s ...\n", v)
				if err := manager.Remove(v); err != nil {
					return err
				}
				fmt.Println("done")
			} else {
				fmt.Println("jdk " + v + " is not installed. Type \"jvms list\" to see what is installed.")
			}
			return nil
		},
	}
}
