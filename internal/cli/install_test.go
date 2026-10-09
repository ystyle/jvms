package cli

import (
	"io"
	"testing"

	"github.com/ystyle/jvms/internal/platform"

	appcfg "github.com/ystyle/jvms/internal/config"
)

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
