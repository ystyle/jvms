package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type subModelUpdate[T any] interface {
	Update(tea.Msg) (T, tea.Cmd)
}

// updateSubModel drives any subModel-shaped component, writes the result
// back into *component, and hands back the tea.Cmd to propagate.
func updateSubModel[T subModelUpdate[T]](component *T, msg tea.Msg) tea.Cmd {
	newComp, cmd := (*component).Update(msg)
	*component = newComp
	return cmd
}

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
		return m, updateSubModel(&m.spinner, tm), true

	case tea.WindowSizeMsg:
		m.width, m.height = tm.Width, tm.Height
		m.list.SetSize(tm.Width, tm.Height)
		return m, nil, true
	}

	return m, nil, false
}
