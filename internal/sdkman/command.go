//go:build !windows

package sdkman

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const sdkmanCommand = `source "$SDKMAN_DIR/bin/sdkman-init.sh"
sdkman_auto_answer=true
sdkman_colour_enable=false
sdkman_selfupdate_feature=false
sdk "$@"`

// run invokes SDKMAN with positional Bash arguments to prevent shell injection.
func (m *service) run(interactive bool, args ...string) (string, error) {
	if err := m.checkHost(); err != nil {
		return "", err
	}

	if err := m.requireInstalled(); err != nil {
		return "", err
	}

	commandArgs := []string{
		"--noprofile",
		"--norc",
		"-c",
		sdkmanCommand,
		"jvms-sdkman",
	}
	commandArgs = append(commandArgs, args...)

	cmd := exec.Command("bash", commandArgs...)
	cmd.Env = m.commandEnv()

	var output bytes.Buffer
	if interactive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(output.String())
		if message != "" {
			return "", fmt.Errorf("sdk %s failed: %s: %w", strings.Join(args, " "), message, err)
		}

		return "", fmt.Errorf("sdk %s failed: %w", strings.Join(args, " "), err)
	}

	return output.String(), nil
}

func (m *service) commandEnv() []string {
	values := map[string]string{
		"SDKMAN_DIR": m.dir,
		"PAGER":      "cat",
	}

	if m.config != nil && strings.TrimSpace(m.config.Proxy) != "" {
		values["HTTP_PROXY"] = m.config.Proxy
		values["HTTPS_PROXY"] = m.config.Proxy
		values["http_proxy"] = m.config.Proxy
		values["https_proxy"] = m.config.Proxy
	}

	return replaceEnv(os.Environ(), values)
}

func replaceEnv(environment []string, values map[string]string) []string {
	result := make([]string, 0, len(environment)+len(values))
	for _, entry := range environment {
		key, _, _ := strings.Cut(entry, "=")
		if _, replaced := values[key]; !replaced {
			result = append(result, entry)
		}
	}

	for key, value := range values {
		result = append(result, key+"="+value)
	}

	return result
}
