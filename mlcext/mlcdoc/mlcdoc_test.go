package mlcdoc

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mlechner911/mdviewer/mlcext"
)

// Tests marked "needs mlcdoc" run against the installed binary and are
// skipped without it.

func client(t *testing.T) *Client {
	t.Helper()
	c, err := New(context.Background(), Options{})
	if errors.Is(err, ErrNotFound) {
		t.Skip("mlcdoc not installed")
	}
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestRenderPDFAndPages(t *testing.T) { // needs mlcdoc
	c := client(t)
	md := "# Test\n\nOrt: [@#ort ____]\n\n<div class=\"page-break\"></div>\n\nSeite 2\n"
	doc, err := c.RenderPDF(context.Background(), Request{Markdown: md})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(doc.PDF, []byte("%PDF-")) || doc.Pages != 2 || doc.Title != "Test" {
		t.Errorf("pdf: %d bytes, %d pages, title %q", len(doc.PDF), doc.Pages, doc.Title)
	}
	pages, err := c.RenderPages(context.Background(), Request{Markdown: md}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages.Images) != 2 || !bytes.HasPrefix(pages.Images[0], []byte("\x89PNG")) {
		t.Fatalf("want 2 PNG pages, got %d", len(pages.Images))
	}
}

func TestImagesFromBaseDir(t *testing.T) { // needs mlcdoc
	c := client(t)
	base := t.TempDir()
	pages, err := c.RenderPages(context.Background(), Request{Markdown: "# A\n"}, 30)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(base, "img"), 0o700)
	_ = os.WriteFile(filepath.Join(base, "img", "p.png"), pages.Images[0], 0o600)
	md := "# Bild\n\n![x](img/p.png)\n"
	without, _ := c.RenderPDF(context.Background(), Request{Markdown: md})
	with, err := c.RenderPDF(context.Background(), Request{Markdown: md, BaseDir: base})
	if err != nil {
		t.Fatal(err)
	}
	if len(without.Warnings) == 0 || len(with.Warnings) != 0 {
		t.Errorf("without base dir want a missing-image warning, with none: %v / %v", without.Warnings, with.Warnings)
	}
}

func TestRenderErrorCarriesTheLine(t *testing.T) { // needs mlcdoc
	c := client(t)
	_, err := c.RenderPDF(context.Background(), Request{Markdown: "# A\n\nOrt: [@#ort ____\n"})
	var re *RenderError
	if !errors.As(err, &re) || re.Line != 3 || !strings.Contains(re.Message, "without closing") {
		t.Errorf("want RenderError line 3, got %#v", err)
	}
}

// Keys must equal the list mlcdoc itself names for an unknown key.
func TestKeysMatchBinary(t *testing.T) { // needs mlcdoc
	c := client(t)
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "k.md"), []byte("---\nzzz_unknown: 1\n---\n# A\n"), 0o600)
	out, _ := exec.Command(c.Path(), "md-to-pdf", filepath.Join(dir, "k.md"), "--out", filepath.Join(dir, "k.pdf")).CombinedOutput()
	_, list, ok := strings.Cut(string(out), "expected one of ")
	if !ok {
		t.Fatalf("unexpected mlcdoc output: %s", out)
	}
	var got []string
	for _, k := range strings.Split(list, ",") {
		got = append(got, strings.Trim(strings.TrimSpace(k), "`"))
	}
	if !slices.Equal(got, Keys) {
		t.Errorf("Keys out of date for %s:\nbinary: %v\nKeys:   %v", c.Version(), got, Keys)
	}
}

func TestPrepare(t *testing.T) {
	for _, c := range []struct{ in, want, ignored string }{
		{"# A\n", "# A\n", ""},
		{"---\ntitle: X\nheader: {left: A}\n---\n# A\n", "---\ntitle: X\nheader: {left: A}\n---\n# A\n", ""},
		{"---\ntitle: X\ntags: [a, b]\naliases: [c]\n---\n# A\n", "---\ntitle: X\n---\n# A\n", "tags,aliases"},
		{"---\nevent_kinds: [DEVIATION]\n---\n\n# A\n", "# A\n", "event_kinds"},
		{"---\ntitle: [kaputt\n---\n# A\n", "---\ntitle: [kaputt\n---\n# A\n", ""},
	} {
		got, ignored := Prepare(c.in)
		if got != c.want || strings.Join(ignored, ",") != c.ignored {
			t.Errorf("Prepare(%q) = %q %v, want %q %s", c.in, got, ignored, c.want, c.ignored)
		}
	}
}

// Every built-in template renders in mlcdoc without warnings.
func TestTemplatesRenderWithoutWarnings(t *testing.T) { // needs mlcdoc
	c := client(t)
	prefix := map[string]string{"sop": "SOP", "va": "VA", "form": "FB", "info": "INFO"}
	for _, tpl := range mlcext.LoadTemplates() {
		r := tpl.Render(map[string]any{"titel": "Test", "bereich": "QM", "kennung": prefix[tpl.ID] + "-QM-001", "version": "1.0", "anlass": []any{"DEVIATION"}})
		if len(r.FieldErrors) > 0 {
			t.Fatalf("%s: %v", tpl.ID, r.FieldErrors)
		}
		doc, err := c.RenderPDF(context.Background(), Request{Markdown: r.Content})
		if err != nil {
			t.Errorf("%s: %v", tpl.ID, err)
			continue
		}
		if len(doc.Warnings) > 0 || doc.Pages == 0 {
			t.Errorf("%s: %d pages, warnings %v", tpl.ID, doc.Pages, doc.Warnings)
		}
	}
}

func TestVersionCheck(t *testing.T) {
	if compare(semver("mlcdoc 0.15.9 (x)"), semver("mlcdoc "+MinVersion)) >= 0 {
		t.Errorf("0.15.9 must be older than %s", MinVersion)
	}
	if compare(semver("mlcdoc 0.16.0 (mlcmd2pdf 3.26.0)"), semver("mlcdoc "+MinVersion)) != 0 {
		t.Errorf("0.16.0 must equal MinVersion")
	}
	if semver("irgendwas") != nil {
		t.Errorf("no version must give nil")
	}
	if _, err := New(context.Background(), Options{Binary: "/does/not/exist"}); err != nil && Binary("/does/not/exist") == "" && !errors.Is(err, ErrNotFound) {
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
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("only a.png expected, got %v", entries)
	}
}

func TestParseError(t *testing.T) {
	err := parseError("mlcdoc: step 1 (md_to_pdf): form field in line 7: field name `x` is used twice\n", nil)
	var re *RenderError
	if !errors.As(err, &re) || re.Line != 7 || re.Message != "field name `x` is used twice" {
		t.Errorf("got %#v", err)
	}
}
