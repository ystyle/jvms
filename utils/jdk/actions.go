package jdk

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/file"
	"github.com/ystyle/jvms/utils/java"
	"github.com/ystyle/jvms/utils/web"
)

// PerformSwitch repoints the JavaHome symlink at the given (already
// installed) JDK version
func PerformSwitch(config *models.Config, v string) error {
	if !IsVersionInstalled(config.Store, v) {
		return fmt.Errorf("jdk %s is not installed", v)
	}

	// Remove existing symlink if it exists
	if file.Exists(config.JavaHome) {
		if err := os.Remove(config.JavaHome); err != nil {
			return fmt.Errorf("failed to remove existing JavaHome symlink at %s: %w\n\nPossible reasons:\n"+
				"- Insufficient permissions (try running as administrator)\n"+
				"- File is in use by another process\n"+
				"- Path points to a directory instead of a symlink\n"+
				"Please manually remove it and try again", config.JavaHome, err)
		}
	}

	// Create or update the symlink
	cmd := exec.Command("cmd", "/C", "setx", "JAVA_HOME", config.JavaHome, "/M")
	if err := cmd.Run(); err != nil {
		return errors.New("set Environment variable `JAVA_HOME` failure: Please run as admin user")
	}

	if err := os.Symlink(filepath.Join(config.Store, v), config.JavaHome); err != nil {
		return errors.New("Switch jdk failed, " + err.Error())
	}

	config.CurrentJDKVersion = v
	return nil
}

// PerformInstall downloads, unzips, and installs the given JDK version.
// Used by both `jvms install <version>` and the install picker's Action.
func PerformInstall(config *models.Config, v string) error {
	versions, err := GetJdkVersions(config)
	if err != nil {
		return err
	}
	for _, version := range versions {
		if version.Version == v {
			dlzipfile, success := web.GetJDK(config.Download, v, version.Url)
			if success {
				fmt.Printf("Installing JDK %s ...\n", v)

				// Extract jdk to the temp directory
				jdktempfile := filepath.Join(config.Download, fmt.Sprintf("%s_temp", v))
				if file.Exists(jdktempfile) {
					err := os.RemoveAll(jdktempfile)
					if err != nil {
						panic(err)
					}
				}
				err := file.Unzip(dlzipfile, jdktempfile)
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
			}
			return nil
		}
	}
	return fmt.Errorf("JDK version %s not found in available versions", v)

}
