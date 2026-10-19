package mlcdoc

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The tests need the mlcdoc binary and are skipped without it.

func TestPreviewRendersPages(t *testing.T) {
	if Binary("") == "" {
		t.Skip("mlcdoc not installed")
	}
	base := t.TempDir()
	png := []byte("\x89PNG\r\n\x1a\n")
	// A real image from the first preview below is used for the copy test.
	res, err := Preview(context.Background(), Request{Markdown: "# Test\n\nOrt: [@#ort ____]\n\n<div class=\"page-break\"></div>\n\nSeite 2\n"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Pages) != 2 || !bytes.HasPrefix(res.Pages[0], png) {
		t.Fatalf("want 2 PNG pages, got %d", len(res.Pages))
	}
	if err := os.MkdirAll(filepath.Join(base, "img"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "img", "p.png"), res.Pages[1], 0o600); err != nil {
		t.Fatal(err)
	}
	withImage, err := Preview(context.Background(), Request{Markdown: "# Bild\n\n![x](img/p.png)\n", BaseDir: base})
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range withImage.Warnings {
		if strings.Contains(w, "p.png") {
			t.Errorf("image not found by mlcdoc: %v", withImage.Warnings)
		}
	}
}

func TestPreviewReportsMlcdocErrors(t *testing.T) {
	if Binary("") == "" {
		t.Skip("mlcdoc not installed")
	}
	_, err := Preview(context.Background(), Request{Markdown: "# A\n\nOrt: [@#ort ____\n"})
	if err == nil || !strings.Contains(err.Error(), "without closing") {
		t.Errorf("want mlcdoc's field error, got %v", err)
	}
}

func TestPreviewWithoutBinary(t *testing.T) {
	if _, err := Preview(context.Background(), Request{Markdown: "# A", Binary: "/does/not/exist"}); err != nil && Binary("/does/not/exist") == "" && err != ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestCopyImagesStaysInsideBaseDir(t *testing.T) {
	base, dir := t.TempDir(), t.TempDir()
	_ = os.WriteFile(filepath.Join(base, "a.png"), []byte("a"), 0o600)
	copyImages("![a](a.png) ![b](../secret.png) ![c](/etc/passwd) ![d](https://x/y.png)", base, dir)
	if _, err := os.Stat(filepath.Join(dir, "a.png")); err != nil {
		t.Errorf("a.png not copied")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("only a.png expected, got %v", entries)
	}
}
