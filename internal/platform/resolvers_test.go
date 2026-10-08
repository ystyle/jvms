package platform

import (
	"errors"
	"testing"

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

func TestVersionAndExplicitIndexResolution(t *testing.T) {
	manager := fakeManager{
		installed: []Installation{{Version: "21-tem"}, {Version: "jdk 17"}, {Version: "1"}},
		available: []jdk.Version{{Version: "21-tem"}, {Version: "jdk 17"}, {Version: "1"}},
	}
	for name, resolve := range map[string]func(Provider, string) (string, error){
		"installed": ResolveVersion,
		"available": ResolveAvailableVersion,
	} {
		t.Run(name, func(t *testing.T) {
			for _, test := range []struct {
				value   string
				want    string
				wantErr bool
			}{
				{value: "1", want: "1"},
				{value: "17", want: "17"},
				{value: "jdk 17", want: "jdk 17"},
				{value: " #1 ", want: "21-tem"},
				{value: "#2", want: "jdk 17"},
				{value: "#3", want: "1"},
				{value: "", wantErr: true},
				{value: " ", wantErr: true},
				{value: "#", wantErr: true},
				{value: "#0", wantErr: true},
				{value: "#-1", wantErr: true},
				{value: "#4", wantErr: true},
				{value: "#invalid", wantErr: true},
				{value: "#999999999999999999999999", wantErr: true},
			} {
				t.Run(test.value, func(t *testing.T) {
					got, err := resolve(manager, test.value)
					if (err != nil) != test.wantErr || got != test.want {
						t.Fatalf("got %q, %v; want %q, error=%t", got, err, test.want, test.wantErr)
					}
				})
			}
		})
	}
}

func TestBareVersionsDoNotQueryProvider(t *testing.T) {
	manager := fakeManager{err: errors.New("provider unavailable")}
	for _, resolve := range []func(Provider, string) (string, error){ResolveVersion, ResolveAvailableVersion} {
		for _, value := range []string{"1", "17", "jdk 17"} {
			got, err := resolve(manager, value)
			if err != nil || got != value {
				t.Fatalf("got %q, %v; want %q without a provider lookup", got, err, value)
			}
		}
	}
}

func TestExplicitIndexPropagatesProviderError(t *testing.T) {
	want := errors.New("provider unavailable")
	for _, resolve := range []func(Provider, string) (string, error){ResolveVersion, ResolveAvailableVersion} {
		_, got := resolve(fakeManager{err: want}, "#1")
		if !errors.Is(got, want) {
			t.Fatalf("error = %v, want %v", got, want)
		}
	}
}

func (m fakeManager) RefreshAvailable() ([]jdk.Version, error) { return m.available, m.err }
