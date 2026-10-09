//go:build !windows

package sdkman

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ystyle/jvms/internal/jdk"
)

func TestParseSDKMANJavaList(t *testing.T) {
	input := `
================================================================================
Available Java Versions for Linux 64bit
================================================================================
 Vendor         | Use | Version            | Identifier
--------------------------------------------------------------------------------
 Corretto       |     | 21.0.12            | 21.0.12-amzn
                |     | 17.0.20            | 17.0.20-amzn
 Temurin        | >>> | 21.0.8             | 21.0.8-tem
 Zulu           |     | 17.0.16            | 17.0.16-zulu
 Temurin        |     | 21.0.8             | 21.0.8-tem
 Liberica       |     | 26.0.2-fx+1.1      | 26.0.2-fx+1.1-librca
--------------------------------------------------------------------------------
 $ sdk install java                 install the default: 25.0.4-tem
`
	want := []jdk.Version{
		{
			Version: "21.0.12-amzn",
			Url:     "Corretto",
		},
		{
			Version: "17.0.20-amzn",
			Url:     "Corretto",
		},
		{
			Version: "21.0.8-tem",
			Url:     "Temurin",
		},
		{
			Version: "17.0.16-zulu",
			Url:     "Zulu",
		},
		{
			Version: "26.0.2-fx+1.1-librca",
			Url:     "Liberica",
		},
	}

	got := parseSDKMANJavaList(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSDKMANJavaList() = %#v, want %#v", got, want)
	}
}

func TestParseSDKMANJavaListStripsANSI(t *testing.T) {
	input := "\x1b[32m Temurin |     | 21.0.8 | tem | | 21.0.8-tem\x1b[0m\n"
	got := parseSDKMANJavaList(input)
	if len(got) != 1 || got[0].Version != "21.0.8-tem" {
		t.Fatalf("parseSDKMANJavaList() = %#v", got)
	}
}

func TestSDKMANOfflineListReturnsInstalledVersionAndError(t *testing.T) {
	output := `==== INTERNET NOT REACHABLE! ===================================================

 SDKMAN! needs an active internet connection for full functionality.

--------------------------------------------------------------------------------
Offline: only showing installed java versions
--------------------------------------------------------------------------------
 > 25.0.2-open
--------------------------------------------------------------------------------`
	versions := parseSDKMANJavaList(output)
	if len(versions) != 1 || versions[0].Version != "25.0.2-open" {
		t.Fatalf("parseSDKMANJavaList() = %#v", versions)
	}

	if err := sdkmanListError(output); err == nil {
		t.Fatal("sdkmanListError() accepted SDKMAN's offline response")
	}
}

func TestSDKMANLocalState(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".sdkman")
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "bin", "sdkman-init.sh"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	linuxJava := filepath.Join(dir, "candidates", "java", "21.0.8-tem", "bin")
	macJava := filepath.Join(dir, "candidates", "java", "17.0.16-zulu", "Contents", "Home", "bin")
	for _, javaBin := range []string{linuxJava, macJava} {
		if err := os.MkdirAll(javaBin, 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(javaBin, "java"), nil, 0755); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.Symlink("21.0.8-tem", filepath.Join(dir, "candidates", "java", "current")); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "candidates", "java", "incomplete", "bin"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "candidates", "java", "not-a-jdk"), 0755); err != nil {
		t.Fatal(err)
	}

	manager := &service{dir: dir}
	versions, err := manager.Installed()
	if err != nil {
		t.Fatal(err)
	}

	want := []jdk.Installation{
		{
			Version: "21.0.8-tem",
			Current: true,
		},
		{
			Version: "17.0.16-zulu",
		},
	}

	if !reflect.DeepEqual(versions, want) {
		t.Fatalf("Installed() = %#v, want %#v", versions, want)
	}
}

func TestSDKMANSwitchPathActivatesValidatedJDKHome(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("SDKMAN correctly rejects root operations")
	}

	dir := filepath.Join(t.TempDir(), ".sdkman")
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "bin", "sdkman-init.sh"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	home := filepath.Join(t.TempDir(), "custom-jdk")
	if err := os.MkdirAll(filepath.Join(home, "bin"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(home, "bin", "java"), nil, 0755); err != nil {
		t.Fatal(err)
	}

	manager := &service{dir: dir}
	if err := manager.SwitchPath(home); err != nil {
		t.Fatal(err)
	}

	target, err := os.Readlink(filepath.Join(dir, "candidates", "java", "current"))
	if err != nil {
		t.Fatal(err)
	}

	resolvedHome, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}

	if target != resolvedHome {
		t.Fatalf("current target = %q, want %q", target, resolvedHome)
	}
}

func TestSDKMANSwitchPathRejectsNonJDK(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("SDKMAN correctly rejects root operations")
	}

	dir := filepath.Join(t.TempDir(), ".sdkman")
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "bin", "sdkman-init.sh"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	if err := (&service{dir: dir}).SwitchPath(t.TempDir()); err == nil {
		t.Fatal("SwitchPath accepted a directory without bin/java")
	}
}

func TestSDKMANSwitchPathAvoidsCurrentLinkRecursion(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("SDKMAN correctly rejects root operations")
	}

	for _, name := range []string{"current", "relative current", "alias to current"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(filepath.Join(dir, "bin", "sdkman-init.sh"), nil, 0644); err != nil {
				t.Fatal(err)
			}

			home := filepath.Join(dir, "candidates", "java", "21.0.8-tem")
			if err := os.MkdirAll(filepath.Join(home, "bin"), 0755); err != nil {
				t.Fatal(err)
			}

			if err := os.WriteFile(filepath.Join(home, "bin", "java"), nil, 0755); err != nil {
				t.Fatal(err)
			}

			current := filepath.Join(dir, "candidates", "java", "current")
			if err := os.Symlink("21.0.8-tem", current); err != nil {
				t.Fatal(err)
			}

			path := current
			switch name {
			case "relative current":
				workingDir, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}

				path, err = filepath.Rel(workingDir, current)
				if err != nil {
					t.Fatal(err)
				}

			case "alias to current":
				path = filepath.Join(dir, "jdk-alias")
				if err := os.Symlink(current, path); err != nil {
					t.Fatal(err)
				}
			}

			if err := (&service{dir: dir}).SwitchPath(path); err != nil {
				t.Fatal(err)
			}

			if _, err := os.Stat(filepath.Join(current, "bin", "java")); err != nil {
				t.Fatalf("active Java became inaccessible: %v", err)
			}

			target, err := os.Readlink(current)
			if err != nil {
				t.Fatal(err)
			}

			if name != "alias to current" && target != "21.0.8-tem" {
				t.Fatalf("current link changed to %q", target)
			}
		})
	}
}
