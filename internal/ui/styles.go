package ui

import "github.com/charmbracelet/lipgloss"

// Palette -- one accent, two muted semantic colors, a gray ramp.
const (
	colorAccent  = lipgloss.Color("111") // soft blue -- focus, titles, borders
	colorSuccess = lipgloss.Color("108") // muted sage green
	colorError   = lipgloss.Color("203") // soft coral red
	colorMuted   = lipgloss.Color("245") // help text, secondary info
	colorDim     = lipgloss.Color("238") // faint borders, disabled state
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	dialogStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).Padding(1, 2)
	errStyle     = lipgloss.NewStyle().Foreground(colorError)
	successStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	helpStyle    = lipgloss.NewStyle().Foreground(colorMuted)
)
