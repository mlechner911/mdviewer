package main

import (
	"os"
	"path/filepath"
	"testing"
)

// The way back from the browse window opens only Markdown files inside
// the browsed folder.
func TestOpenBrowseInEditor(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "sop.mlcmd"), []byte("# SOP\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bild.png"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "kapitel"), 0o755); err != nil {
		t.Fatal(err)
	}
	a := &App{}
	if err := a.OpenBrowseInEditor(root, "sop.mlcmd"); err != nil {
		t.Errorf("document: %v", err)
	}
	for _, p := range []string{"../etc/passwd", "bild.png", "kapitel", "fehlt.md"} {
		if err := a.OpenBrowseInEditor(root, p); err == nil {
			t.Errorf("%s: no error", p)
		}
	}
}
