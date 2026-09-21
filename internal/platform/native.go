//go:build windows

package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/fsutil"
	"github.com/ystyle/jvms/internal/httpclient"
	"github.com/ystyle/jvms/internal/java"
	"github.com/ystyle/jvms/internal/jdk"
	"github.com/ystyle/jvms/internal/privilege"
)

// nativeManager extracts the existing Windows operations. Other backends are deferred.
type nativeManager struct {
	config  *appcfg.Config
	isAdmin func() bool
	run     func(string, ...string) error
}

var _ Provider = (*nativeManager)(nil)

func NewProvider(config *appcfg.Config) Provider {
	return &nativeManager{config: config, isAdmin: privilege.IsAdmin,
		run: func(name string, args ...string) error { return exec.Command(name, args...).Run() }}
}
func (m *nativeManager) Name() string                         { return "JVMS native Windows" }
func (m *nativeManager) Available() ([]jdk.Version, error)    { return jdk.GetJdkVersions(m.config, true) }
func (m *nativeManager) StreamAvailable() <-chan VersionEvent { return jdk.StreamJdkVersions(m.config) }
func (m *nativeManager) Installed() ([]Installation, error) {
	var installed []Installation
	for _, v := range jdk.GetInstalled(m.config.Store) {
		if !jdk.IsVersionInstalled(m.config.Store, v) {
			continue
		}
		installed = append(installed, Installation{Version: v, Current: v == m.config.CurrentJDKVersion})
	}
	return installed, nil
}

func installPrerequisites(config *appcfg.Config) {
	if !fsutil.Exists(config.Download) {
		os.MkdirAll(config.Download, 0777)
	}
	if !fsutil.Exists(config.Store) {
		os.MkdirAll(config.Store, 0777)
	}
}

func (m *nativeManager) Install(v string) error {
	config := m.config
	if err := validateVersionIdentifier(v); err != nil {
		return err
	}

	if config.Proxy != "" {
		httpclient.SetProxy(config.Proxy)
	}
	if v == "" {
		return errors.New("invalid version., Type \"jvms rls\" to see what is available for install")
	}

	if jdk.IsVersionInstalled(config.Store, v) {
		fmt.Println("Version " + v + " is already installed.")
		return nil
	}
	versions, err := jdk.GetJdkVersions(config, false)
	if err != nil {
		return err
	}

	installPrerequisites(config)
	for _, version := range versions {
		if version.Version == v {
			dlzipfile, success := httpclient.GetJDK(config.Download, v, version.Url)
			if success {
				fmt.Printf("Installing JDK %s ...\n", v)

				// Extract jdk to the temp directory
				jdktempfile := filepath.Join(config.Download, fmt.Sprintf("%s_temp", v))
				if fsutil.Exists(jdktempfile) {
					err := os.RemoveAll(jdktempfile)
					if err != nil {
						panic(err)
					}
				}
				err := fsutil.Unzip(dlzipfile, jdktempfile)
				if err != nil {
					return fmt.Errorf("unzip failed: %w", err)
				}

				// Copy the jdk files to the installation directory
				temJavaHome := java.GetJavaHome(jdktempfile)
				err = os.Rename(temJavaHome, filepath.Join(config.Store, v))
				if err != nil {
					return fmt.Errorf("unzip failed: %w", err)
				}

				// Remove the temp directory
				// may consider keep the temp files here
				os.RemoveAll(jdktempfile)
				fmt.Printf("Installation completedly succesfully. Use: jvms switch %v, if you'd like to use this version", v)
			} else {
				fmt.Printf("\nCould not download JDK %s executable.\n\n", v)
				fmt.Println("Possible solutions:")
				fmt.Println("1. Check your internet connection")
				fmt.Println("2. Set a proxy if you're behind a firewall:")
				fmt.Println("   jvms config proxy http://127.0.0.1:1080")
				fmt.Println("   or set environment variable: set http_proxy=http://127.0.0.1:1080")
				fmt.Println("3. Try again later as the server might be temporarily unavailable")
				fmt.Println("4. Or manually download and add the JDK (see README for details)")
				return fmt.Errorf("could not download JDK %s", v)
			}
			return nil
		}
	}
	return errors.New("invalid version., Type \"jvms rls\" to see what is available for install")
}

