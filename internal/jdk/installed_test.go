package jdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstalledManualFolderAndPathBoundaries(t *testing.T) {
	root := t.TempDir()
	store := filepath.Join(root, "store")
	for _, folder := range []string{store, filepath.Join(store, "jdk 17"), filepath.Join(root, "outside")} {
		if err := os.MkdirAll(filepath.Join(folder, "bin"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, "bin", "javac.exe"), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !IsVersionInstalled(store, "jdk 17") {
		t.Fatal("manual folder containing a space is not recognized")
	}
	for _, invalid := range []string{"", ".", "..", "../outside", `..\outside`, "jdk 17/.", filepath.Join(root, "outside")} {
		if IsVersionInstalled(store, invalid) {
			t.Errorf("recognized %q as a direct child of the store", invalid)
		}
	}
}
