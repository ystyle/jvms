package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

func TestStreamErrorDisplaysFailureWithoutAddingJDK(t *testing.T) {
	for _, done := range []bool{false, true} {
		name := "error without done"
		if done {
			name = "terminal error"
		}

		t.Run(name, func(t *testing.T) {
			wantErr := errors.New("provider requires Windows")
			model := newPickerModel(appcfg.NewConfig(), nil, Action{
				Execute: func(*appcfg.Config, jdk.JdkVersion) error {
					t.Fatal("error event must not execute an action")
					return nil
				},
				SuccessText: func(jdk.JdkVersion) string { return "done" },
			})
			model.state = stateList
			model.spinnerActive = true
			model.Update(versionEventMsg{event: jdk.VersionEvent{Err: wantErr, Done: done}})
			// A later channel-close notification must not erase the failure.
			model.Update(versionEventMsg{event: jdk.VersionEvent{Done: true}})
			if len(model.list.Items()) != 0 {
				t.Fatal("error event added a selectable JDK")
			}

			if !errors.Is(model.loadErr, wantErr) || !strings.Contains(model.View(), wantErr.Error()) {
				t.Fatalf("provider error not displayed: %s", model.View())
			}

			if !model.versionsDone || model.spinnerActive {
				t.Fatal("picker still loading after stream failure")
			}

			_, quit := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if quit == nil {
				t.Fatal("error screen did not allow exit")
			}

			if _, ok := quit().(tea.QuitMsg); !ok {
				t.Fatal("expected quit command from error screen")
			}
		})
	}
}
