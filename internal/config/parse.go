package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
)

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

type ResolutionPriority string

const (
	VersionFirst ResolutionPriority = "version,index"
	IndexFirst   ResolutionPriority = "index,version"
)

func ParseResolutionPriority(value string) (ResolutionPriority, error) {
	priority := ResolutionPriority(strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", "")))
	if priority == "" {
		return DefaultResolutionPriority, nil
	}
	if priority != VersionFirst && priority != IndexFirst {
		return "", fmt.Errorf("resolution priority must be %q or %q", VersionFirst, IndexFirst)
	}
	return priority, nil
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
