package ui

import (
	bubbleView "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/jdk"
)

type execDoneMsg struct{ err error }
type pickerState uint
type pickerModel struct {
	config        *models.Config
	action        Action
	state         pickerState
	list          bubbleView.Model
	spinner       spinner.Model
	spinnerActive bool
	versionEvents <-chan jdk.VersionEvent
	versionsDone  bool
	loadErr       error

	width, height int // terminal size
	selected      models.JdkVersion
	err           error // set only on execute failure
}

const (
	stateEmpty pickerState = iota
	stateList
	stateConfirm
	stateExecuting
	stateDone
)

// RunJdkPicker starts the picker TUI over the given versions and blocks.
func RunJdkPicker(config *models.Config, versions []models.JdkVersion, action Action) error {
	finalModel, err := tea.NewProgram(newPickerModel(config, versions, action), tea.WithAltScreen()).Run()
	if err != nil {
		return err
	}
	if pm, ok := finalModel.(*pickerModel); ok {
		return pm.err
	}
	return nil
}

func newPickerModel(config *models.Config, versions []models.JdkVersion, action Action) *pickerModel {
	if action.Execute == nil || action.SuccessText == nil {
		panic("ui: Action.Execute and Action.SuccessText are required")
	}

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

func (m *pickerModel) Init() tea.Cmd {
	if m.versionEvents != nil {
		if m.spinnerActive {
			return tea.Batch(waitForVersionEvent(m.versionEvents), m.spinner.Tick)
		}
		return waitForVersionEvent(m.versionEvents)
	}
	return nil
}
