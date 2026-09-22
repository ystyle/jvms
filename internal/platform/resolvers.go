package platform

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	appconfig "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

type installationLister interface {
	Installed() ([]Installation, error)
}

func IsInstalled(manager installationLister, version string) (bool, error) {
	if err := jdk.ValidateVersionIdentifier(version); err != nil {
		return false, err
	}

	installed, err := manager.Installed()
	if err != nil {
		return false, err
	}

	for _, item := range installed {
		if item.Version == version {
			return true, nil
		}
	}
	return false, nil
}

// ResolveVersion applies the configured precedence to ambiguous numeric arguments.
func ResolveVersion(manager Provider, value string, allowUninstalled bool, priority appconfig.ResolutionPriority) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("a JDK version or index number is required")
	}

	if strings.HasPrefix(value, "#") {
		return resolveExplicitIndex(manager, value)
	}

	index, err := strconv.Atoi(value)
	if err != nil || index <= 0 {
		return value, nil
	}

	installed, err := manager.Installed()
	if err != nil {
		return "", err
	}

	exactVersion := false
	for _, item := range installed {
		if item.Version == value {
			exactVersion = true
			break
		}
	}

	if priority == "" {
		priority = appconfig.DefaultResolutionPriority
	}
	if priority != appconfig.VersionFirst && priority != appconfig.IndexFirst {
		return "", fmt.Errorf("unsupported resolution priority %q", priority)
	}
	if priority == appconfig.VersionFirst && (exactVersion || allowUninstalled) {
		return value, nil
	}
	if index <= len(installed) {
		return selectIndex(installed, index), nil
	}
	if exactVersion || allowUninstalled {
		return value, nil
	}

	return "", fmt.Errorf("invalid index %d (expected 1-%d) and version %q is not installed", index, len(installed), value)
}

// ResolveAvailableVersion expands indexes from the catalog shown by `jvms rls`.
func ResolveAvailableVersion(manager Provider, value string, priority appconfig.ResolutionPriority) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("a JDK version or available index number is required")
	}

	explicitIndex := strings.HasPrefix(value, "#")
	number := strings.TrimPrefix(value, "#")
	index, err := strconv.Atoi(number)
	if err != nil || index <= 0 {
		if explicitIndex {
			return "", fmt.Errorf("invalid available JDK index %q", value)
		}
		return value, nil
	}

	if priority == "" {
		priority = appconfig.DefaultResolutionPriority
	}
	if priority != appconfig.VersionFirst && priority != appconfig.IndexFirst {
		return "", fmt.Errorf("unsupported resolution priority %q", priority)
	}
	if !explicitIndex && priority == appconfig.VersionFirst {
		installed, err := IsInstalled(manager, value)
		if err != nil {
			return "", err
		}
		if installed {
			return value, nil
		}
	}

	available, err := manager.Available()
	if err != nil && len(available) == 0 {
		return "", err
	}
	if !explicitIndex && priority == appconfig.VersionFirst {
		for _, candidate := range available {
			if candidate.Version == value {
				return value, nil
			}
		}
	}
	if err != nil {
		return "", fmt.Errorf("resolve available JDK index: %w", err)
	}
	if index <= len(available) {
		version := available[index-1].Version
		fmt.Printf("Using available index %d to select JDK %s\n", index, version)
		return version, nil
	}
	if explicitIndex {
		return "", fmt.Errorf("invalid available JDK index %s (expected #1-#%d)", value, len(available))
	}

	return value, nil
}

func resolveExplicitIndex(manager Provider, value string) (string, error) {
	index, err := strconv.Atoi(strings.TrimPrefix(value, "#"))
	if err != nil || index <= 0 {
		return "", fmt.Errorf("invalid installed JDK index %q", value)
	}

	installed, err := manager.Installed()
	if err != nil {
		return "", err
	}
	if index > len(installed) {
		return "", fmt.Errorf("invalid installed JDK index %s (expected #1-#%d)", value, len(installed))
	}

	return selectIndex(installed, index), nil
}

func selectIndex(installed []Installation, index int) string {
	version := installed[index-1].Version
	fmt.Printf("Using index %d to select JDK %s\n", index, version)
	return version
}
