package java

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func GetJavaHome(jdkTempFile string) string {
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
