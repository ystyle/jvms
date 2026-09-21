package tui

import (
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

// Action describes what happens once a version is picked. ConfirmText is
// optional -- leave it nil to execute immediately on selection.
type Action struct {
	Title       string
	ConfirmText func(jdk.JdkVersion) string
	Execute     func(*appcfg.Config, jdk.JdkVersion) error
	SuccessText func(jdk.JdkVersion) string
}

func (a Action) needsConfirm() bool { return a.ConfirmText != nil }
