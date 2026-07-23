package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystyle/jvms/internal/models"
)

func (m *pickerModel) executingView() string {
	return fmt.Sprintf("\n  %s %s: %s...\n", m.spinner.View(), m.action.Title, m.selected.Version)
}

func (m *pickerModel) selectItem(v models.JdkVersion) tea.Cmd {
	m.selected = v
	if !m.action.needsConfirm() {
		return m.beginExecute()
	}
	m.state = stateConfirm
	return nil
}

func (m *pickerModel) beginExecute() tea.Cmd {
	m.state = stateExecuting
	m.err = nil
	m.spinnerActive = true
	execute := m.action.Execute
	config := m.config
	v := m.selected
	run := func() tea.Msg { return execDoneMsg{execute(config, v)} }
	return tea.Batch(run, m.spinner.Tick)
}

func (m *pickerModel) finishExecute(err error) {
	m.err = err
	m.spinnerActive = false
	m.state = stateDone
}

func (m *pickerModel) executeFunc(msg tea.Msg) (tea.Model, tea.Cmd) {
	dm, ok := msg.(execDoneMsg)
	if !ok {
		return m, nil // keypresses while executing are swallowed here -- can't double-run
	}
	m.finishExecute(dm.err)
	return m, nil
}
