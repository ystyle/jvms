package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	successPrefix     = "✓"
	fallbackHelpStyle = "press enter to exit"
)

// dialog.go — wrap the rendered dialog in Place using the model's known size
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

	box := dialogStyle.Render(b.String())

	if m.width == 0 || m.height == 0 {
		return box // no size known yet (e.g. first frame) — fall back ungracefully
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
