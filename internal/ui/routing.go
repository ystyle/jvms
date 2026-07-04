package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if tm, ok := msg.(spinner.TickMsg); ok {
		if !m.spinnerActive {
			return m, nil // stale tick from a state we've already left
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(tm)
		return m, cmd
	}

	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(wm.Width-h, wm.Height-v)
		return m, nil
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
