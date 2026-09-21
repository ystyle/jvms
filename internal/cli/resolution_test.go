package cli

import (
	"testing"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
	"github.com/ystyle/jvms/internal/platform"
)

func TestInstallAndUseFallBackToAvailableIndex(t *testing.T) {
	for _, use := range []bool{false, true} {
		manager := &recordingManager{available: []jdk.Version{{Version: "21-tem"}, {Version: "17-tem"}}}
		config := &appcfg.Config{}
		action := installFunc(config, manager)
		if use {
			action = useFunc(config, manager)
		}
		if err := action(commandContext(t, "2")); err != nil {
			t.Fatal(err)
		}
		if manager.installedV != "17-tem" {
			t.Fatalf("use=%t: installed %q, want 17-tem", use, manager.installedV)
		}
		if use && manager.switchedV != "17-tem" {
			t.Fatalf("switched %q, want 17-tem", manager.switchedV)
		}
	}
}

func TestExactPickerOperationsPreserveNumericIdentifier(t *testing.T) {
	for _, operation := range []string{"install", "use", "switch"} {
		t.Run(operation, func(t *testing.T) {
			manager := &recordingManager{available: []jdk.Version{{Version: "21-tem"}, {Version: "17-tem"}, {Version: "2"}}}
			config := &appcfg.Config{ResolutionPriority: appcfg.IndexFirst}
			var err error
			switch operation {
			case "install":
				err = installVersion(manager, "2")
			case "use":
				err = switchExactVersion(config, manager, "2", true)
			case "switch":
				manager.installed = []platform.Installation{{Version: "2"}, {Version: "17-tem"}}
				err = switchExactVersion(config, manager, "2", false)
			}
			if err != nil {
				t.Fatal(err)
			}
			if operation != "switch" && manager.installedV != "2" {
				t.Fatalf("installed %q, want 2", manager.installedV)
			}
			if operation != "install" && (manager.switchedV != "2" || config.CurrentJDKVersion != "2") {
				t.Fatalf("switched %q, current %q; want 2", manager.switchedV, config.CurrentJDKVersion)
			}
		})
	}
}
