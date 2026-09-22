package platform

import (
	"errors"
	"testing"

	appconfig "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

type fakeManager struct {
	installed []Installation
	available []jdk.Version
	err       error
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
