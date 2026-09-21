package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var versionIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)

func validateVersionIdentifier(version string) error {
	if !versionIdentifierPattern.MatchString(version) {
		return fmt.Errorf("invalid JDK version identifier %q", version)
	}
	return nil
}

func validateJavaHome(path, executable string) (string, error) {
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
