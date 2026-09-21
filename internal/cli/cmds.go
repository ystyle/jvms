package cli

import (
	"log"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/platform"
)

func CommandNotFound(c *cli.Context, command string) {
	log.Fatal("Command Not Found")
}

// Commands registers all CLI commands.
func Commands(c *appcfg.Config, m platform.Provider) []cli.Command {
	return []cli.Command{
		*tui(c, m), *switch_(c, m), *install(c, m), *remove(m),
		*init_(c, m), *proxy(c), *configCommand(c), *list(m), *use(c, m), *rls(m),
	}
}
