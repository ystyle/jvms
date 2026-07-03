package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m pickerModel) confirmView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.action.Title))
	b.WriteString("\n\n")
	b.WriteString(m.action.ConfirmText(m.selected))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("y/enter: confirm • n/esc: cancel"))
	return dialogStyle.Render(b.String())
}

func (m pickerModel) handleConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "y", "enter":
		return m, m.beginExecute()
	case "n", "esc":
		m.state = stateList
		return m, nil
	}
	return m, nil
}
