package cli

import (
	"io"
	"testing"

	"github.com/ystyle/jvms/internal/platform"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

func TestInstallResolvesIndexFromAvailableCatalog(t *testing.T) {
	manager := &recordingManager{available: []jdk.Version{
		{Version: "26.0.2-amzn"},
		{Version: "25.0.4-amzn"},
	}}
	config := &appcfg.Config{ResolutionPriority: appcfg.IndexFirst}
	if err := installFunc(config, manager)(commandContext(t, "2")); err != nil {
		t.Fatal(err)
	}
	if manager.installedV != "25.0.4-amzn" {
		t.Fatalf("installed %q, want catalog index 2", manager.installedV)
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
