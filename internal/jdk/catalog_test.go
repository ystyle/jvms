package jdk

import (
	"errors"
	"reflect"
	"testing"

	appcfg "github.com/ystyle/jvms/internal/config"
)

type catalogTestProvider struct {
	versions []Version
	err      error
}

func (p catalogTestProvider) Name() string { return "test" }
func (p catalogTestProvider) Fetch(out chan<- Version) error {
	for _, version := range p.versions {
		out <- version
	}
	return p.err
}

func TestCatalogIndexesIgnoreResponseOrder(t *testing.T) {
	want := []Version{
		{Version: "21", Url: "https://a.example/jdk"},
		{Version: "21", Url: "https://b.example/jdk"},
		{Version: "17", Url: "https://example.com/jdk"},
	}
	orders := [][]Version{
		{want[2], want[1], want[0]},
		{want[1], want[0], want[2]},
	}
	for _, providerErr := range []error{nil, errors.New("partial catalog")} {
		for _, order := range orders {
			got, errs, err := fetchJdkVersions([]jdkProvider{catalogTestProvider{versions: order, err: providerErr}}, true)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("catalog = %v, want %v", got, want)
			}
			if providerErr != nil && (len(errs) != 1 || !errors.Is(errs[0], providerErr)) {
				t.Fatalf("provider errors = %v, want %v", errs, providerErr)
			}
		}
	}
}

func TestCachedCatalogUsesSameIndexesAsFreshCatalog(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())
	config := &appcfg.Config{CacheEnabled: true, CacheTTL: appcfg.DefaultCacheTTL}
	versions := []Version{{Version: "17"}, {Version: "21"}}
	// Streaming and older releases can cache versions in arrival order.
	if err := CacheVersions(config, nativeCacheSource, versions); err != nil {
		t.Fatal(err)
	}
	cached, err := GetJdkVersions(config, true)
	if err != nil {
		t.Fatal(err)
	}
	fresh, _, err := fetchJdkVersions([]jdkProvider{catalogTestProvider{versions: versions}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cached, fresh) {
		t.Fatalf("cached catalog = %v, fresh catalog = %v", cached, fresh)
	}
}
