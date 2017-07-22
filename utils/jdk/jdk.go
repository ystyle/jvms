package jdk

import (
	"io/ioutil"
	"github.com/ystyle/jvms/utils/file"
	"fmt"
)

func GetInstalled(root string) []string {
	list := make([]string, 0)
	files, _ := ioutil.ReadDir(root)
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
