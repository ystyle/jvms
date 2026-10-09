//go:build !windows

package sdkman

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Live checks bypass JVMS's cache. Use -count=1 to bypass Go's test cache.
// SDKMAN absence and -short skip these checks.
func TestSDKMANLive(t *testing.T) {
	manager := liveSDKMANService(t)

	t.Run("Version", func(t *testing.T) {
		output := runLiveSDKMAN(t, manager, "version")
		if !strings.Contains(output, "SDKMAN!") {
			t.Fatalf("unrecognized SDKMAN version response: %s", output)
		}

		t.Log(strings.TrimSpace(output))
	})

	t.Run("JavaList", func(t *testing.T) {
		output := runLiveSDKMAN(t, manager, "list", "java")
		if err := sdkmanListError(output); err != nil {
			expected := liveOfflineRows(t, output)
			assertLiveParsedVersions(t, output, expected)
			t.Logf("%v; verified %d identifiers from the live offline list", err, len(expected))
			return
		}

		expected := liveCatalogRows(t, output)
		assertLiveParsedVersions(t, output, expected)
		t.Logf("verified %d identifiers and vendor names from the live Java table", len(expected))
	})
}

func liveSDKMANService(t *testing.T) *service {
	t.Helper()

	if testing.Short() {
		t.Skip("live SDKMAN checks disabled by -short")
	}

	manager := NewService(nil).(*service)
	if _, err := os.Stat(manager.initScript()); errors.Is(err, os.ErrNotExist) {
		t.Skip("SDKMAN is not installed")
	} else if err != nil {
		t.Fatalf("inspect SDKMAN installation: %v", err)
	}

	if err := manager.checkHost(); err != nil {
		t.Fatal(err)
	}

	return manager
}

func runLiveSDKMAN(t *testing.T, manager *service, args ...string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	commandArgs := []string{
		"--noprofile",
		"--norc",
		"-c",
		sdkmanCommand,
		"jvms-sdkman-live-test",
	}
	commandArgs = append(commandArgs, args...)

	command := exec.CommandContext(ctx, "bash", commandArgs...)
	command.Env = manager.commandEnv()
	command.WaitDelay = 2 * time.Second

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("live sdk %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}

	return string(output)
}

func liveCatalogRows(t *testing.T, output string) map[string]string {
	t.Helper()

	expected := make(map[string]string)
	vendorIndex, identifierIndex, columnCount := -1, -1, 0
	vendor := ""

	// Read named table columns independently of the production identifier regex.
	for _, line := range strings.Split(ansiPattern.ReplaceAllString(output, ""), "\n") {
		if !strings.Contains(line, "|") {
			continue
		}

		columns := strings.Split(line, "|")
		if columnCount == 0 {
			for i, column := range columns {
				switch strings.TrimSpace(column) {
				case "Vendor":
					vendorIndex = i
				case "Identifier":
					identifierIndex = i
				}
			}

			if vendorIndex < 0 || identifierIndex < 0 {
				t.Fatalf("SDKMAN's Java table header changed: %s", line)
			}

			columnCount = len(columns)
			continue
		}

		if len(columns) != columnCount {
			t.Fatalf("SDKMAN's Java table row shape changed: %s", line)
		}

		if value := strings.TrimSpace(columns[vendorIndex]); value != "" {
			vendor = value
		}

		identifier := strings.TrimSpace(columns[identifierIndex])
		if vendor == "" || len(strings.Fields(identifier)) != 1 {
			t.Fatalf("invalid vendor or identifier in live Java row: %s", line)
		}

		expected[identifier] = vendor
	}

	if columnCount == 0 || len(expected) == 0 {
		t.Fatalf("SDKMAN returned no recognizable Java table rows: %s", output)
	}

	return expected
}

func liveOfflineRows(t *testing.T, output string) map[string]string {
	t.Helper()

	expected := make(map[string]string)
	plain := ansiPattern.ReplaceAllString(output, "")
	for _, line := range strings.Split(plain, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && (fields[0] == ">" || fields[0] == "*") {
			expected[fields[1]] = ""
		}
	}

	if len(expected) == 0 && !strings.Contains(plain, "None installed!") {
		t.Fatalf("unrecognized live offline installation rows: %s", output)
	}

	return expected
}

func assertLiveParsedVersions(t *testing.T, output string, expected map[string]string) {
	t.Helper()

	seen := make(map[string]bool)
	for _, version := range parseSDKMANJavaList(output) {
		vendor, exists := expected[version.Version]
		if !exists {
			t.Fatalf("parser returned %q, which is absent from SDKMAN's version rows", version.Version)
		}

		if seen[version.Version] {
			t.Fatalf("parser duplicated live identifier %q", version.Version)
		}

		if version.Url != vendor {
			t.Fatalf("vendor for %q = %q, live table says %q", version.Version, version.Url, vendor)
		}

		seen[version.Version] = true
	}

	for identifier := range expected {
		if !seen[identifier] {
			t.Fatalf("parser missed live identifier %q", identifier)
		}
	}
}
