package ui

import (
	bubbleView "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// update
func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m.updateEmpty(msg)
	case stateList:
		return m.updateList(msg)
	case stateConfirm:
		return m.updateConfirm(msg)
	case stateExecuting:
		return m.updateExecuting(msg)
	case stateDone:
		return m.updateDone(msg)
	}
	return m, nil
}

func (m pickerModel) updateEmpty(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "q", "esc", "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pickerModel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok || m.list.FilterState() == bubbleView.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
	switch km.String() {
	case "q":
		return m, tea.Quit
	case "enter":
		if it, ok := m.list.SelectedItem().(jdkItem); ok {
			cmd := m.selectItem(it.JdkVersion)
			return m, cmd
		}
		// filter matched zero items -> nothing to select, ignore
		return m, nil
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m pickerModel) updateExecuting(msg tea.Msg) (tea.Model, tea.Cmd) {
	dm, ok := msg.(execDoneMsg)
	if !ok {
		return m, nil // keypresses while executing are swallowed here -- can't double-run
	}
	m.finishExecute(dm.err)
	return m, nil
}

func (m pickerModel) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
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
