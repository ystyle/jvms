package ui

import (
	bubbleView "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m pickerModel) listView() string {
	return m.list.View()
}

func (m pickerModel) listFunc(msg tea.Msg) (tea.Model, tea.Cmd) {
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
