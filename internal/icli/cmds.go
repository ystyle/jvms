package icli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/admin"
)

// Commands Register all commands
func Commands(c *models.Config) []cli.Command {
	return []cli.Command{
		*switch_(c), *install(c), *remove(c), *init_(c),
		*proxy(c), *list(c), *use(c), *rls(c),
	}
}

// switchPrerequisites checks if the config is initialized and if the user has admin privileges.
// If the config is not initialized and the user is an admin, it attempts to initialize itself automatically.
func switchPrerequisites(config *models.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if !config.JavaHomeNotSet() {
			return nil
		}

		if !admin.IsAdmin() {
			return fmt.Errorf("jvms is not initialized and you're not running as administrator.\n" +
				"Please run \"jvms init\" to initialize jvms or run the command as administrator to allow automatic initialization.\n" +
				"You may also specify the JAVA_HOME location using the --java_home flag when running \"jvms init\".",
			)
		}

		fmt.Println("Jvms not initialized. Attempting to initialize jvms automatically...")

		if err := init_(config).Run(c); err != nil {
			return fmt.Errorf(
				`could not initialize jvms automatically. Please run "jvms init" manually: %w`,
				err,
			)
		}

		return nil
	}
}
