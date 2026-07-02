package icli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/entity"
)

func proxy(config *entity.Config) *cli.Command {
	cmd := &cli.Command{
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
				fmt.Printf("Current proxy: %s\n", config.Proxy)
				return nil
			}
			if c.IsSet("set") {
				config.Proxy = c.String("set")
			}
			return nil
		},
	}
	return cmd
}
