package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func isolateConfig(t *testing.T) string {
	t.Helper()
	// The pinned tucnak/store version uses APPDATA on Windows and
	// XDG_CONFIG_HOME on other platforms. Redirect both to a disposable root;
	// t.Setenv restores the environment when this test finishes.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	return filepath.Join(dir, ProjectConfigDir, ConfigFileName)
}

type configDefaultsCase struct {
	name string
	data string
}

type configFieldExpectation struct {
	name string
	got  string
	want string
}

func TestLoadSetsDefaultsForUnsetFields(t *testing.T) {
	for _, tc := range []configDefaultsCase{
		{name: "missing file"},
		{name: "empty object", data: `{}`},
		{name: "empty settings", data: `{"java_home":"","original_path":"","resolution_priority":"","cache_ttl":""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := isolateConfig(t)
			if tc.data != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
					t.Fatal(err)
				}
			}

			// Start with zero values to verify Load supplies the defaults itself.
			var c Config
			if err := c.Load(); err != nil {
				t.Fatal(err)
			}
			for _, field := range []configFieldExpectation{
				{"JavaHome", c.JavaHome, DefaultJavaHome},
				{"OriginalPath", c.OriginalPath, DefaultOriginalPath},
				{"ResolutionPriority", string(c.ResolutionPriority), string(DefaultResolutionPriority)},
				{"CacheTTL", c.CacheTTL, DefaultCacheTTL},
			} {
				if field.got == "" {
					t.Errorf("%s is required but was empty after Load", field.name)
					continue
				}
				if field.got != field.want {
					t.Errorf("unset %s loaded as %q, want default %q", field.name, field.got, field.want)
				}
			}
			if !c.CacheEnabled {
				t.Error("cache should be enabled by default")
			}
			if c.Store == "" || c.Download == "" {
				t.Errorf("required runtime paths missing: Store=%q, Download=%q", c.Store, c.Download)
			}
		})
	}
}

func TestLoadPreservesSavedSettingsAndIsIdempotent(t *testing.T) {
	path := isolateConfig(t)
	want := Config{
		JavaHome:           "test-java-home",
		CurrentJDKVersion:  "21",
		OriginalPath:       "https://example.com/test-index.json",
		ResolutionPriority: VersionFirst,
		CacheEnabled:       false,
		CacheTTL:           "45m",
		Store:              "test-store",
		Download:           "test-download",
	}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	// Runtime paths are supplied by the caller; user settings come from disk.
	c := Config{Store: want.Store, Download: want.Download}
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c, want) {
		t.Fatalf("saved settings changed: got %+v, want %+v", c, want)
	}
	firstLoad := c
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c, firstLoad) {
		t.Errorf("repeated Load changed config: first %+v, second %+v", firstLoad, c)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != string(data) {
		t.Errorf("Load changed saved config: %q, %v", got, err)
	}
}

func TestLegacyConfigRoundTrip(t *testing.T) {
	path := isolateConfig(t)
	legacy := map[string]string{"java_home": "C:/Java/current", "current_jdk_version": "21", "original_path": "https://example.com/index.json", "proxy": ""}
	data, _ := json.Marshal(legacy)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	c := &Config{Store: "test-store", Download: "test-download"}
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != string(data) {
		t.Fatalf("load changed existing config: %q, %v", got, err)
	}
	if c.JavaHome != legacy["java_home"] || c.CurrentJDKVersion != "21" || c.OriginalPath != legacy["original_path"] {
		t.Fatalf("legacy settings changed: %+v", c)
	}
	if c.ResolutionPriority != IndexFirst || !c.CacheEnabled || c.CacheTTL != "24h" {
		t.Fatalf("defaults changed: %+v", c)
	}
	if c.Store != "test-store" || c.Download != "test-download" {
		t.Fatalf("runtime overrides lost: %+v", c)
	}
	c.CacheEnabled = false
	c.CacheTTL = "30m"
	c.ResolutionPriority = VersionFirst
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	loaded := NewConfig()
	if err := loaded.Load(); err != nil {
		t.Fatal(err)
	}
	if loaded.CacheEnabled || loaded.CacheTTL != "30m" || loaded.ResolutionPriority != VersionFirst {
		t.Fatalf("policy did not persist: %+v", loaded)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]any
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	for key, want := range legacy {
		if saved[key] != want {
			t.Errorf("%s = %v, want %v", key, saved[key], want)
		}
	}
	if len(saved) != 7 {
		t.Fatalf("unexpected persisted fields: %v", saved)
	}
	if saved["Store"] != nil || saved["store"] != nil || saved["Download"] != nil {
		t.Fatal("runtime paths persisted")
	}
}

func TestInvalidConfigSurvivesShutdown(t *testing.T) {
	for _, input := range []string{`{"cache_ttl":"bad"}`, `{"resolution_priority":"bad"}`, `{"cache_enabled":"bad"}`, `{`} {
		t.Run(input, func(t *testing.T) {
			path := isolateConfig(t)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			c := NewConfig()
			if err := c.Load(); err == nil {
				t.Fatal("invalid config accepted")
			}
			if err := c.Save(); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != input {
				t.Fatalf("failed startup overwrote config: %s", got)
			}
		})
	}
}

func TestParsePolicies(t *testing.T) {
	for input, want := range map[string]ResolutionPriority{"": IndexFirst, " INDEX, version ": IndexFirst, "version,index": VersionFirst} {
		got, err := ParseResolutionPriority(input)
		if err != nil || got != want {
			t.Fatalf("%q: %q, %v", input, got, err)
		}
	}
	if _, err := ParseResolutionPriority("random"); err == nil {
		t.Fatal("invalid priority accepted")
	}
	for _, input := range []string{"30m", "24h", "1h30m"} {
		if _, err := ParseCacheTTL(input); err != nil {
			t.Fatal(err)
		}
	}
	for _, input := range []string{"", "0s", "-1h", "bad"} {
		if _, err := ParseCacheTTL(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}

func TestLoadDoesNotPartiallyApplyInvalidSettings(t *testing.T) {
	path := isolateConfig(t)
	c := NewConfig()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	before := *c
	if err := os.WriteFile(path, []byte(`{"java_home":"changed","cache_ttl":"invalid"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := c.Load(); err == nil {
		t.Fatal("invalid TTL accepted")
	}
	c.loadFailed = false
	if !reflect.DeepEqual(*c, before) {
		t.Fatalf("partial update: %+v", c)
	}
}
