package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/jdk"
)

type versionEventMsg struct{ event jdk.VersionEvent }

func RunStreamingJdkPicker(config *models.Config, action Action) error {
	model := newPickerModel(config, nil, action)
	model.state = stateList
	model.versionEvents = jdk.StreamJdkVersions(config)
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

func (m *pickerModel) appendVersion(v models.JdkVersion) tea.Cmd {
	if m.state == stateEmpty && !m.versionsDone {
		m.state = stateList
	}
	items := append(m.list.Items(), jdkItem{v})
	return m.list.SetItems(items)
}
