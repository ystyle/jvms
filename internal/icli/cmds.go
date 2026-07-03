package icli

import (
	"errors"
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
		if !admin.IsAdmin() {
			return errors.New("this command requires administrator privileges.")
		}

		if !config.JavaHomeNotSet() {
			return nil
		}

		fmt.Println("Jvms not initialized. Attempting to initialize jvms automatically...")
		if err := init_(config).Run(c); err != nil {
			return fmt.Errorf("could not initialize jvms automatically. Please run \"jvms init\" manually: %w", err)
		}

		return nil
	}
}
