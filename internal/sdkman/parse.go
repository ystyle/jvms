//go:build !windows

package sdkman

import (
	"errors"
	"regexp"
	"strings"

	"github.com/ystyle/jvms/internal/jdk"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
var sdkmanIdentifierPattern = regexp.MustCompile(`\b[0-9][A-Za-z0-9._+]*-[A-Za-z0-9][A-Za-z0-9._+-]*\b`)

func sdkmanListError(output string) error {
	plain := ansiPattern.ReplaceAllString(output, "")
	if strings.Contains(plain, "INTERNET NOT REACHABLE") {
		return errors.New("SDKMAN cannot reach its API; check your internet connection or proxy settings")
	}

	if strings.Contains(plain, "Offline: only showing installed") {
		return errors.New("SDKMAN is offline and cannot list available Java versions")
	}

	return nil
}

func parseSDKMANJavaList(output string) []jdk.Version {
	seen := make(map[string]bool)
	var versions []jdk.Version
	vendor := ""
	plain := ansiPattern.ReplaceAllString(output, "")
	offline := strings.Contains(plain, "Offline: only showing installed")

	for _, line := range strings.Split(plain, "\n") {
		columns := strings.Split(line, "|")
		identifier := ""
		if len(columns) >= 4 {
			if name := strings.TrimSpace(columns[0]); name != "" {
				vendor = name
			}

			identifier = columns[len(columns)-1]
		} else if offline {
			fields := strings.Fields(line)
			if len(fields) == 2 && (fields[0] == ">" || fields[0] == "*") {
				identifier = fields[1]
			}
		}

		matches := sdkmanIdentifierPattern.FindAllString(identifier, -1)
		if len(matches) == 0 {
			continue
		}

		id := matches[len(matches)-1]
		if seen[id] {
			continue
		}

		seen[id] = true
		versions = append(versions, jdk.Version{
			Version: id,
			Url:     vendor,
		})
	}

	return versions
}
