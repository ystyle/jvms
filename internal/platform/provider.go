package platform

import "github.com/ystyle/jvms/internal/jdk"

// Provider defines platform-independent JDK management for the CLI and TUI.
type Provider interface {
	Name() string
	Ensure() error
	Available() ([]jdk.Version, error)
	RefreshAvailable() ([]jdk.Version, error)
	StreamAvailable() <-chan VersionEvent
	Installed() ([]Installation, error)
	Install(version string) error
	Remove(version string) error
	Switch(version string) error
	SwitchPath(path string) error
}

// VersionEvent carries a discovered JDK or the final completion status.
type VersionEvent = jdk.VersionEvent

type Installation = jdk.Installation
