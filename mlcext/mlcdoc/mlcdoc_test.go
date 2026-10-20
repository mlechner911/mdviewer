package mlcdoc

import (
	"github.com/mlechner911/mdviewer/mlcext"

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

// Every built-in template renders in mlcdoc without warnings.
func TestTemplatesRenderWithoutWarnings(t *testing.T) {
	if Binary("") == "" {
		t.Skip("mlcdoc not installed")
	}
	prefix := map[string]string{"sop": "SOP", "va": "VA", "form": "FB", "info": "INFO"}
	for _, tpl := range mlcext.LoadTemplates() {
		r := tpl.Render(map[string]any{"titel": "Test", "bereich": "QM", "kennung": prefix[tpl.ID] + "-QM-001", "version": "1.0", "anlass": []any{"DEVIATION"}})
		if len(r.FieldErrors) > 0 {
			t.Fatalf("%s: %v", tpl.ID, r.FieldErrors)
		}
		res, err := Preview(context.Background(), Request{Markdown: mlcext.StoredText(r.Content)})
		if err != nil {
			t.Errorf("%s: %v", tpl.ID, err)
			continue
		}
		if len(res.Warnings) > 0 || len(res.Pages) == 0 {
			t.Errorf("%s: %d pages, warnings %v", tpl.ID, len(res.Pages), res.Warnings)
		}
	}
}

// mlcdoc rejects the upload front matter (event_kinds); StoredText removes
// it like the Sensor Monitor does on upload.
func TestUploadFrontMatterNeedsStoredText(t *testing.T) {
	if Binary("") == "" {
		t.Skip("mlcdoc not installed")
	}
	md := "---\ntitle: X\nevent_kinds: [DEVIATION]\n---\n# A\n"
	if _, err := Preview(context.Background(), Request{Markdown: md}); err == nil || !strings.Contains(err.Error(), "event_kinds") {
		t.Errorf("mlcdoc accepted event_kinds (behavior changed?): %v", err)
	}
	if _, err := Preview(context.Background(), Request{Markdown: mlcext.StoredText(md)}); err != nil {
		t.Errorf("stored text: %v", err)
	}
}
