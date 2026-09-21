package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/jdk"
)

type refreshRecordingManager struct {
	recordingManager
	refreshed bool
	err       error
}

func (m *refreshRecordingManager) RefreshAvailable() ([]jdk.Version, error) {
	m.refreshed = true
	return []jdk.Version{{Version: "21.0.4-tem"}}, m.err
}

func TestRLSRefreshesCatalogThroughProvider(t *testing.T) {
	for _, providerErr := range []error{nil, errors.New("partial catalog")} {
		manager := &refreshRecordingManager{err: providerErr}
		var output bytes.Buffer
		app := cli.NewApp()
		app.Writer = &output
		app.Commands = []cli.Command{*rls(manager)}
		err := app.Run([]string{"jvms", "rls"})
		if !errors.Is(err, providerErr) {
			t.Fatalf("error = %v, want %v", err, providerErr)
		}
		if !manager.refreshed {
			t.Fatal("rls did not request a provider refresh")
		}
		if !strings.Contains(output.String(), "1) 21.0.4-tem") {
			t.Fatalf("catalog missing from app writer: %s", output.String())
		}
	}
}
