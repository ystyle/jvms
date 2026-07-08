package models

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/tucnak/store"
	"github.com/ystyle/jvms/utils/file"
	"github.com/ystyle/jvms/utils/web"
)

const (
	ProjectConfigDir    = "jvms"
	ConfigFileName      = "jvms.json"
	DefaultOriginalPath = "https://raw.githubusercontent.com/ystyle/jvms/new/jdkdlindex.json"
)

var DefaultJavaHome = filepath.Join(os.Getenv("ProgramFiles"), "jdk")

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
func NewConfig() *Config {
	return &Config{}
}

func (c *Config) JavaHomeNotSet() bool {
	return c.JavaHome == ""
}

func (c *Config) Save() error {
	if err := store.Save(ConfigFileName, c); err != nil {
		return errors.New("failed to save the config:" + err.Error())
	}

	return nil
}

// Load initializes the config with default values if they are not set
func (c *Config) Load() *Config {
	if c.OriginalPath == "" {
		c.OriginalPath = DefaultOriginalPath
	}

	if c.JavaHome == "" {
		c.JavaHome = DefaultJavaHome
	}

	dir := file.GetCurrentPath()
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
		web.SetProxy(c.Proxy)
	}

	return c
}
