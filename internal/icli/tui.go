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
	cli.BoolFlag{Name: "s", Usage: "Switch to use the specified version or index number."},
}

func tui(config *models.Config) *cli.Command {
	return &cli.Command{
		Name:      "tui",
		ShortName: "tui",
		Usage:     "Run the interactive TUI interface.",
		Flags:     tuiFlags,
		Action: func(c *cli.Context) error {
			switch {
			case c.Bool("s"):
				versions := []models.JdkVersion{}
				for _, v := range jdk.GetInstalled(config.Store) {
					versions = append(versions, models.JdkVersion{Version: v})
				}
				return RunJdkPicker(config, versions, Action{
					Title: "Switch JDK",
					ConfirmText: func(v models.JdkVersion) string {
						return fmt.Sprintf("Switch to %s?", v.Version)
					},
					Execute: func(config *models.Config, v models.JdkVersion) error {
						return switchFunc(config)(withArg(c, v.Version))
					},
					SuccessText: func(v models.JdkVersion) string {
						return fmt.Sprintf("Now using JDK %s", v.Version)
					},
				})

			case c.Bool("u"):
				fmt.Println("Getting available versions...")
				versions, err := jdk.GetJdkVersions(config)
				if err != nil {
					return err
				}
				return RunJdkPicker(config, versions, Action{
					Title: "Switch or Install JDK",
					ConfirmText: func(v models.JdkVersion) string {
						return fmt.Sprintf("Switch to %s?", v.Version)
					},
					Execute: func(config *models.Config, v models.JdkVersion) error {
						return useFunc(config)(withArg(c, v.Version))
					},
					SuccessText: func(v models.JdkVersion) string {
						return fmt.Sprintf("Now using JDK %s", v.Version)
					},
				})

			case c.Bool("i"):
				fmt.Println("Getting available versions...")
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
				return errors.New("specify -u, -s to switch or -i to install")
			}
		},
	}
}
