package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/platform"
)

func init_(config *appcfg.Config, provider platform.Provider) *cli.Command {
	return &cli.Command{
		Name:        "init",
		Usage:       "Initialize and verify the platform JDK provider",
		Description: "Set up the platform-specific provider used to manage JDK installations.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  appcfg.JavaHomeFlag,
				Usage: "the JAVA_HOME location",
				Value: appcfg.DefaultJavaHome,
			},
			cli.StringFlag{
				Name:  appcfg.OriginalPathFlag,
				Usage: "the jdk download index file url.",
				Value: appcfg.DefaultOriginalPath,
			},
		},
		Action: initFunc(config, provider),
	}

}

func initPrerequisites(c *cli.Context, config *appcfg.Config) error {
	if c.IsSet(appcfg.JavaHomeFlag) || config.JavaHome == "" {
		config.JavaHome = c.String(appcfg.JavaHomeFlag)
	}

	if c.IsSet(appcfg.OriginalPathFlag) || config.OriginalPath == "" {
		config.OriginalPath = c.String(appcfg.OriginalPathFlag)
	}

	return nil
}

func initFunc(config *appcfg.Config, provider platform.Provider) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := initPrerequisites(c, config); err != nil {
			return err
		}

		if err := provider.Ensure(); err != nil {
			return err
		}

		fmt.Fprintf(c.App.Writer, "%s provider is installed and working.\n", provider.Name())
		return nil
	}
}
