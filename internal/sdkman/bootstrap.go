//go:build !windows

package sdkman

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	appconfig "github.com/ystyle/jvms/internal/config"
)

const maxInstallerSize = 10 << 20

func (m *service) Ensure() error {
	if _, err := os.Stat(m.initScript()); err == nil {
		if err := requireSDKMANUser(); err != nil {
			return err
		}

		if err := m.checkHost(); err != nil {
			return err
		}

		return m.Health()
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect SDKMAN installation: %w", err)
	}

	if !m.config.AllowAutoInstallSDKMAN {
		return errors.New("SDKMAN is not installed; install it manually or enable automatic installation with `jvms config --allow-auto-install-sdkman=true`")
	}

	if err := requireSDKMANUser(); err != nil {
		return err
	}

	if err := m.checkHost(); err != nil {
		return err
	}

	fmt.Printf("SDKMAN! was not found; installing it in %s ...\n", m.dir)
	installer, err := m.downloadInstaller()
	if err != nil {
		return err
	}
	defer os.Remove(installer)

	cmd := exec.Command("bash", installer)
	cmd.Env = m.commandEnv()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SDKMAN installer failed: %w", err)
	}

	if err := m.Health(); err != nil {
		return fmt.Errorf("SDKMAN was installed but is not working: %w", err)
	}

	return nil
}

func (m *service) checkHost() error {
	if m.dir == "" {
		return errors.New("cannot determine SDKMAN_DIR")
	}

	for _, name := range []string{"bash", "curl", "zip", "unzip"} {
		if _, err := exec.LookPath(name); err != nil {
			return fmt.Errorf("SDKMAN requires %s on PATH", name)
		}
	}

	return nil
}

func (m *service) downloadInstaller() (string, error) {
	installerURL, err := appconfig.ParseSDKMANInstallURL(m.config.SDKMANInstallURL)
	if err != nil {
		return "", fmt.Errorf("invalid SDKMAN installer URL: %w", err)
	}

	body, err := m.download(installerURL)
	if err != nil {
		return "", fmt.Errorf("download SDKMAN installer: %w", err)
	}

	if len(body) > maxInstallerSize {
		return "", errors.New("SDKMAN installer is unexpectedly large")
	}

	file, err := os.CreateTemp("", "jvms-sdkman-*.sh")
	if err != nil {
		return "", err
	}

	name := file.Name()
	if _, err = file.Write(body); err == nil {
		err = file.Chmod(0700)
	}

	if closeErr := file.Close(); err == nil {
		err = closeErr
	}

	if err != nil {
		os.Remove(name)
		return "", fmt.Errorf("save SDKMAN installer: %w", err)
	}

	return name, nil
}
