package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

type versionEventMsg struct{ event jdk.VersionEvent }

func RunStreamingJdkPicker(config *appcfg.Config, events <-chan jdk.VersionEvent, action Action) error {
	model := newPickerModel(config, nil, action)
	model.state = stateList
	model.versionEvents = events
	model.spinnerActive = true

	finalModel, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		return err
	}
	if pm, ok := finalModel.(*pickerModel); ok {
		if pm.err != nil {
			return pm.err
		}
		if pm.loadErr != nil && len(pm.list.Items()) == 0 {
			return pm.loadErr
		}
	}
	return nil
}

func waitForVersionEvent(events <-chan jdk.VersionEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return versionEventMsg{event: jdk.VersionEvent{Done: true}}
		}
		return versionEventMsg{event: event}
	}
}

func (m *pickerModel) appendVersion(v jdk.JdkVersion) tea.Cmd {
	if m.state == stateEmpty && !m.versionsDone {
		m.state = stateList
	}
	items := append(m.list.Items(), jdkItem{v})
	return m.list.SetItems(items)
}
