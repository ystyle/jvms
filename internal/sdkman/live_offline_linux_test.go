package sdkman

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Only the SDKMAN subprocess and its children lose network access.
func TestSDKMANLiveOffline(t *testing.T) {
	manager := liveSDKMANService(t)
	unshare, err := exec.LookPath("unshare")
	if err != nil {
		t.Skip("offline SDKMAN check requires unshare")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespaceArgs := []string{"--user", "--map-root-user", "--net", "--"}
	probeArgs := append(append([]string{}, namespaceArgs...), "true")
	probe := exec.CommandContext(ctx, unshare, probeArgs...)
	if output, err := probe.CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			t.Fatalf("network namespace probe timed out: %v", ctx.Err())
		}

		t.Skipf("network namespaces unavailable: %v: %s", err, output)
	}

	// Enable detection in this shell even if the user's config disables it.
	script := `source "$SDKMAN_DIR/bin/sdkman-init.sh"
sdkman_healthcheck_enable=true
sdkman_curl_connect_timeout=2
sdkman_curl_max_time=3
sdkman_auto_answer=true
sdkman_colour_enable=false
sdkman_selfupdate_feature=false
sdk "$@"`
	commandArgs := append(namespaceArgs,
		"bash",
		"--noprofile",
		"--norc",
		"-c",
		script,
		"jvms-sdkman-offline-test",
		"list",
		"java",
	)

	command := exec.CommandContext(ctx, unshare, commandArgs...)
	command.Env = manager.commandEnv()
	command.WaitDelay = 2 * time.Second

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("isolated sdk list java failed: %v\n%s", err, output)
	}

	plain := ansiPattern.ReplaceAllString(string(output), "")
	if !strings.Contains(plain, "INTERNET NOT REACHABLE") || !strings.Contains(plain, "Offline: only showing installed java versions") {
		t.Fatalf("SDKMAN did not report the blocked network and offline list:\n%s", output)
	}

	if err := sdkmanListError(string(output)); err == nil {
		t.Fatalf("sdkmanListError() accepted the live offline response:\n%s", output)
	}

	expected := liveOfflineRows(t, string(output))
	assertLiveParsedVersions(t, string(output), expected)
	t.Logf("verified offline detection and %d identifiers with SDKMAN networking blocked", len(expected))
}
