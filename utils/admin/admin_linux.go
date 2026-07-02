//go:build !windows

package admin

// IsAdmin checks if the current process has administrative privileges
func IsAdmin() bool {
	// This returns stub for non-windows platforms. It always returns false otherwise.
	return false // To support development on linux without importing "golang.org/x/sys/windows" on non-windows platforms
}
