package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
)

func ParseSDKMANInstallURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.Scheme != "https" {
		return "", errors.New("SDKMAN installer URL must be an HTTPS URL")
	}

	return value, nil
}

func DefaultPlatformJavaHome() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("ProgramFiles"), "jdk")
	}

	dir := strings.TrimSpace(os.Getenv("SDKMAN_DIR"))
	if dir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			dir = filepath.Join(home, ".sdkman")
		}
	}

	if dir == "" {
		return ""
	}

	return filepath.Join(dir, "candidates", "java", "current")
}

// SetBool applies a boolean flag only when it is explicitly set.
func (config *Config) SetBool(c *cli.Context, name string, target *bool) error {
	if !c.IsSet(name) {
		return nil
	}

	value, err := strconv.ParseBool(c.String(name))
	if err != nil {
		return fmt.Errorf("%s must be true or false", name)
	}

	*target = value
	return nil
}

func ParseCacheTTL(value string) (string, error) {
	value = strings.TrimSpace(value)
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return "", errors.New("cache TTL must be a positive duration such as 30m or 24h")
	}

	return value, nil
}

func (c *Config) CacheDuration() time.Duration {
	value := DefaultCacheTTL
	if c != nil && c.CacheTTL != "" {
		value = c.CacheTTL
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		duration, _ = time.ParseDuration(DefaultCacheTTL)
	}

	return duration
}
