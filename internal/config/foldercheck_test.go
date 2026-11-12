package config

import (
	"path/filepath"
	"testing"
)

func TestTrustFolderCheck(t *testing.T) {
	m := &ConfigManager{path: filepath.Join(t.TempDir(), "config.json")}
	if m.IsFolderCheckTrusted("/a", "h1") {
		t.Fatal("nothing is trusted at first")
	}
	if err := m.TrustFolderCheck("/a", "h1"); err != nil {
		t.Fatal(err)
	}
	if !m.IsFolderCheckTrusted("/a", "h1") || m.IsFolderCheckTrusted("/a", "h2") || m.IsFolderCheckTrusted("/b", "h1") {
		t.Error("trust must match folder and hash")
	}
	if m.IsFolderCheckTrusted("/x", "") {
		t.Error("an empty hash is never trusted")
	}
	if err := m.Load(); err != nil || !m.IsFolderCheckTrusted("/a", "h1") {
		t.Errorf("trust must survive a reload: %v", err)
	}
}

func TestWhitelistedRootOf(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "home", "u", "docs")
	deep := filepath.Join(root, "repo")
	m := &ConfigManager{config: AppConfig{WhitelistedPaths: []string{root, deep}}}
	if got := m.WhitelistedRootOf(filepath.Join(deep, "x", "a.md")); got != deep {
		t.Errorf("got %q, want deepest %q", got, deep)
	}
	if got := m.WhitelistedRootOf(root + "-private/a.md"); got != "" {
		t.Errorf("got %q, want none", got)
	}
}
