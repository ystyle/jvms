package jdk

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var versionIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)

// ValidateVersionIdentifier checks that a version is safe to use as a directory name.
func ValidateVersionIdentifier(version string) error {
	if !versionIdentifierPattern.MatchString(version) {
		return fmt.Errorf("invalid JDK version identifier %q", version)
	}

	return nil
}

// ValidateJavaHome resolves a JDK home and checks for its compiler executable.
func ValidateJavaHome(path, executable string) (string, error) {
	home, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", fmt.Errorf("resolve JDK home: %w", err)
	}

	info, err := os.Stat(filepath.Join(home, "bin", executable))
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a JDK home (missing bin/%s)", path, executable)
	}

	return home, nil
}
