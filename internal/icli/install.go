package icli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/file"
	"github.com/ystyle/jvms/utils/java"
	"github.com/ystyle/jvms/utils/jdk"
	"github.com/ystyle/jvms/utils/web"
)

func install(config *models.Config) *cli.Command {
	return &cli.Command{
		Name:      "install",
		ShortName: "i",
		Usage:     "Install available remote jdk",
		Action:    installFunc(config),
	}
}

func installPrerequisites(config *models.Config) {
	if !file.Exists(config.Download) {
		os.MkdirAll(config.Download, 0777)
	}
	if !file.Exists(config.Store) {
		os.MkdirAll(config.Store, 0777)
	}
}

func installFunc(config *models.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if config.Proxy != "" {
			web.SetProxy(config.Proxy)
		}
		v := c.Args().Get(0)
		if v == "" {
			return errors.New("invalid version., Type \"jvms rls\" to see what is available for install")
		}

		if jdk.IsVersionInstalled(config.Store, v) {
			fmt.Println("Version " + v + " is already installed.")
			return nil
		}
		versions, err := jdk.GetJdkVersions(config)
		if err != nil {
			return err
		}

		installPrerequisites(config)
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
		return errors.New("invalid version., Type \"jvms rls\" to see what is available for install")
	}
}
