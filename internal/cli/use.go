package cli

import (
	"errors"
	"strings"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/platform"
)

func use(config *appcfg.Config, manager platform.Provider) *cli.Command {
	return &cli.Command{
		Name:      "use",
		ShortName: "u",
		Usage:     "Activate a JDK by version or explicit available index (for example, \"#17\"), installing it if needed.",
		Flags:     switchFlags,
		Action:    useFunc(config, manager),
	}
}

func useFunc(config *appcfg.Config, manager platform.Provider) func(*cli.Context) error {
	return func(c *cli.Context) error {
		v := strings.TrimSpace(c.Args().Get(0))
		if v == "" {
			return errors.New("a JDK version, available index, or path is required; run `jvms rls` to see available versions")
		}

		return switchVersionWithOutput(config, manager, v, c.Bool("as_path") || c.Bool("p"), true, c.App.Writer)
	}
}
