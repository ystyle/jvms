package jdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/baneeishaque/adoptium_jdk_go"
	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/models"
	"github.com/ystyle/jvms/utils/file"
	"github.com/ystyle/jvms/utils/web"
)

func ResolveJdkVersion(c *cli.Context, config *models.Config, v string) (string, error) {
	// If the user has specified the --as_path or -p flag, treat the argument as a direct path
	if c.Bool("as_path") || c.Bool("p") {
		return v, nil
	}

	// If the input is numeric, resolve it as an index or numeric version.
	index, err := strconv.Atoi(v)
	if err == nil && index > 0 {
		// If not as_path, try index expansion
		installed := GetInstalled(config.Store)
		if len(installed) == 0 {
			return "", errors.New("no JDK installations found")
		}
		// Check if index is within valid range
		if index <= len(installed) {
			// Valid index, use it to select JDK
			v = installed[index-1]
			fmt.Printf("Using index %d to select JDK %s\n", index, v)
		} else {
			// Index out of range, check if there's a version folder with this numeric name (e.g., "17", "21")
			// Keep the original input as version name
			if IsVersionInstalled(config.Store, v) {
				// Version folder with numeric name exists, proceed with it
				fmt.Printf("Using version name %s\n", v)
			} else {
				// Neither valid index nor matching version folder
				return "", fmt.Errorf("invalid index: %d (should be between 1 and %d) and version '%s' is not installed", index, len(installed), v)
			}
		}
	}
	return v, nil
}

func GetJdkVersions(config *models.Config) ([]models.JdkVersion, error) {
	if versions, err := loadCachedJdkVersions(config); err == nil {
		return versions, nil
	}

	jsonContent, err := web.GetRemoteTextFile(config.OriginalPath)
	if err != nil {
		return nil, err
	}
	var versions []models.JdkVersion
	err = json.Unmarshal([]byte(jsonContent), &versions)
	if err != nil {
		return nil, err
	}
	//fmt.Println(versions)
	adoptiumJdks := strings.Split(adoptium_jdk_go.ApiListReleases(), "\n")
	for _, adoptiumJdkUrl := range adoptiumJdks {
		fileSeparatorIndex := strings.LastIndex(adoptiumJdkUrl, "/")
		fileName := adoptiumJdkUrl[fileSeparatorIndex+1:]
		fileVersion := strings.TrimSuffix(fileName, ".zip")
		//fmt.Println(fileVersion)
		versions = append(versions, models.JdkVersion{Version: fileVersion, Url: adoptiumJdkUrl})
	}

	//Azul JDKs
	azulJdks := AzulJDKs()
	for _, azulJdk := range azulJdks {
		versions = append(versions, models.JdkVersion{Version: azulJdk.ShortName, Url: azulJdk.DownloadURL})
	}

	// Cache the fetched versions for future use
	if err := cacheJdkVersions(config, versions); err != nil {
		return nil, err
	}

	return versions, nil
}

func GetInstalled(root string) []string {
	list := make([]string, 0)
	files, _ := os.ReadDir(root)
	for i := len(files) - 1; i >= 0; i-- {
		if files[i].IsDir() {
			list = append(list, files[i].Name())
		}
	}
	return list
}

func IsVersionInstalled(root string, version string) bool {
	path := filepath.Join(root, version, "bin", "javac.exe")
	isInstalled := file.Exists(path)
	return isInstalled
}
