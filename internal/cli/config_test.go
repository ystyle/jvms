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

func TestConfigRejectsAllChangesWhenAnyOptionIsInvalid(t *testing.T) {
	for _, invalidOption := range []string{"--cache_toggle=bad", "--cache_ttl=-1h"} {
		t.Run(invalidOption, func(t *testing.T) {
			cfg := appcfg.NewConfig()
			cfg.JavaHome = "test-original-java-home"
			before := *cfg

			ctx := configContext(t, cfg, "--java_home=test-changed-java-home", invalidOption)
			if err := configFunc(cfg)(ctx); err == nil {
				t.Fatalf("expected command to reject invalid option %s", invalidOption)
			}

			if !reflect.DeepEqual(*cfg, before) {
				t.Errorf("rejected command must leave all settings unchanged: before %+v, after %+v", before, *cfg)
			}
		})
	}
}
