package sdkman

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	appconfig "github.com/ystyle/jvms/internal/config"
)

func TestEnsureRejectsRootBeforeDownloading(t *testing.T) {
	if os.Geteuid() == 0 {
		config := appconfig.NewConfig()
		config.AllowAutoInstallSDKMAN = true

		manager := &service{
			config: config,
			dir:    t.TempDir(),
			download: func(string) ([]byte, error) {
				t.Fatal("Ensure attempted a download as root")
				return nil, nil
			},
		}

		err := manager.Ensure()
		if err == nil || !strings.Contains(err.Error(), "regular user") {
			t.Fatalf("Ensure() error = %v, want root rejection", err)
		}

		return
	}

	unshare, err := exec.LookPath("unshare")
	if err != nil {
		t.Skip("root rejection check requires unshare")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	namespaceArgs := []string{"--user", "--map-root-user", "--"}
	probeArgs := append(append([]string{}, namespaceArgs...), "true")
	if output, err := exec.CommandContext(ctx, unshare, probeArgs...).CombinedOutput(); err != nil {
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}

		t.Skipf("user namespaces unavailable: %v: %s", err, output)
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	// Root exists only inside this subprocess's user namespace.
	args := append(namespaceArgs, executable, "-test.run=^TestEnsureRejectsRootBeforeDownloading$", "-test.v")
	command := exec.CommandContext(ctx, unshare, args...)
	command.WaitDelay = time.Second
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("root bootstrap check failed: %v\n%s", err, output)
	}
}
