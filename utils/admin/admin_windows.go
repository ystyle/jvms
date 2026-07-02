//go:build windows

package admin

import (
	"golang.org/x/sys/windows"
)

// IsAdmin checks if the current process has administrative privileges
func IsAdmin() bool {
	var sid *windows.SID

	// Although this looks huge, it's the equivalent of
	// set "sid=S-1-5-32-544"
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

	// This appears to cast a null pointer, but that is actually fine.
	// CheckTokenMembership allows the token handle to be 0 (current process token).
	// It's checked against the sid created above.
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}
