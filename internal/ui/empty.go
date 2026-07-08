package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *pickerModel) emptyView() string {
	if m.loadErr != nil {
		return "\n  Failed to load JDK versions: " + m.loadErr.Error() + "\n\n" + helpStyle.Render("press any key to exit")
	}
	return "\n  No JDK versions available.\n\n" + helpStyle.Render("press any key to exit")
}

func (m *pickerModel) emptyFunc(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}
