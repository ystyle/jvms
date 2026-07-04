package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// globalFunc handles messages that apply regardless of state (quit,
// spinner ticks, resize). It returns handled=true if it fully processed
// the message, so Update should stop and return immediately.
func (m *pickerModel) globalFunc(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch tm := msg.(type) {
	case tea.KeyMsg:
		if tm.String() == "ctrl+c" {
			return m, tea.Quit, true
		}

	case spinner.TickMsg:
		if !m.spinnerActive {
			return m, nil, true // stale tick from a state we've already left
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(tm)
		return m, cmd, true

	case tea.WindowSizeMsg:
		m.width, m.height = tm.Width, tm.Height
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(tm.Width-h, tm.Height-v)
		return m, nil, true
	}

	return m, nil, false
}
