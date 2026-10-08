package cli

import (
	"testing"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
	"github.com/ystyle/jvms/internal/platform"
)

func TestInstallAndUseNumericResolution(t *testing.T) {
	for _, operation := range []string{"install", "use"} {
		for _, test := range []struct {
			name  string
			value string
			want  string
		}{
			{name: "default version", value: "17", want: "17"},
			{name: "explicit index", value: "#17", want: "zulu24.32.13-ca-jdk24.0.2"},
		} {
			t.Run(operation+"/"+test.name, func(t *testing.T) {
				available := make([]jdk.Version, 73)
				available[16].Version = "zulu24.32.13-ca-jdk24.0.2"
				manager := &recordingManager{available: available}
				config := appcfg.NewConfig()
				action := installFunc(manager)
				if operation == "use" {
					action = useFunc(config, manager)
				}
				if err := action(commandContext(t, test.value)); err != nil {
					t.Fatal(err)
				}
				if manager.installedV != test.want {
					t.Fatalf("installed %q, want %q", manager.installedV, test.want)
				}
				if operation == "use" && manager.switchedV != test.want {
					t.Fatalf("switched %q, want %q", manager.switchedV, test.want)
				}
			})
		}
	}
}

func TestUseAlreadyInstalledNumericVersionWithLargeCatalog(t *testing.T) {
	available := make([]jdk.Version, 73)
	available[16].Version = "zulu24.32.13-ca-jdk24.0.2"
	manager := &recordingManager{
		available: available,
		installed: []platform.Installation{{Version: "17"}},
	}
	if err := useFunc(appcfg.NewConfig(), manager)(commandContext(t, "17")); err != nil {
		t.Fatal(err)
	}
	if manager.installedV != "" || manager.switchedV != "17" {
		t.Fatalf("installed %q, switched %q; want no install and switch to 17", manager.installedV, manager.switchedV)
	}
}

func TestExactPickerOperationsPreserveNumericIdentifier(t *testing.T) {
	for _, operation := range []string{"install", "use", "switch"} {
		t.Run(operation, func(t *testing.T) {
			manager := &recordingManager{available: []jdk.Version{{Version: "21-tem"}, {Version: "17-tem"}, {Version: "2"}}}
			config := appcfg.NewConfig()
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

func TestSwitchBareNumberIsAlwaysAVersion(t *testing.T) {
	manager := &recordingManager{installed: []platform.Installation{{Version: "jdk 17"}}}
	config := appcfg.NewConfig()
	if err := switchFunc(config, manager)(commandContext(t, "1")); err == nil {
		t.Fatal("bare 1 selected an installed index")
	}
	if manager.switchedV != "" {
		t.Fatalf("unexpected switch to %q", manager.switchedV)
	}

	manager.installed = append(manager.installed, platform.Installation{Version: "1"})
	if err := switchFunc(config, manager)(commandContext(t, "1")); err != nil {
		t.Fatal(err)
	}
	if manager.switchedV != "1" || config.CurrentJDKVersion != "1" {
		t.Fatalf("switched %q, current %q; want version 1", manager.switchedV, config.CurrentJDKVersion)
	}
}
