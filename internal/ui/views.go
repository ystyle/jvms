package ui

import (
	"fmt"
	"strings"
)

// View renders the current state of the picker model to a string for display in the terminal.
func (m pickerModel) View() string {
	switch m.state {
	case stateEmpty:
		return m.viewEmpty()
	case stateList:
		return m.viewList()
	case stateConfirm:
		return m.viewConfirm()
	case stateExecuting:
		return m.viewExecuting()
	case stateDone:
		return m.viewDone()
	}
	return ""
}

func (m pickerModel) viewEmpty() string {
	return "\n  No JDK versions available.\n\n" + helpStyle.Render("press any key to exit")
}

func (m pickerModel) viewList() string {
	return m.list.View()
}

func (m pickerModel) viewConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.action.Title))
	b.WriteString("\n\n")
	b.WriteString(m.action.ConfirmText(m.selected))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("y/enter: confirm • n/esc: cancel"))
	return dialogStyle.Render(b.String())
}

func (m pickerModel) viewExecuting() string {
	return fmt.Sprintf("\n  %s %s: %s...\n", m.spinner.View(), m.action.Title, m.selected.Version)
}

func (m pickerModel) viewDone() string {
	if m.err != nil {
		return errStyle.Render(fmt.Sprintf("\n  %s failed: %v\n\n", m.action.Title, m.err)) +
			helpStyle.Render("press enter to exit")
	}
	text := fmt.Sprintf("%s.", m.selected.Version)
	if m.action.SuccessText != nil {
		text = m.action.SuccessText(m.selected)
	}
	return successStyle.Render(fmt.Sprintf("\n  ✓ %s\n\n", text)) + helpStyle.Render("press enter to exit")
}
