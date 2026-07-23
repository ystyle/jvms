package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystyle/jvms/utils/jdk"
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
// spinner ticks, resize).
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

	case versionEventMsg:
		return m, m.handleVersionEvent(tm.event), true

	case tea.WindowSizeMsg:
		m.width, m.height = tm.Width, tm.Height
		m.list.SetSize(tm.Width, tm.Height)
		return m, nil, true
	}

	return m, nil, false
}

func (m *pickerModel) handleVersionEvent(event jdk.VersionEvent) tea.Cmd {
	if event.Done {
		m.versionsDone = true
		if m.state != stateExecuting {
			m.spinnerActive = false
		}
		m.loadErr = event.Err
		if len(m.list.Items()) == 0 {
			m.state = stateEmpty
		}
		return nil
	}

	return tea.Batch(m.appendVersion(event.Version), waitForVersionEvent(m.versionEvents))
}
