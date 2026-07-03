package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m pickerModel) executingView() string {
	return fmt.Sprintf("\n  %s %s: %s...\n", m.spinner.View(), m.action.Title, m.selected.Version)
}

func (m pickerModel) handleExecuting(msg tea.Msg) (tea.Model, tea.Cmd) {
	dm, ok := msg.(execDoneMsg)
	if !ok {
		return m, nil // keypresses while executing are swallowed here -- can't double-run
	}
	m.finishExecute(dm.err)
	return m, nil
}
