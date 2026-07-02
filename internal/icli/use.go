package icli

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/entity"
	"github.com/ystyle/jvms/utils/jdk"
)

func use(config *entity.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "use",
		ShortName: "u",
		Usage:     "Switch to use the specified version or index number and install it if not installed.",
		Flags:     switchFlags,
		Action:    intercept(config),
	}
	return cmd
}

func intercept(config *entity.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		v := strings.TrimSpace(c.Args().Get(0))
		isInstalled := jdk.IsVersionInstalled(config.Store, v)
		if v != "" {
			// If not installed, redirect to install
			if !isInstalled {
				fmt.Printf("Version %s is not installed. Installing now...\n", v)
				err := installFunc(config)(c)
				if err != nil {
					return err
				}
			}
		}

		return switchFunc(config)(c)
	}
}
