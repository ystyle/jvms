package icli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/jdk"
)

func list(config *models.Config) *cli.Command {
	return &cli.Command{
		Name:      "list",
		ShortName: "ls",
		Usage:     "List current JDK installations.",
		Action: func(c *cli.Context) error {
			fmt.Println("Installed jdk (* marks in use):")
			v := jdk.GetInstalled(config.Store)
			for i, version := range v {
				str := ""
				if config.CurrentJDKVersion == version {
					str = fmt.Sprintf("%s  * %d) %s", str, i+1, version)
				} else {
					str = fmt.Sprintf("%s    %d) %s", str, i+1, version)
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
