package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
)

func proxy(config *appcfg.Config) *cli.Command {
	return &cli.Command{
		Name:  "proxy",
		Usage: "Set a proxy to use for downloads.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "show",
				Usage: "show proxy.",
			},
			cli.StringFlag{
				Name:  "set",
				Usage: "set proxy.",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("show") {
				fmt.Fprintf(c.App.Writer, "Current proxy: %s\n", config.Proxy)
				return nil
			}
			if c.IsSet("set") {
				config.Proxy = c.String("set")
			}
			return nil
		},
	}
}
