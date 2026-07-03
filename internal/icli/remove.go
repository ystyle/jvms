package icli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/jdk"
)

func remove(config *models.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "remove",
		ShortName: "rm",
		Usage:     "Remove a specific version.",
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			if v == "" {
				return errors.New("you should input a version, Type \"jvms list\" to see what is installed")
			}
			if jdk.IsVersionInstalled(config.Store, v) {
				fmt.Printf("Remove JDK %s ...\n", v)
				if config.CurrentJDKVersion == v {
					os.Remove(config.JavaHome)
				}
				dir := filepath.Join(config.Store, v)
				e := os.RemoveAll(dir)
				if e != nil {
					fmt.Println("Error removing jdk " + v)
					fmt.Println("Manually remove " + dir + ".")
				} else {
					fmt.Printf(" done")
				}
			} else {
				fmt.Println("jdk " + v + " is not installed. Type \"jvms list\" to see what is installed.")
			}
			return nil
		},
	}
	return cmd
}
