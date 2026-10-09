//go:build !windows

package sdkman

import (
	"errors"
	"os"
)

func requireSDKMANUser() error {
	if os.Geteuid() == 0 {
		return errors.New("SDKMAN operations must run as a regular user; do not use sudo or root")
	}

	return nil
}
