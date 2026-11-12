package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"marksafe/internal/config"
)

func folderCheckApp(t *testing.T) (*App, string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("AppData", t.TempDir())
	cfg, err := config.NewConfigManager()
	if err != nil {
		t.Fatal(err)
	}
	allowed := t.TempDir()
	if err := cfg.AddPath(allowed); err != nil {
		t.Fatal(err)
	}
	return &App{config: cfg}, allowed
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFolderCheckNeedsTrust(t *testing.T) {
	a, allowed := folderCheckApp(t)
	writeFile(t, filepath.Join(allowed, ".marksafe.yaml"), "version: 1\ncheck:\n  command: [\"go\", \"version\"]\n")
	doc := filepath.Join(allowed, "sub", "a.md")
	writeFile(t, doc, "# a")

	info, err := a.GetFolderCheck(doc)
	if err != nil || info.Spec == nil || info.Trusted {
		t.Fatalf("info = %+v, err = %v", info, err)
	}
	if _, err := a.RunFolderCheck(doc); !errors.Is(err, errFolderCheckNotTrusted) {
		t.Fatalf("untrusted check must not run: %v", err)
	}
	if err := a.TrustFolderCheck(doc, "stale-hash"); err == nil {
		t.Fatal("a stale hash must not be trusted")
	}
	if err := a.TrustFolderCheck(doc, info.Spec.Hash); err != nil {
		t.Fatal(err)
	}
	if info, _ = a.GetFolderCheck(doc); !info.Trusted {
		t.Fatal("check should be trusted now")
	}

	// Changing the command withdraws the permission.
	writeFile(t, filepath.Join(allowed, ".marksafe.yaml"), "version: 1\ncheck:\n  command: [\"go\", \"env\"]\n")
	if info, _ = a.GetFolderCheck(doc); info.Trusted {
		t.Fatal("a changed command must ask again")
	}
}

func TestFolderCheckOnlyInAllowedFolders(t *testing.T) {
	a, _ := folderCheckApp(t)
	other := t.TempDir()
	writeFile(t, filepath.Join(other, ".marksafe.yaml"), "version: 1\ncheck:\n  command: [\"go\"]\n")
	doc := filepath.Join(other, "a.md")
	writeFile(t, doc, "# a")
	info, err := a.GetFolderCheck(doc)
	if err != nil || info.Spec != nil {
		t.Fatalf("a folder outside the whitelist must be ignored: %+v %v", info, err)
	}
}

func TestFolderCheckBrokenFileIsReported(t *testing.T) {
	a, allowed := folderCheckApp(t)
	writeFile(t, filepath.Join(allowed, ".marksafe.yaml"), "version: 7\n")
	doc := filepath.Join(allowed, "a.md")
	writeFile(t, doc, "# a")
	if _, err := a.GetFolderCheck(doc); err == nil {
		t.Fatal("an unknown version must be reported")
	}
}
