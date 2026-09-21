package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tucnak/store"
	"github.com/ystyle/jvms/internal/fsutil"
	"github.com/ystyle/jvms/internal/httpclient"
)

const (
	ProjectConfigDir          = "jvms"
	ConfigFileName            = "jvms.json"
	DefaultResolutionPriority = IndexFirst
	DefaultCacheTTL           = "24h"
	DefaultOriginalPath       = "https://raw.githubusercontent.com/ystyle/jvms/new/jdkdlindex.json"
)

var DefaultJavaHome = filepath.Join(os.Getenv("ProgramFiles"), "jdk")

type Config struct {
	JavaHome          string `json:"java_home"`
	CurrentJDKVersion string `json:"current_jdk_version"`
	OriginalPath      string `json:"original_path"`
	Proxy             string `json:"proxy"`

	ResolutionPriority ResolutionPriority `json:"resolution_priority"`
	CacheEnabled       bool               `json:"cache_enabled"`
	CacheTTL           string             `json:"cache_ttl"`

	// Runtime values (not persisted)
	Store      string `json:"-"`
	Download   string `json:"-"`
	loadFailed bool
}

// NewConfig creates a new Config instance with default values
func NewConfig() *Config {
	return &Config{ResolutionPriority: DefaultResolutionPriority, CacheEnabled: true, CacheTTL: DefaultCacheTTL}
}

func (c *Config) JavaHomeNotSet() bool {
	return c.JavaHome == ""
}

func (c *Config) Save() error {
	// CLI shutdown also runs after a failed startup; preserve the original file.
	if c.loadFailed {
		return nil
	}
	if err := store.Save(ConfigFileName, c); err != nil {
		return errors.New("failed to save the config:" + err.Error())
	}

	return nil
}

// applyDefaults initializes existing settings and runtime paths.
func (c *Config) applyDefaults() {
	if c.OriginalPath == "" {
		c.OriginalPath = DefaultOriginalPath
	}

	if c.JavaHome == "" {
		c.JavaHome = DefaultJavaHome
	}

	dir := fsutil.GetCurrentPath()
	if c.Store == "" {
		c.Store = filepath.Join(dir, "store")
		// Override store path by storepath file
		if storepath, err := os.ReadFile(filepath.Join(dir, "storepath")); err == nil {
			c.Store = (string(storepath))
		}
	}

	if c.Download == "" {
		c.Download = filepath.Join(dir, "download")
	}

	if c.Proxy != "" {
		httpclient.SetProxy(c.Proxy)
	}

}

// Load owns store initialization and overlays persisted settings on policy defaults.
func (c *Config) Load() error {
	c.loadFailed = true
	next := *NewConfig()
	next.Store, next.Download = c.Store, c.Download
	store.Register("json", marshalFunc, json.Unmarshal)
	store.Init(ProjectConfigDir)
	err := store.Load(ConfigFileName, &next)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if next.ResolutionPriority, err = ParseResolutionPriority(string(next.ResolutionPriority)); err != nil {
		return err
	}
	if next.CacheTTL == "" {
		next.CacheTTL = DefaultCacheTTL
	}
	if next.CacheTTL, err = ParseCacheTTL(next.CacheTTL); err != nil {
		return err
	}
	next.applyDefaults()
	*c = next
	return nil
}

func marshalFunc(v interface{}) ([]byte, error) { return json.MarshalIndent(v, "", "    ") }
