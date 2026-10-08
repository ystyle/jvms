package platform

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type installationLister interface {
	Installed() ([]Installation, error)
}

func IsInstalled(manager installationLister, version string) (bool, error) {
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

// ResolveVersion treats bare values as versions and #N as an installed index.
func ResolveVersion(manager Provider, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("a JDK version or explicit index (#N) is required")
	}
	if !strings.HasPrefix(value, "#") {
		return value, nil
	}

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

	version := installed[index-1].Version
	fmt.Printf("Using index %d to select JDK %s\n", index, version)
	return version, nil
}

// ResolveAvailableVersion treats bare values as versions and #N as an available index.
func ResolveAvailableVersion(manager Provider, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("a JDK version or explicit available index (#N) is required")
	}
	if !strings.HasPrefix(value, "#") {
		return value, nil
	}

	index, err := strconv.Atoi(strings.TrimPrefix(value, "#"))
	if err != nil || index <= 0 {
		return "", fmt.Errorf("invalid available JDK index %q", value)
	}

	available, err := manager.Available()
	if err != nil && len(available) == 0 {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("resolve available JDK index: %w", err)
	}
	if index <= len(available) {
		version := available[index-1].Version
		fmt.Printf("Using available index %d to select JDK %s\n", index, version)
		return version, nil
	}
	return "", fmt.Errorf("invalid available JDK index %s (expected #1-#%d)", value, len(available))
}
