//go:build !windows

package admin

// IsAdmin checks if the current process has administrative privileges
// checks are only implemented on Windows to allow development on Linux.
func IsAdmin() bool { // Used to develop ui for now, will implement later
	return false // To support development on linux without importing "golang.org/x/sys/windows" on non-windows platforms
}
