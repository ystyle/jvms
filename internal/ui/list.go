package ui

import (
	bubbleView "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *pickerModel) listView() string {
	m.list.Title = m.action.Title
	if m.versionEvents != nil && !m.versionsDone {
		m.list.Title = m.spinner.View() + " " + m.list.Title
	}
	return m.list.View()
}

func (m *pickerModel) listFunc(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok || m.list.FilterState() == bubbleView.Filtering {
		return m, updateSubModel(&m.list, msg)

	}
	// Unhandled keys fall through to the list for navigation, filtering, and help.
	switch km.String() {
	case "q":
		return m, tea.Quit
	case "enter":
		if it, ok := m.list.SelectedItem().(jdkItem); ok {
			return m, m.selectItem(it.JdkVersion)
		}
		// filter matched zero items -> nothing to select, ignore
		return m, nil
	}

	return m, updateSubModel(&m.list, msg)
}
