package cli

import (
	"flag"
	"testing"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

type refreshRecordingManager struct {
	recordingManager
	cacheInvalidated bool
	config           *appcfg.Config
}

func (m *refreshRecordingManager) Available() ([]jdk.Version, error) {
	_, err := jdk.LoadCachedVersions(m.config, "refresh-test", false)
	m.cacheInvalidated = err != nil
	return []jdk.Version{{Version: "21.0.4-tem"}}, nil
}

func TestRLSRefreshesCatalogBeforeListing(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("APPDATA", configDir)
	config := &appcfg.Config{CacheEnabled: true, CacheTTL: appcfg.DefaultCacheTTL}
	if err := jdk.CacheVersions(config, "refresh-test", []jdk.Version{{Version: "stale"}}); err != nil {
		t.Fatal(err)
	}

	set := flag.NewFlagSet("rls", flag.ContinueOnError)
	set.Bool("a", false, "")
	manager := &refreshRecordingManager{config: config}
	action := rls(manager).Action.(func(*cli.Context) error)
	if err := action(cli.NewContext(cli.NewApp(), set, nil)); err != nil {
		t.Fatal(err)
	}
	if !manager.cacheInvalidated {
		t.Fatal("rls did not invalidate the cached catalog before loading available versions")
	}
}
