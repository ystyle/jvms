package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m pickerModel) confirmView() string {
	return m.newDialog(m.action.ConfirmText(m.selected),
		helpStyle.Render("y/enter: confirm • n/esc: cancel"))
}

func (m pickerModel) confirmFunc(msg tea.Msg) (tea.Model, tea.Cmd) {
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
