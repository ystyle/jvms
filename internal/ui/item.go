package ui

import "github.com/ystyle/jvms/internal/models"

type jdkItem struct{ models.JdkVersion }

func (i jdkItem) Title() string { return i.Version }
func (i jdkItem) Description() string {
	if i.Url != "" {
		return i.Url
	}
	return "installed"
}
func (i jdkItem) FilterValue() string { return i.Version }
