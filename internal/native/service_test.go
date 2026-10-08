//go:build windows

package native

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	appcfg "github.com/ystyle/jvms/internal/config"
	"github.com/ystyle/jvms/internal/jdk"
)

func nativeFixture(t *testing.T) *service {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())
	root := t.TempDir()
	cfg := appcfg.NewConfig()
	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}
	cfg.Store = filepath.Join(root, "store")
	cfg.Download = filepath.Join(root, "download")
	cfg.JavaHome = filepath.Join(root, "current")
	m := NewService(cfg).(*service)
	m.isAdmin = func() bool { return true }
	m.run = func(string, ...string) error { return nil }
	return m
}

func addNativeJDK(t *testing.T, m *service, v string) string {
	t.Helper()
	home := filepath.Join(m.config.Store, v)
	if err := os.MkdirAll(filepath.Join(home, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "bin", "javac.exe"), []byte("fixture"), 0644); err != nil {
		t.Fatal(err)
	}
	return home
}
func TestNativePrivilegeFailurePreventsCommands(t *testing.T) {
	m := nativeFixture(t)
	m.isAdmin = func() bool { return false }
	m.run = func(string, ...string) error { t.Fatal("command executed without admin"); return nil }
	if err := m.Ensure(); err == nil {
		t.Fatal("ensure accepted non-admin")
	}
	if err := m.Switch("21"); err == nil {
		t.Fatal("switch accepted non-admin")
	}
	if err := m.SwitchPath(t.TempDir()); err == nil {
		t.Fatal("path switch accepted non-admin")
	}
}

func TestNativeSwitchPathOutsideStore(t *testing.T) {
	m := nativeFixture(t)
	home := filepath.Join(t.TempDir(), "external JDK")
	if err := os.MkdirAll(filepath.Join(home, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "bin", "javac.exe"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	old := addNativeJDK(t, m, "17")
	if err := os.Symlink(old, m.config.JavaHome); err != nil {
		t.Skipf("host cannot create symlinks: %v", err)
	}
	m.config.CurrentJDKVersion = "17"
	if err := m.SwitchPath(home); err != nil {
		t.Fatal(err)
	}
	target, err := os.Readlink(m.config.JavaHome)
	if err != nil || target != home {
		t.Fatalf("link = %q, %v; want %q", target, err, home)
	}
	if m.config.CurrentJDKVersion != "" {
		t.Fatal("external JDK retained managed version")
	}
}

func TestNativeSwitchReplacesDanglingLink(t *testing.T) {
	for _, asPath := range []bool{false, true} {
		t.Run(fmt.Sprintf("asPath=%t", asPath), func(t *testing.T) {
			m := nativeFixture(t)
			home := addNativeJDK(t, m, "21")
			old := addNativeJDK(t, m, "17")
			if err := os.Symlink(old, m.config.JavaHome); err != nil {
				t.Skipf("host cannot create symlinks: %v", err)
			}
			if err := os.RemoveAll(old); err != nil {
				t.Fatal(err)
			}
			var err error
			if asPath {
				err = m.SwitchPath(home)
			} else {
				err = m.Switch("21")
			}
			if err != nil {
				t.Fatal(err)
			}
			target, err := os.Readlink(m.config.JavaHome)
			if err != nil || target != home {
				t.Fatalf("link = %q, %v; want %q", target, err, home)
			}
		})
	}
}

func TestNativeSwitchPathRejectsInvalidHome(t *testing.T) {
	m := nativeFixture(t)
	m.config.CurrentJDKVersion = "17"
	m.run = func(string, ...string) error { t.Fatal("command executed for invalid JDK"); return nil }
	if err := m.SwitchPath(t.TempDir()); err == nil {
		t.Fatal("accepted directory without javac.exe")
	}
	if m.config.CurrentJDKVersion != "17" {
		t.Fatal("invalid path changed current version")
	}
}

func TestNativeSwitchPreservesNonSymlinkHome(t *testing.T) {
	m := nativeFixture(t)
	addNativeJDK(t, m, "21")
	if err := os.Mkdir(m.config.JavaHome, 0755); err != nil {
		t.Fatal(err)
	}
	if err := m.Switch("21"); err == nil {
		t.Fatal("replaced a real directory")
	}
	info, err := os.Lstat(m.config.JavaHome)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("original directory not preserved: %v, %v", info, err)
	}
}
func TestNativeSwitchAndRemove(t *testing.T) {
	m := nativeFixture(t)
	home := addNativeJDK(t, m, "21")
	// Verify the host allows symlinks before exercising the operation.
	probe := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(home, probe); err != nil {
		t.Skipf("host cannot create symlinks: %v", err)
	}
	if err := m.Switch("21"); err != nil {
		t.Fatal(err)
	}
	target, err := os.Readlink(m.config.JavaHome)
	if err != nil {
		t.Fatal(err)
	}
	if target != home || m.config.CurrentJDKVersion != "21" {
		t.Fatalf("switch state = %s, %s", target, m.config.CurrentJDKVersion)
	}
	installed, err := m.Installed()
	if err != nil {
		t.Fatal(err)
	}
	if len(installed) != 1 || !installed[0].Current {
		t.Fatalf("installed: %v", installed)
	}
	if err := m.Remove("21"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(home); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("installation still exists: %v", err)
	}
	if _, err := os.Lstat(m.config.JavaHome); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("link still exists: %v", err)
	}
	if m.config.CurrentJDKVersion != "" {
		t.Fatal("removed version remains current")
	}
}
func TestNativeSwitchFailureKeepsCurrentVersion(t *testing.T) {
	m := nativeFixture(t)
	addNativeJDK(t, m, "21")
	m.config.CurrentJDKVersion = "17"
	m.run = func(string, ...string) error { return errors.New("setx failed") }
	if err := m.Switch("21"); err == nil {
		t.Fatal("command failure swallowed")
	}
	if m.config.CurrentJDKVersion != "17" {
		t.Fatal("failed switch changed version")
	}
	if err := m.Switch("missing"); err == nil {
		t.Fatal("missing version accepted")
	}
}
func TestNativeInstallExtractsCatalogArchive(t *testing.T) {
	m := nativeFixture(t)
	var archive bytes.Buffer
	w := zip.NewWriter(&archive)
	header := &zip.FileHeader{Name: "jdk-21/bin/javac.exe", Method: zip.Deflate}
	header.SetMode(0755)
	f, err := w.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("compiler")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++; w.Write(archive.Bytes()) }))
	defer server.Close()
	if err := jdk.CacheVersions(m.config, "native-windows", []jdk.Version{{Version: "21", Url: server.URL + "/jdk.zip"}}); err != nil {
		t.Fatal(err)
	}
	if err := m.Install("21"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(m.config.Store, "21", "bin", "javac.exe"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "compiler" {
		t.Fatalf("unexpected archive contents: %s", data)
	}
	if err := m.Install("21"); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("downloaded %d times", requests)
	}
	if _, err := os.Stat(filepath.Join(m.config.Download, "21_temp")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary directory not cleaned: %v", err)
	}
}
func TestNativeInstallReturnsDownloadFailure(t *testing.T) {
	m := nativeFixture(t)
	// An empty URL fails locally and does not contact the network.
	if err := jdk.CacheVersions(m.config, "native-windows", []jdk.Version{{Version: "21"}}); err != nil {
		t.Fatal(err)
	}
	if err := m.Install("21"); err == nil {
		t.Fatal("download failure swallowed")
	}
}

