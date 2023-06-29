package jdk

import (
	"fmt"
	"github.com/ystyle/jvms/utils/file"
	"os"
)

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
	isInstalled := file.Exists(fmt.Sprintf("%s/%s/bin/javac.exe", root, version))
	return isInstalled
}
