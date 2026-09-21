package cli

import (
	"flag"
	"reflect"
	"testing"

	"github.com/codegangsta/cli"
	appcfg "github.com/ystyle/jvms/internal/config"
)

func configContext(t *testing.T, cfg *appcfg.Config, args ...string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("config", flag.ContinueOnError)
	for _, f := range configCommand(cfg).Flags {
		f.Apply(set)
	}
	if err := set.Parse(args); err != nil {
		t.Fatal(err)
	}
	return cli.NewContext(cli.NewApp(), set, nil)
}
func TestConfigUpdatesAndClearsSettings(t *testing.T) {
	cfg := appcfg.NewConfig()
	cfg.Proxy = "http://localhost:8080"
	ctx := configContext(t, cfg, "--java-home=C:/jdk", "--original-path=https://example.com/index.json", "--proxy=", "--resolution-priority=version,index", "--cache-toggle=false", "--cache-ttl=30m")
	if err := configFunc(cfg)(ctx); err != nil {
		t.Fatal(err)
	}
	if cfg.JavaHome != "C:/jdk" || cfg.OriginalPath != "https://example.com/index.json" || cfg.Proxy != "" || cfg.ResolutionPriority != appcfg.VersionFirst || cfg.CacheEnabled || cfg.CacheTTL != "30m" {
		t.Fatalf("unexpected settings: %+v", cfg)
	}
}
func TestConfigRejectsAllChangesWhenAnyOptionIsInvalid(t *testing.T) {
	for _, invalidOption := range []string{"--resolution-priority=bad", "--cache-toggle=bad", "--cache-ttl=-1h"} {
		t.Run(invalidOption, func(t *testing.T) {
			cfg := appcfg.NewConfig()
			cfg.JavaHome = "test-original-java-home"
			before := *cfg

			// Submit a valid Java home change together with an invalid option.
			// The entire command must fail without applying either change.
			ctx := configContext(t, cfg, "--java-home=test-changed-java-home", invalidOption)
			if err := configFunc(cfg)(ctx); err == nil {
				t.Fatalf("expected command to reject invalid option %s", invalidOption)
			}
			if cfg.JavaHome != before.JavaHome {
				t.Errorf("rejected command applied the Java home change: got %q, want original %q", cfg.JavaHome, before.JavaHome)
			}
			if !reflect.DeepEqual(*cfg, before) {
				t.Errorf("rejected command must leave all settings unchanged: before %+v, after %+v", before, *cfg)
			}
		})
	}
}
func TestConfigPersistsThroughApplicationLifecycle(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())
	cfg := appcfg.NewConfig()
	app := cli.NewApp()
	app.Commands = Commands(cfg, &recordingManager{})
	app.Before = func(*cli.Context) error { return cfg.Load() }
	app.After = func(*cli.Context) error { return cfg.Save() }
	if err := app.Run([]string{"jvms", "config", "--cache-toggle=false", "--cache-ttl=45m", "--resolution-priority=version,index"}); err != nil {
		t.Fatal(err)
	}
	loaded := appcfg.NewConfig()
	if err := loaded.Load(); err != nil {
		t.Fatal(err)
	}
	if loaded.CacheEnabled || loaded.CacheTTL != "45m" || loaded.ResolutionPriority != appcfg.VersionFirst {
		t.Fatalf("not persisted: %+v", loaded)
	}
}
