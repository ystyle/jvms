package tui

import "github.com/ystyle/jvms/internal/jdk"

type jdkItem struct{ jdk.JdkVersion }

func (i jdkItem) Title() string { return i.Version }

func (i jdkItem) Description() string {
	if i.Url != "" {
		return i.Url
	}

	return "installed"
}

func (i jdkItem) FilterValue() string { return i.Version }
