package ui

import "github.com/ystyle/jvms/internal/models"

// Action describes what happens once a version is picked. ConfirmText is
// optional -- leave it nil to execute immediately on selection.
type Action struct {
	Title       string
	ConfirmText func(models.JdkVersion) string
	Execute     func(*models.Config, models.JdkVersion) error
	SuccessText func(models.JdkVersion) string
}

func (a Action) needsConfirm() bool { return a.ConfirmText != nil }
