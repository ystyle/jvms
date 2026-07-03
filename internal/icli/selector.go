package icli

import (
	"fmt"
	"strings"

	bubbleView "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ystyle/jvms/internal/models"
)

// Action describes what happens once a version is picked. ConfirmText is
// optional — leave it nil to execute immediately on selection.
type Action struct {
	Title       string
	ConfirmText func(models.JdkVersion) string
	Execute     func(*models.Config, models.JdkVersion) error
	SuccessText func(models.JdkVersion) string
}

func (a Action) needsConfirm() bool { return a.ConfirmText != nil }

// RunJdkPicker starts the picker TUI over the given versions and blocks
func RunJdkPicker(config *models.Config, versions []models.JdkVersion, action Action) error {
	_, err := tea.NewProgram(newPickerModel(config, versions, action), tea.WithAltScreen()).Run()
	return err
}

type jdkItem struct{ models.JdkVersion }
type execDoneMsg struct{ err error }
type pickerState uint
type pickerModel struct {
	config        *models.Config
	action        Action
	state         pickerState
	list          bubbleView.Model
	spinner       spinner.Model
	spinnerActive bool

	selected models.JdkVersion
	err      error // set only on execute failure
}

const (
	stateEmpty pickerState = iota
	stateList
	stateConfirm
	stateExecuting
	stateDone
)

func (i jdkItem) Title() string       { return i.Version }
func (i jdkItem) Description() string { return i.Url }
func (i jdkItem) FilterValue() string { return i.Version }

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	dialogStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("69"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func newPickerModel(config *models.Config, versions []models.JdkVersion, action Action) pickerModel {
	l := bubbleView.New(nil, bubbleView.NewDefaultDelegate(), 0, 0)
	l.Title = action.Title
	l.SetFilteringEnabled(true)

	items := make([]bubbleView.Item, len(versions))
	for i, v := range versions {
		items[i] = jdkItem{v}
	}
	l.SetItems(items)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	state := stateList
	if len(versions) == 0 {
		state = stateEmpty
	}

	return pickerModel{
		config:  config,
		action:  action,
		state:   state,
		list:    l,
		spinner: sp,
	}
}

func (m pickerModel) Init() tea.Cmd { return nil }

// transitions
// selectItem decides whether a pick goes straight to execution or via
// confirmation, per Action.ConfirmText.
func (m *pickerModel) selectItem(v models.JdkVersion) tea.Cmd {
	m.selected = v
	if !m.action.needsConfirm() {
		return m.beginExecute()
	}
	m.state = stateConfirm
	return nil
}

func (m *pickerModel) beginExecute() tea.Cmd {
	m.state = stateExecuting
	m.err = nil
	m.spinnerActive = true
	execute := m.action.Execute
	config := m.config
	v := m.selected
	run := func() tea.Msg { return execDoneMsg{execute(config, v)} }
	return tea.Batch(run, m.spinner.Tick)
}

func (m *pickerModel) finishExecute(err error) {
	m.err = err
	m.spinnerActive = false
	m.state = stateDone
}

// update
func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if tm, ok := msg.(spinner.TickMsg); ok {
		if !m.spinnerActive {
			return m, nil // stale tick from a state we've already left
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(tm)
		return m, cmd
	}

	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(wm.Width-h, wm.Height-v)
		return m, nil
	}

	switch m.state {
	case stateEmpty:
		return m.updateEmpty(msg)
	case stateList:
		return m.updateList(msg)
	case stateConfirm:
		return m.updateConfirm(msg)
	case stateExecuting:
		return m.updateExecuting(msg)
	case stateDone:
		return m.updateDone(msg)
	}
	return m, nil
}

func (m pickerModel) updateEmpty(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "q", "esc", "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m pickerModel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m pickerModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "y", "enter":
		return m, m.beginExecute()
	case "n", "esc":
		m.state = stateList
		return m, nil
	}
	return m, nil
}

func (m pickerModel) updateExecuting(msg tea.Msg) (tea.Model, tea.Cmd) {
	dm, ok := msg.(execDoneMsg)
	if !ok {
		return m, nil // keypresses while executing are swallowed here — can't double-run
	}
	m.finishExecute(dm.err)
	return m, nil
}

func (m pickerModel) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "enter", "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

// view
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
