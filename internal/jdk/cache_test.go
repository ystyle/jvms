package jdk

import (
	"github.com/tucnak/store"
	"reflect"
	"testing"
	"time"

	appcfg "github.com/ystyle/jvms/internal/config"
)

func TestCacheUsesConfiguredTTL(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())
	config := &appcfg.Config{CacheEnabled: true, CacheTTL: "30m"}
	want := []Version{{Version: "21", Url: "https://example.com/jdk.zip"}}
	store.Init(appcfg.ProjectConfigDir)
	if err := store.Save(cacheFileName, &JdkVersionCache{Versions: want, Source: "ttl-test", LastUpdated: time.Now().Add(-time.Hour).Unix()}); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCachedVersions(config, "ttl-test", false); err == nil {
		t.Fatal("LoadCachedVersions accepted a stale cache entry")
	}
	if got, err := LoadCachedVersions(config, "ttl-test", true); err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("stale fallback = %v, %v; want %v", got, err, want)
	}
	config.CacheTTL = "2h"
	if got, err := LoadCachedVersions(config, "ttl-test", false); err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("same entry with longer TTL = %v, %v; want %v", got, err, want)
	}
}

func TestLegacyCacheAndStreamingCompatibility(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())
	cfg := appcfg.NewConfig()
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}
	store.Init(appcfg.ProjectConfigDir)
	// The upstream format has no source field.
	if err := store.Save(cacheFileName, map[string]any{"versions": []Version{{Version: "21", Url: "https://example.com/jdk.zip"}}, "last_updated": time.Now().Unix()}); err != nil {
		t.Fatal(err)
	}
	versions, err := LoadCachedVersions(cfg, nativeCacheSource, false)
	if err != nil || len(versions) != 1 || versions[0].Version != "21" {
		t.Fatalf("legacy cache: %v, %v", versions, err)
	}
	var received []Version
	done := false
	for event := range StreamJdkVersions(cfg) {
		if event.Err != nil {
			t.Fatal(event.Err)
		}
		if event.Done {
			done = true
		} else {
			received = append(received, event.Version)
		}
	}
	if !done || !reflect.DeepEqual(received, versions) {
		t.Fatalf("stream = %v, done=%t", received, done)
	}
	if _, err := LoadCachedVersions(cfg, "another-provider", false); err == nil {
		t.Fatal("wrong source accepted")
	}
	cfg.CacheEnabled = false
	if err := CacheVersions(cfg, nativeCacheSource, []Version{{Version: "17"}}); err != nil {
		t.Fatal(err)
	}
	cfg.CacheEnabled = true
	unchanged, err := LoadCachedVersions(cfg, nativeCacheSource, false)
	if err != nil || !reflect.DeepEqual(unchanged, versions) {
		t.Fatalf("disabled cache overwrote entries: %v, %v", unchanged, err)
	}
}
