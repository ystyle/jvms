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
	if jdk.IsVersionInstalled(config.Store, v) {
		return nil
	}

	if err := jdk.ValidateVersionIdentifier(v); err != nil {
		return err
	}

	if config.Proxy != "" {
		httpclient.SetProxy(config.Proxy)
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
			} else {
				return fmt.Errorf("could not download JDK %s; check your connection or configure a proxy with `jvms config --proxy=<url>`", v)
			}

			return nil
		}
	}

	return errors.New("invalid version., Type \"jvms rls\" to see what is available for install")
}
