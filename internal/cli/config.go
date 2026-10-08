package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
)

func configCommand(config *appcfg.Config) *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Show or update JVMS configuration.",
		Flags: []cli.Flag{
			cli.StringFlag{Name: appcfg.JavaHomeFlag, Usage: "set the JAVA_HOME location"},
			cli.StringFlag{Name: appcfg.OriginalPathFlag, Usage: "set the Windows JDK download index URL or path"},
			cli.StringFlag{Name: appcfg.ProxyFlag, Usage: "set the download proxy; pass an empty value to clear it"},
			cli.StringFlag{Name: appcfg.CacheToggleFlag, Usage: "enable or disable the JDK catalog cache (true or false)"},
			cli.StringFlag{Name: appcfg.CacheTTLFlag, Usage: "set cache lifetime, for example 30m or 24h"},
		},
		Action: configFunc(config),
	}
}

func configFunc(config *appcfg.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		next := *config
		if c.IsSet(appcfg.JavaHomeFlag) {
			next.JavaHome = strings.TrimSpace(c.String(appcfg.JavaHomeFlag))
		}
		if c.IsSet(appcfg.OriginalPathFlag) {
			next.OriginalPath = strings.TrimSpace(c.String(appcfg.OriginalPathFlag))
		}
		if c.IsSet(appcfg.ProxyFlag) {
			next.Proxy = strings.TrimSpace(c.String(appcfg.ProxyFlag))
		}
		if err := next.SetBool(c, appcfg.CacheToggleFlag, &next.CacheEnabled); err != nil {
			return err
		}
		if c.IsSet(appcfg.CacheTTLFlag) {
			value, err := appcfg.ParseCacheTTL(c.String(appcfg.CacheTTLFlag))
			if err != nil {
				return err
			}
			next.CacheTTL = value
		}

		*config = next
		printConfig(c.App.Writer, config)
		return nil
	}
}

func printConfig(w io.Writer, config *appcfg.Config) {
	fmt.Fprintln(w, "Configuration flags (set with `jvms config --flag=value`):")
	fmt.Fprintf(w, "  --%s=%s\n", appcfg.JavaHomeFlag, config.JavaHome)
	fmt.Fprintf(w, "  --%s=%s\n", appcfg.OriginalPathFlag, config.OriginalPath)
	fmt.Fprintf(w, "  --%s=%s  (empty clears it)\n", appcfg.ProxyFlag, config.Proxy)
	fmt.Fprintf(w, "  --%s=%t  (true | false)\n", appcfg.CacheToggleFlag, config.CacheEnabled)
	fmt.Fprintf(w, "  --%s=%s  (duration, e.g. 30m or 24h)\n", appcfg.CacheTTLFlag, config.CacheTTL)
}
