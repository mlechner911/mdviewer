package markdown

import (
	"os"
	"strings"
	"testing"
)

// Cases from mlcmd2pdf (src/markdown.rs, slugs_headings_like_github):
// preview anchors must equal the PDF anchors.
func TestHeadingSlugMatchesMlcdoc(t *testing.T) {
	tests := map[string]string{
		"EV-2026-0666 · Kritisch": "ev-2026-0666--kritisch",
		"Übersicht der Messwerte": "übersicht-der-messwerte",
		"1. Ziel & Zweck (v2)":    "1-ziel--zweck-v2",
		"snake_case ✅":            "snake_case-",
	}
	for in, want := range tests {
		if got := HeadingSlug(in); got != want {
			t.Errorf("HeadingSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDialectAnchorsAndPageBreak(t *testing.T) {
	src := "# Bericht\n\n## Übersicht\n\n## Übersicht\n\n## Explizit {#fest}\n\n" +
		"<div class=\"page-break\"></div>\n\nText <a id=\"marke\"></a> weiter.\n"
	html, err := NewRenderer().Render(src, "github")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<h1 id="bericht">`,
		`<h2 id="übersicht">`,
		`<h2 id="übersicht-1">`,
		`<h2 id="fest">Explizit</h2>`,
		`<div class="page-break"></div>`,
		`<a id="marke"></a>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s in:\n%s", want, html)
		}
	}
}

func TestDialectDropsOtherRawHTML(t *testing.T) {
	src := "<script>alert(1)</script>\n\n<div class=\"page-break\" onclick=\"x()\"></div>\n\n" +
		"Text <span onclick=\"x()\">s</span> <a id=\"x\" href=\"javascript:x()\">l</a>\n"
	html, err := NewRenderer().Render(src, "github")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"<script", "onclick", "javascript:", "<span", `id="x"`} {
		if strings.Contains(html, bad) {
			t.Errorf("unexpected %q in:\n%s", bad, html)
		}
	}
}

// The shared fixture covers the dialect end to end; the same file is the
// visual reference for document.css.
func TestDialectFixture(t *testing.T) {
	src, err := os.ReadFile("testdata/mlc-markdown.md")
	if err != nil {
		t.Fatal(err)
	}
	html, err := NewRenderer().Render(string(src), "github")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`frontmatter-container`,
		`markdown-alert markdown-alert-note`,
		`markdown-alert markdown-alert-caution`,
		`<h2 id="boxen">`,
		`href="#boxen"`,
		`<div class="page-break"></div>`,
		`<table>`,
		`type="checkbox"`,
		`class="chroma"`,
		`language-mermaid`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("fixture: missing %s", want)
		}
	}
}
