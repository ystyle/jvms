package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/ystyle/jvms/internal/platform"

	appcfg "github.com/ystyle/jvms/internal/config"
)

type installWarningManager struct {
	recordingManager
}

func (m *installWarningManager) InstallWarning() string {
	return "SDKMAN may change your default/current Java version."
}

func TestInstallPrintsWarningWithoutConfirmation(t *testing.T) {
	manager := &installWarningManager{}
	var output bytes.Buffer
	if err := installVersionWithOutput(manager, "21.0.8-tem", &output); err != nil {
		t.Fatal(err)
	}

	if manager.installedV != "21.0.8-tem" || manager.switchedV != "" {
		t.Fatalf("install should delegate directly without switching: %+v", manager)
	}

	if !strings.Contains(output.String(), manager.InstallWarning()) || strings.Contains(output.String(), "Continue?") {
		t.Fatalf("expected an informational warning without a confirmation: %s", output.String())
	}
}

func TestAlreadyInstalledQuietActions(t *testing.T) {
	manager := &recordingManager{installed: []platform.Installation{{Version: "25"}}}
	config := &appcfg.Config{}
	if err := installVersionWithOutput(manager, "25", io.Discard); err != nil {
		t.Fatal(err)
	}

	if manager.installedV != "" {
		t.Fatal("reinstalled existing version")
	}

	if err := switchExactVersionWithOutput(config, manager, "25", true, io.Discard); err != nil {
		t.Fatal(err)
	}

	if manager.installedV != "" || manager.switchedV != "25" || config.CurrentJDKVersion != "25" {
		t.Fatalf("existing version was not activated correctly: %+v", manager)
	}
}
