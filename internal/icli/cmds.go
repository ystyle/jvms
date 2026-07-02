package icli

import (
	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
)

const (
	DefaultOriginalpath = "https://raw.githubusercontent.com/ystyle/jvms/new/jdkdlindex.json"
)

// Commands Registrer
func Commands(c *models.Config) []cli.Command {
	return []cli.Command{
		*switch_(c),
		*install(c),
		*remove(c),
		*init_(c),
		*proxy(c),
		*list(c),
		*use(c),
		*rls(c),
	}
}
