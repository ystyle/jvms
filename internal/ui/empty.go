package ui

import tea "github.com/charmbracelet/bubbletea"

func (m *pickerModel) emptyView() string {
	return "\n  No JDK versions available.\n\n" + helpStyle.Render("press any key to exit")
}

func (m *pickerModel) emptyFunc(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "q", "esc", "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}
