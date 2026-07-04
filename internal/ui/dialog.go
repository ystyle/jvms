package ui

import (
	"strings"
)

const (
	successPrefix     = "✓"
	fallbackHelpStyle = "press enter to exit"
)

func (m *pickerModel) newDialog(text, help string) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.action.Title))
	b.WriteString("\n\n")
	b.WriteString(text)
	b.WriteString("\n\n")

	if help == "" {
		b.WriteString(helpStyle.Render(fallbackHelpStyle))
	} else {
		b.WriteString(helpStyle.Render(help))
	}

	return dialogStyle.Render(b.String())
}
