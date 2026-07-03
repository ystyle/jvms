package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m pickerModel) doneView() string {
	if m.err != nil {
		return errStyle.Render(fmt.Sprintf("\n  %s failed: %v\n\n", m.action.Title, m.err)) +
			helpStyle.Render("press enter to exit")
	}
	text := fmt.Sprintf("%s.", m.selected.Version)
	if m.action.SuccessText != nil {
		text = m.action.SuccessText(m.selected)
	}
	return successStyle.Render(fmt.Sprintf("\n  ✓ %s\n\n", text)) + helpStyle.Render("press enter to exit")
}

func (m pickerModel) handleDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "enter", "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}
