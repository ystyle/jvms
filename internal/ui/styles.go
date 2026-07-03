package ui

import "github.com/charmbracelet/lipgloss"

// Palette -- grayscale for structure/emphasis, two muted semantic colors
// for outcomes. Nothing draws attention except success/error states.
const (
	colorAccent  = lipgloss.Color("253") // bright grey — titles, borders, focus
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
