package jdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateVersionIdentifier(t *testing.T) {
	for _, valid := range []string{"21.0.8-tem", "8.0.472.fx-zulu", "22.3.r17-grl"} {
		if err := ValidateVersionIdentifier(valid); err != nil {
			t.Errorf("ValidateVersionIdentifier(%q): %v", valid, err)
		}
	}
	for _, invalid := range []string{"", "../java", "21 tem", "/tmp/jdk"} {
		if err := ValidateVersionIdentifier(invalid); err == nil {
			t.Errorf("ValidateVersionIdentifier(%q) unexpectedly succeeded", invalid)
		}
	}
}

func TestValidateJavaHome(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateJavaHome(home, "java"); err == nil {
		t.Fatal("validateJavaHome accepted a directory without bin/java")
	}
	if err := os.WriteFile(filepath.Join(home, "bin", "java"), nil, 0755); err != nil {
		t.Fatal(err)
	}
	got, err := ValidateJavaHome(home, "java")
	if err != nil {
		t.Fatal(err)
	}
	if got != home {
		t.Fatalf("ValidateJavaHome() = %q, want %q", got, home)
	}
}