func TestNativeManualJDKActions(t *testing.T) {
	m := nativeFixture(t)
	version := "jdk 17"
	home := addNativeJDK(t, m, version)
	probe := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(home, probe); err != nil {
		t.Skipf("host cannot create symlinks: %v", err)
	}
	installed, err := m.Installed()
	if err != nil || len(installed) != 1 || installed[0].Version != version {
		t.Fatalf("installed = %v, %v; want %q", installed, err, version)
	}

	if err := m.Install(version); err != nil {
		t.Fatal(err)
	}
	if err := m.Switch(version); err != nil {
		t.Fatal(err)
	}
	target, err := os.Readlink(m.config.JavaHome)
	if err != nil || target != home || m.config.CurrentJDKVersion != version {
		t.Fatalf("switch target = %q, %v; current = %q", target, err, m.config.CurrentJDKVersion)
	}
	if err := m.Remove(version); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(home); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("manual JDK was not removed: %v", err)
	}
	if m.config.CurrentJDKVersion != "" {
		t.Fatal("removed version remains current")
	}
}

func TestNativeActionsRejectStoreTraversal(t *testing.T) {
	m := nativeFixture(t)
	outside := addNativeJDK(t, m, "../outside")
	m.run = func(string, ...string) error {
		t.Fatal("command executed for a JDK outside the store")
		return nil
	}
	for _, version := range []string{"../outside", `..\outside`, outside} {
		if err := m.Switch(version); err == nil {
			t.Errorf("switch accepted %q", version)
		}
		if err := m.Remove(version); err == nil {
			t.Errorf("remove accepted %q", version)
		}
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("JDK outside the store was changed: %v", err)
	}
}
