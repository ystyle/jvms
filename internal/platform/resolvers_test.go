package platform

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

type fakeManager struct {
	installed []Installation
	available []jdk.Version
	err       error
}

type partialCatalogManager struct {
	fakeManager
	catalogErr error
}

func (m partialCatalogManager) Available() ([]jdk.Version, error) {
	return m.available, m.catalogErr
}

func TestResolveAvailableVersionRejectsPartialCatalogIndexes(t *testing.T) {
	wantErr := errors.New("catalog is offline")
	manager := partialCatalogManager{
		fakeManager: fakeManager{available: []jdk.Version{{Version: "21-tem"}, {Version: "17-tem"}}},
		catalogErr:  wantErr,
	}
	for _, priority := range []appconfig.ResolutionPriority{appconfig.VersionFirst, appconfig.IndexFirst} {
		for _, value := range []string{"#2", "2"} {
			t.Run(string(priority)+"/"+value, func(t *testing.T) {
				got, err := ResolveAvailableVersion(manager, value, priority)
				if got != "" || !errors.Is(err, wantErr) {
					t.Fatalf("got %q, %v; want no version and catalog error", got, err)
				}
			})
		}
	}
}

func (m fakeManager) Name() string                         { return "fake" }
func (m fakeManager) Ensure() error                        { return m.err }
func (m fakeManager) Available() ([]jdk.Version, error)    { return m.available, m.err }
func (m fakeManager) StreamAvailable() <-chan VersionEvent { return nil }
func (m fakeManager) Installed() ([]Installation, error)   { return m.installed, m.err }
func (m fakeManager) Install(string) error                 { return m.err }
func (m fakeManager) Remove(string) error                  { return m.err }
func (m fakeManager) Switch(string) error                  { return m.err }
func (m fakeManager) SwitchPath(string) error              { return m.err }

func TestResolveVersionPropagatesProviderError(t *testing.T) {
	want := errors.New("provider unavailable")
	_, got := ResolveVersion(fakeManager{err: want}, "1", false, appconfig.IndexFirst)
	if !errors.Is(got, want) {
		t.Fatalf("ResolveVersion error = %v, want %v", got, want)
	}
}

func TestResolveVersionRejectsOutOfRangeIndex(t *testing.T) {
	manager := fakeManager{installed: []Installation{{Version: "21.0.4-tem"}}}
	if _, err := ResolveVersion(manager, "25", false, appconfig.IndexFirst); err == nil {
		t.Fatal("ResolveVersion accepted an out-of-range switch index")
	}
}

func TestResolveVersionUsesConfiguredPriority(t *testing.T) {
	manager := fakeManager{installed: []Installation{
		{Version: "17"},
		{Version: "21"},
	}}

	got, err := ResolveVersion(manager, "2", true, appconfig.VersionFirst)
	if err != nil {
		t.Fatal(err)
	}
	if got != "2" {
		t.Fatalf("version-first resolution = %q, want literal version 2", got)
	}

	got, err = ResolveVersion(manager, "2", true, appconfig.IndexFirst)
	if err != nil {
		t.Fatal(err)
	}
	if got != "21" {
		t.Fatalf("index-first resolution = %q, want installed version 21", got)
	}
}

func TestResolveVersionExplicitIndexOverridesPriority(t *testing.T) {
	manager := fakeManager{installed: []Installation{{Version: "17"}, {Version: "21"}}}
	got, err := ResolveVersion(manager, "#2", true, appconfig.VersionFirst)
	if err != nil {
		t.Fatal(err)
	}
	if got != "21" {
		t.Fatalf("explicit index resolution = %q, want 21", got)
	}
}

func TestValidateVersionIdentifier(t *testing.T) {
	for _, valid := range []string{"21.0.8-tem", "8.0.472.fx-zulu", "22.3.r17-grl"} {
		if err := validateVersionIdentifier(valid); err != nil {
			t.Errorf("validateVersionIdentifier(%q): %v", valid, err)
		}
	}
	for _, invalid := range []string{"", "../java", "21 tem", "/tmp/jdk"} {
		if err := validateVersionIdentifier(invalid); err == nil {
			t.Errorf("validateVersionIdentifier(%q) unexpectedly succeeded", invalid)
		}
	}
}

func TestValidateJavaHome(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := validateJavaHome(home, "java"); err == nil {
		t.Fatal("validateJavaHome accepted a directory without bin/java")
	}
	if err := os.WriteFile(filepath.Join(home, "bin", "java"), nil, 0755); err != nil {
		t.Fatal(err)
	}
	got, err := validateJavaHome(home, "java")
	if err != nil {
		t.Fatal(err)
	}
	if got != home {
		t.Fatalf("validateJavaHome() = %q, want %q", got, home)
	}
}

func TestAvailableVersionPriorityFallback(t *testing.T) {
	for _, test := range []struct {
		name      string
		value     string
		priority  appconfig.ResolutionPriority
		installed []Installation
		want      string
	}{
		{name: "default falls back", value: "2", want: "17-tem"},
		{name: "version first falls back", value: "2", priority: appconfig.VersionFirst, want: "17-tem"},
		{name: "exact installed wins", priority: appconfig.VersionFirst, value: "2", installed: []Installation{{Version: "2"}}, want: "2"},
		{name: "exact available wins", priority: appconfig.VersionFirst, value: "1", want: "1"},
		{name: "explicit index wins", value: "#1", want: "21-tem"},
		{name: "index first wins", value: "1", priority: appconfig.IndexFirst, want: "21-tem"},
	} {
		t.Run(test.name, func(t *testing.T) {
			manager := fakeManager{installed: test.installed, available: []jdk.Version{{Version: "21-tem"}, {Version: "17-tem"}, {Version: "1"}}}
			got, err := ResolveAvailableVersion(manager, test.value, test.priority)
			if err != nil || got != test.want {
				t.Fatalf("got %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func (m fakeManager) RefreshAvailable() ([]jdk.Version, error) { return m.available, m.err }
