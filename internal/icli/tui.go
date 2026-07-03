// internal/icli/tui.go
package icli

import (
	"errors"
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/jdk"
)

var tuiFlags = []cli.Flag{
	cli.BoolFlag{Name: "u", Usage: "Launch the switch picker"},
	cli.BoolFlag{Name: "i", Usage: "Launch the install picker"},
}

func tui(config *models.Config) *cli.Command {
	return &cli.Command{
		Name:      "tui",
		ShortName: "tui",
		Usage:     "Run the interactive TUI interface.",
		Flags:     tuiFlags,
		Action: func(c *cli.Context) error {
			switch {
			case c.Bool("u"):
				versions, err := jdk.GetJdkVersions(config)
				if err != nil {
					return err
				}
				return RunJdkPicker(config, versions, Action{
					Title: "Switch JDK",
					ConfirmText: func(v models.JdkVersion) string {
						return fmt.Sprintf("Switch to %s?", v.Version)
					},
					// Delegate to the existing, untouched CLI path: build a
					// synthetic context whose one positional arg is the
					// version the user picked, exactly as if they'd typed
					// `jvms switch <version>`.
					Execute: func(config *models.Config, v models.JdkVersion) error {
						return switchFunc(config)(withArg(c, v.Version))
					},
					SuccessText: func(v models.JdkVersion) string {
						return fmt.Sprintf("Now using JDK %s", v.Version)
					},
				})

			case c.Bool("i"):
				versions, err := jdk.GetJdkVersions(config)
				if err != nil {
					return err
				}
				return RunJdkPicker(config, versions, Action{
					Title: "Install JDK",
					ConfirmText: func(v models.JdkVersion) string {
						return fmt.Sprintf("Install %s?", v.Version)
					},
					Execute: func(config *models.Config, v models.JdkVersion) error {
						return installFunc(config)(withArg(c, v.Version))
					},
					SuccessText: func(v models.JdkVersion) string {
						return fmt.Sprintf("Installed %s", v.Version)
					},
				})

			default:
				return errors.New("specify -u to switch or -i to install")
			}
		},
	}
}
