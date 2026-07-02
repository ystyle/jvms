package icli

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/baneeishaque/adoptium_jdk_go"
	"github.com/ystyle/jvms/internal/entity"
	"github.com/ystyle/jvms/utils/jdk"
	"github.com/ystyle/jvms/utils/web"
)

// getSImilarAvailableVersions by version to support version not found error
// func getSimilarAvailableVersions(version string) {

// }

func getJavaHome(jdkTempFile string) string {
	var javaHome string
	fs.WalkDir(os.DirFS(jdkTempFile), ".", func(path string, d fs.DirEntry, err error) error {
		if filepath.Base(path) == "javac.exe" {
			temPath := strings.Replace(path, "bin/javac.exe", "", -1)
			javaHome = filepath.Join(jdkTempFile, temPath)
			return fs.SkipDir
		}
		return nil
	})
	return javaHome
}

func getJdkVersions(config *entity.Config) ([]entity.JdkVersion, error) {
	jsonContent, err := web.GetRemoteTextFile(config.Originalpath)
	if err != nil {
		return nil, err
	}
	var versions []entity.JdkVersion
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
		versions = append(versions, entity.JdkVersion{Version: fileVersion, Url: adoptiumJdkUrl})
	}

	//Azul JDKs
	azulJdks := jdk.AzulJDKs()
	for _, azulJdk := range azulJdks {
		versions = append(versions, entity.JdkVersion{Version: azulJdk.ShortName, Url: azulJdk.DownloadURL})
	}

	//fmt.Println(versions)
	return versions, nil
}
