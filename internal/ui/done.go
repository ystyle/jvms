package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m pickerModel) doneView() string {
	if m.err != nil {
		m.newDialog(errStyle.Render(fmt.Sprintf("\n  %s failed: %v\n\n",
			m.action.Title, m.err)), "")
	}

	return m.newDialog(successStyle.Render(fmt.Sprintf("\n  %s %s\n\n",
		successPrefix, m.action.SuccessText(m.selected))), "")
}

func (m pickerModel) doneFunc(msg tea.Msg) (tea.Model, tea.Cmd) {
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
