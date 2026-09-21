package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/platform"
)

func list(manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "list",
		ShortName: "ls",
		Usage:     "List current JDK installations.",
		Action: func(c *cli.Context) error {
			fmt.Println("Installed jdk (* marks in use):")
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
				fmt.Println(str)
			}
			if len(v) == 0 {
				fmt.Println("No installations recognized.")
			}
			return nil
		},
	}
}
