package icli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/jdk"
)

func use(config *models.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "use",
		ShortName: "u",
		Usage:     "Switch to use the specified version or index number and install it if not installed.",
		Flags:     switchFlags,
		Action:    useFunc(config),
	}
	return cmd
}

func useFunc(config *models.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		v := strings.TrimSpace(c.Args().Get(0))
		if v == "" {
			return errors.New("you should input a version or index number, Type \"jvms list\" to see what is installed")
		}
		isInstalled := jdk.IsVersionInstalled(config.Store, v)
		if !isInstalled {
			fmt.Printf("Version %s is not installed. Installing now...\n", v)
			err := installFunc(config)(c)
			if err != nil {
				return err
			}
		}

		return switchFunc(config)(c)
	}
}
