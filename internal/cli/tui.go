package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
	"github.com/ystyle/jvms/internal/platform"
	apptui "github.com/ystyle/jvms/internal/tui"
)

var tuiFlags = []cli.Flag{
	cli.BoolFlag{Name: "u", Usage: "Launch the switch picker"},
	cli.BoolFlag{Name: "i", Usage: "Launch the install picker"},
	cli.BoolFlag{Name: "s", Usage: "Switch to use the specified version or index number."},
}

func tui(config *appcfg.Config, manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "tui",
		ShortName: "tui",
		Usage:     "Run the interactive TUI interface.",
		Flags:     tuiFlags,
		Action:    tuiFunc(config, manager),
	}
}

func tuiFuncUse(config *appcfg.Config, manager platform.Provider) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		return apptui.RunStreamingJdkPicker(config, manager.StreamAvailable(), apptui.Action{
			Title: "Use JDK/Install JDK",
			ConfirmText: func(v jdk.JdkVersion) string {
				return fmt.Sprintf("Use %s?", v.Version)
			},
			Execute: func(config *appcfg.Config, v jdk.JdkVersion) error {
				return switchExactVersion(config, manager, v.Version, true)
			},
			SuccessText: func(v jdk.JdkVersion) string {
				return fmt.Sprintf("Now using JDK %s", v.Version)
			},
		})
	}
}

func tuiFunc(config *appcfg.Config, manager platform.Provider) func(*cli.Context) error {
	return func(c *cli.Context) error {
		switch {
		case c.Bool("u"):
			return tuiFuncUse(config, manager)(c)

		case c.Bool("s"):
			versions := []jdk.JdkVersion{}
			installed, err := manager.Installed()
			if err != nil {
				return err
			}
			for _, v := range installed {
				versions = append(versions, jdk.JdkVersion{Version: v.Version})
			}
			return apptui.RunJdkPicker(config, versions, apptui.Action{
				Title: "Switch JDK",
				ConfirmText: func(v jdk.JdkVersion) string {
					return fmt.Sprintf("Switch to %s?", v.Version)
				},
				Execute: func(config *appcfg.Config, v jdk.JdkVersion) error {
					return switchExactVersion(config, manager, v.Version, false)
				},
				SuccessText: func(v jdk.JdkVersion) string {
					return fmt.Sprintf("Now using JDK %s", v.Version)
				},
			})

		case c.Bool("i"):
			return apptui.RunStreamingJdkPicker(config, manager.StreamAvailable(), apptui.Action{
				Title: "Install JDK",
				ConfirmText: func(v jdk.JdkVersion) string {
					return fmt.Sprintf("Install %s?", v.Version)
				},
				Execute: func(config *appcfg.Config, v jdk.JdkVersion) error {
					return installVersion(manager, v.Version)
				},
				SuccessText: func(v jdk.JdkVersion) string {
					return fmt.Sprintf("Installed %s", v.Version)
				},
			})

		default:
			return tuiFuncUse(config, manager)(c)
		}
	}
}
