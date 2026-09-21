//go:build !windows

package privilege

import "os"

// IsAdmin reports whether the process is running with root privileges.
func IsAdmin() bool {
	return os.Geteuid() == 0
}
