package jdk

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ystyle/jvms/internal/fsutil"
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
	// Existing folders may contain spaces, but must be direct children of the store.
	if version == "." || !filepath.IsLocal(version) || filepath.Base(version) != version || strings.ContainsAny(version, `/\:`) {
		return false
	}

	path := filepath.Join(root, version, "bin", "javac.exe")
	isInstalled := fsutil.Exists(path)
	return isInstalled
}
