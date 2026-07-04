package ui

import (
	bubbleView "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystyle/jvms/internal/models"
)

type jdkItem struct{ models.JdkVersion }

func (i jdkItem) Title() string       { return i.Version }
func (i jdkItem) Description() string {
	if i.Url != "" {
		return i.Url
	}
	return "installed"
}
func (i jdkItem) FilterValue() string { return i.Version }

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

// RunJdkPicker starts the picker TUI over the given versions and blocks
func RunJdkPicker(config *models.Config, versions []models.JdkVersion, action Action) error {
	_, err := tea.NewProgram(newPickerModel(config, versions, action), tea.WithAltScreen()).Run()
	return err
}

func newPickerModel(config *models.Config, versions []models.JdkVersion, action Action) *pickerModel {
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

	return &pickerModel{
		config:  config,
		action:  action,
		state:   state,
		list:    l,
		spinner: sp,
	}
}

func (m *pickerModel) Init() tea.Cmd { return nil }

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
