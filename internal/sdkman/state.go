//go:build !windows

package sdkman

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ystyle/jvms/internal/jdk"
)

func (m *service) Installed() ([]jdk.Installation, error) {
	if err := m.requireInstalled(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(m.candidatesDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read SDKMAN Java installations: %w", err)
	}

	current, err := m.current()
	if err != nil {
		return nil, err
	}

	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Name() == "current" || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if isSDKMANJavaInstallation(filepath.Join(m.candidatesDir(), entry.Name())) {
			versions = append(versions, entry.Name())
		}
	}

	slices.Sort(versions)
	slices.Reverse(versions)

	installed := make([]jdk.Installation, len(versions))
	for i, version := range versions {
		installed[i] = jdk.Installation{
			Version: version,
			Current: version == current,
		}
	}

	return installed, nil
}

func isSDKMANJavaInstallation(candidate string) bool {
	for _, executable := range []string{
		filepath.Join(candidate, "bin", "java"),
		filepath.Join(candidate, "Contents", "Home", "bin", "java"),
	} {
		if info, err := os.Stat(executable); err == nil && info.Mode().IsRegular() {
			return true
		}
	}

	return false
}

func (m *service) current() (string, error) {
	current := filepath.Join(m.candidatesDir(), "current")
	target, err := os.Readlink(current)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("read SDKMAN current Java version: %w", err)
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(current), target)
	}

	target, err = filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolve SDKMAN current Java version: %w", err)
	}

	candidates, err := filepath.Abs(m.candidatesDir())
	if err != nil {
		return "", fmt.Errorf("resolve SDKMAN candidates directory: %w", err)
	}

	if filepath.Dir(filepath.Clean(target)) != filepath.Clean(candidates) {
		return "", nil
	}

	return filepath.Base(filepath.Clean(target)), nil
}