func (m *nativeManager) Remove(v string) error {
	config := m.config
	if err := validateVersionIdentifier(v); err != nil {
		return err
	}

	if v == "" {
		return errors.New("you should input a version, Type \"jvms list\" to see what is installed")
	}
	if jdk.IsVersionInstalled(config.Store, v) {
		fmt.Printf("Remove JDK %s ...\n", v)
		if config.CurrentJDKVersion == v {
			if err := os.Remove(config.JavaHome); err != nil {
				fmt.Printf("Warning: failed to remove JavaHome symlink: %v\n", err)
			}
		}
		dir := filepath.Join(config.Store, v)
		e := os.RemoveAll(dir)
		if e != nil {
			fmt.Println("Error removing jdk " + v)
			fmt.Println("Manually remove " + dir + ".")
			return e
		} else {
			if config.CurrentJDKVersion == v {
				config.CurrentJDKVersion = ""
			}
			fmt.Printf(" done")
		}
	} else {
		fmt.Println("jdk " + v + " is not installed. Type \"jvms list\" to see what is installed.")
	}
	return nil
}

func (m *nativeManager) Ensure() error {
	config := m.config
	if !m.isAdmin() {
		return errors.New("jvms init requires administrator privileges. Please run as administrator")
	}

	err := m.run("cmd", "/C", "setx", "JAVA_HOME", config.JavaHome, "/M")
	if err != nil {
		return fmt.Errorf("failed to set JAVA_HOME environment variable to %s: %w\n\n"+
			"Possible reasons:\n"+
			"- Insufficient permissions (try running as administrator)\n"+
			"- Command execution failed\n"+
			"- Invalid path format\n"+
			"Please run Command Prompt as Administrator and try again", config.JavaHome, err)
	}
	fmt.Println("set `JAVA_HOME` Environment variable to ", config.JavaHome)
	path := fmt.Sprintf(`%s/bin;%s;%s`, config.JavaHome, os.Getenv("PATH"), fsutil.GetCurrentPath())
	err = m.run("cmd", "/C", "setx", "path", path, "/m")
	if err != nil {
		return fmt.Errorf("failed to add jvms.exe to PATH environment variable: %w\n\n"+
			"Possible reasons:\n"+
			"- Insufficient permissions (try running as administrator)\n"+
			"- PATH variable is too long (Windows has a 2048 character limit)\n"+
			"- Command execution failed\n"+
			"Please run Command Prompt as Administrator and try again", err)
	}
	fmt.Println("add jvms.exe to `path` Environment variable")
	return nil
}

func (m *nativeManager) Switch(v string) error {
	config := m.config
	if !m.isAdmin() {
		return errors.New("this command requires administrator privileges.")
	}
	if config.JavaHomeNotSet() {
		if err := m.Ensure(); err != nil {
			return err
		}
	}

	if !jdk.IsVersionInstalled(config.Store, v) {
		return fmt.Errorf("JDK %s is not installed", v)
	}
	// Create or update the symlink
	if fsutil.Exists(config.JavaHome) {
		err := os.Remove(config.JavaHome)
		if err != nil {
			return fmt.Errorf("failed to remove existing JavaHome symlink at %s: %w\n\nPossible reasons:\n"+
				"- Insufficient permissions (try running as administrator)\n"+
				"- File is in use by another process\n"+
				"- Path points to a directory instead of a symlink\n"+
				"Please manually remove it and try again", config.JavaHome, err)
		}
	}
	err := m.run("cmd", "/C", "setx", "JAVA_HOME", config.JavaHome, "/M")
	if err != nil {
		return errors.New("set Environment variable `JAVA_HOME` failure: Please run as admin user")
	}
	err = os.Symlink(filepath.Join(config.Store, v), config.JavaHome)
	if err != nil {
		return errors.New("Switch jdk failed, " + err.Error())
	}
	fmt.Println("Switch success.\nNow using JDK " + v)
	config.CurrentJDKVersion = v
	return nil
}

// SwitchPath retains the upstream path handling; native path improvements are deferred.
func (m *nativeManager) SwitchPath(path string) error { return m.Switch(path) }
