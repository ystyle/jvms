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
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)
	return filepath.Join(dir, ProjectConfigDir, ConfigFileName)
}

func writeConfig(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadSetsDefaultsForUnsetFields(t *testing.T) {
	for _, data := range []string{"", `{}`, `{"java_home":"","original_path":"","resolution_priority":"","cache_ttl":""}`} {
		t.Run(data, func(t *testing.T) {
			path := isolateConfig(t)
			if data != "" {
				writeConfig(t, path, data)
			}

			var c Config
			if err := c.Load(); err != nil {
				t.Fatal(err)
			}

			if c.JavaHome != DefaultJavaHome || c.OriginalPath != DefaultOriginalPath || c.CacheTTL != DefaultCacheTTL || !c.CacheEnabled {
				t.Fatalf("unexpected defaults: %+v", c)
			}

			if c.Store == "" || c.Download == "" {
				t.Fatalf("missing runtime paths: %+v", c)
			}
		})
	}
}

func TestConfigRoundTripAndRepeatedLoad(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		path := isolateConfig(t)
		want := Config{
			JavaHome:          "C:/Java/current",
			CurrentJDKVersion: "21",
			OriginalPath:      "https://example.com/index.json",
			CacheEnabled:      false,
			CacheTTL:          "45m",
			Store:             "test-store",
			Download:          "test-download",
		}
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}

		if legacy {
			data = []byte(`{"java_home":"C:/Java/current","current_jdk_version":"21","original_path":"https://example.com/index.json","proxy":""}`)
			want.CacheEnabled, want.CacheTTL = true, DefaultCacheTTL
		}

		writeConfig(t, path, string(data))
		c := Config{Store: want.Store, Download: want.Download}

		for i := 0; i < 2; i++ {
			if err := c.Load(); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(c, want) {
				t.Fatalf("legacy=%t load %d: got %+v, want %+v", legacy, i, c, want)
			}
		}

		if got, err := os.ReadFile(path); err != nil || string(got) != string(data) {
			t.Fatalf("Load changed saved config: %q, %v", got, err)
		}

		if err := c.Save(); err != nil {
			t.Fatal(err)
		}

		if err := c.Load(); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(c, want) {
			t.Fatalf("round trip changed settings: %+v", c)
		}

		data, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var saved map[string]any
		if err := json.Unmarshal(data, &saved); err != nil {
			t.Fatal(err)
		}

		if len(saved) != 6 || saved["Store"] != nil || saved["store"] != nil || saved["Download"] != nil || saved["download"] != nil {
			t.Fatalf("unexpected persisted fields: %v", saved)
		}
	}
}

func TestLoadIgnoresRetiredResolutionPriority(t *testing.T) {
	for _, priority := range []string{"version,index", "index,version", "invalid"} {
		t.Run(priority, func(t *testing.T) {
			path := isolateConfig(t)
			data := `{"java_home":"C:/Java/current","resolution_priority":"` + priority + `","cache_enabled":false,"cache_ttl":"45m"}`
			writeConfig(t, path, data)

			var config Config
			if err := config.Load(); err != nil {
				t.Fatal(err)
			}

			if config.JavaHome != "C:/Java/current" || config.CacheEnabled || config.CacheTTL != "45m" {
				t.Fatalf("settings changed while ignoring retired field: %+v", config)
			}

			if err := config.Save(); err != nil {
				t.Fatal(err)
			}

			savedData, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			var saved map[string]any
			if err := json.Unmarshal(savedData, &saved); err != nil {
				t.Fatal(err)
			}

			if _, exists := saved["resolution_priority"]; exists {
				t.Fatalf("retired field persisted: %s", savedData)
			}
		})
	}
}

func TestParseCacheTTL(t *testing.T) {
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
