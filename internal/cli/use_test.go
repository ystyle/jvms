package cli

import (
	"flag"
	"testing"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
	"github.com/ystyle/jvms/internal/platform"
)

type recordingManager struct {
	installed  []platform.Installation
	available  []jdk.Version
	installedV string
	switchedV  string
	switchedP  string
}

func (m *recordingManager) Name() string                      { return "recording" }
func (m *recordingManager) Ensure() error                     { return nil }
func (m *recordingManager) Available() ([]jdk.Version, error) { return m.available, nil }
func (m *recordingManager) StreamAvailable() <-chan platform.VersionEvent {
	return nil
}
func (m *recordingManager) Installed() ([]platform.Installation, error) { return m.installed, nil }
func (m *recordingManager) Install(version string) error {
	m.installedV = version
	m.installed = append(m.installed, platform.Installation{Version: version})
	return nil
}
func (m *recordingManager) Remove(string) error { return nil }
func (m *recordingManager) Switch(version string) error {
	m.switchedV = version
	return nil
}
func (m *recordingManager) SwitchPath(path string) error {
	m.switchedP = path
	return nil
}

func commandContext(t *testing.T, args ...string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	for _, commandFlag := range switchFlags {
		commandFlag.Apply(set)
	}
	if err := set.Parse(args); err != nil {
		t.Fatal(err)
	}
	return cli.NewContext(cli.NewApp(), set, nil)
}

func TestUseInstallsAndSwitchesNumericVersion(t *testing.T) {
	manager := &recordingManager{}
	if err := useFunc(&appcfg.Config{}, manager)(commandContext(t, "25")); err != nil {
		t.Fatal(err)
	}
	if manager.installedV != "25" || manager.switchedV != "25" {
		t.Fatalf("install = %q, switch = %q; want both 25", manager.installedV, manager.switchedV)
	}
}

func TestSwitchMissingVersionReturnsError(t *testing.T) {
	manager := &recordingManager{}
	err := switchFunc(&appcfg.Config{}, manager)(commandContext(t, "missing"))
	if err == nil {
		t.Fatal("switch succeeded for a missing JDK")
	}
	if manager.switchedV != "" {
		t.Fatalf("unexpected switch to %q", manager.switchedV)
	}
}

func (m *recordingManager) RefreshAvailable() ([]jdk.Version, error) { return m.available, nil }
