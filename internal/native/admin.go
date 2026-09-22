//go:build windows

package native

import (
	"golang.org/x/sys/windows"
)

// isAdmin checks if the current process has administrative privileges
func isAdmin() bool {
	var sid *windows.SID

	// Build the well-known Administrators SID (S-1-5-32-544).
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	// A zero token asks IsMember to check the current process token.
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}
