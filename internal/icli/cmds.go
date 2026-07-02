package icli

import (
	"errors"
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/admin"
)

const (
	DefaultOriginalpath = "https://raw.githubusercontent.com/ystyle/jvms/new/jdkdlindex.json"
)

// ensureConfigInitializedAndIsAdmin checks if the config is initialized and if the user has admin privileges.
// If the config is not initialized and the user is an admin, it attempts to initialize itself automatically.
func ensureConfigInitializedAndIsAdmin(config *models.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if config.JavaHomeNotSet() {
			if admin.IsAdmin() {
				fmt.Println("Jvms not initialized. Attempting to initialize jvms automatically...")
				if err := init_(config).Run(&cli.Context{}); err != nil {
					return fmt.Errorf("could not initialize jvms automatically. Please run \"jvms init\" to initialize jvms\n%v", err)
				}
				return errors.New("jvms is not initialized and you're not running as administrator.\n" +
					"Please run \"jvms init\" to initialize jvms or run the command as administrator to allow automatic initialization. \n" +
					"You may also specify the JAVA_HOME location using the --java_home flag when running \"jvms init\".")
			}
		}
		return nil
	}
}

// Commands Register all commands
func Commands(c *models.Config) []cli.Command {
	return []cli.Command{
		*switch_(c), *install(c), *remove(c), *init_(c),
		*proxy(c), *list(c), *use(c), *rls(c),
	}
}
