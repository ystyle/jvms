package models

import (
	"os"
	"path/filepath"

	"github.com/ystyle/jvms/utils/file"
)

const (
	ProjectConfigDir    = "jvms"
	ConfigFileName      = "jvms.json"
	DefaultOriginalPath = "https://raw.githubusercontent.com/ystyle/jvms/new/jdkdlindex.json"
)

var (
	DefaultJavaHome = filepath.Join(os.Getenv("ProgramFiles"), "jdk")
)

type Config struct {
	JavaHome          string `json:"java_home"`
	CurrentJDKVersion string `json:"current_jdk_version"`
	OriginalPath      string `json:"original_path"`
	Proxy             string `json:"proxy"`

	// Runtime values (not persisted)
	Store    string `json:"-"`
	Download string `json:"-"`
}

// NewConfig creates a new Config instance with default values
func NewConfigPtr() *Config {
	return &Config{}
}

// IdempotentSeed initializes the config with default values if they are not set
func (c *Config) IdempotentSeed() *Config {
	if c.OriginalPath == "" {
		c.SetOriginalPath(DefaultOriginalPath)
	}

	if c.JavaHome == "" {
		c.SetJavaHome(DefaultJavaHome)
	}

	dir := file.GetCurrentPath()
	if c.Store == "" {
		c.SetStore(filepath.Join(dir, "store"))
		// Override store path by storepath file
		if storepath, err := os.ReadFile(filepath.Join(dir, "storepath")); err == nil {
			c.SetStore(string(storepath))
		}
	}

	if c.Download == "" {
		c.SetDownload(filepath.Join(dir, "download"))
	}
	return c
}

func (c *Config) JavaHomeNotSet() bool {
	return c.JavaHome == ""
}

func (c *Config) SetJavaHome(javaHome string) *Config {
	c.JavaHome = javaHome
	return c
}

func (c *Config) SetCurrentJDKVersion(version string) *Config {
	c.CurrentJDKVersion = version
	return c
}

func (c *Config) SetOriginalPath(path string) *Config {
	c.OriginalPath = path
	return c
}

func (c *Config) SetDownload(download string) *Config {
	c.Download = download
	return c
}

func (c *Config) SetStore(store string) *Config {
	c.Store = store
	return c
}

func (c *Config) SetProxy(proxy string) *Config {
	c.Proxy = proxy
	return c
}
