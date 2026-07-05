package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// View renders the current state of the picker model to a string for display in the terminal.
func (m *pickerModel) View() string {
	switch m.state {
	case stateEmpty:
		return m.emptyView()
	case stateList:
		return m.listView()
	case stateConfirm:
		return m.confirmView()
	case stateExecuting:
		return m.executingView()
	case stateDone:
		return m.doneView()
	}
	return ""
}

func (m *pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := m.globalFunc(msg); handled {
		return model, cmd
	}

	switch m.state {
	case stateEmpty:
		return m.emptyFunc(msg)
	case stateList:
		return m.listFunc(msg)
	case stateConfirm:
		return m.confirmFunc(msg)
	case stateExecuting:
		return m.executeFunc(msg)
	case stateDone:
		return m.doneFunc(msg)
	}
	return m, nil
}
