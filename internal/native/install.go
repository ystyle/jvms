//go:build windows

package native

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/fsutil"
	"github.com/ystyle/jvms/internal/httpclient"
	"github.com/ystyle/jvms/internal/java"
	"github.com/ystyle/jvms/internal/jdk"
)

func installPrerequisites(config *appcfg.Config) {
	if !fsutil.Exists(config.Download) {
		os.MkdirAll(config.Download, 0777)
	}
	if !fsutil.Exists(config.Store) {
		os.MkdirAll(config.Store, 0777)
	}
}

func (m *service) Install(v string) error {
	config := m.config
	if err := jdk.ValidateVersionIdentifier(v); err != nil {
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
