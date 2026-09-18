package mdbrowse

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeGluedHeadings(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Backups## 1. Titel", "Backups\n## 1. Titel"},
		{"text ## Heading", "text\n## Heading"},
		{"---## 2. Titel", "---\n## 2. Titel"},
		{"I love C# rocks and Issue #123.", "I love C# rocks and Issue #123."},
		{"Use `a##b` here.", "Use `a##b` here."},
		{"***bold*** stays.", "***bold*** stays."},
		{"# Fine\n\nNormal *list* item.", "# Fine\n\nNormal *list* item."},
	}
	for _, tc := range cases {
		if got := normalizePastedMarkdown(tc.in); got != tc.want {
			t.Errorf("normalize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeClosesFence(t *testing.T) {
	out := normalizePastedMarkdown("# T\n\n```text\ncode ## kept\n")
	if !strings.Contains(out, "code ## kept") {
		t.Errorf("fenced content altered: %q", out)
	}
	if strings.Count(out, "```")%2 != 0 {
		t.Errorf("fence left unclosed: %q", out)
	}
}

func TestFuzzyQuality(t *testing.T) {
	if q := fuzzyQuality("validierung", "validirung"); q < 0.5 {
		t.Errorf("typo should match, got %v", q)
	}
	if q := fuzzyQuality("backup", "zzzqqq"); q != 0 {
		t.Errorf("garbage should not match, got %v", q)
	}
	contig := fuzzyQuality("mqtt", "mqtt")
	gapped := fuzzyQuality("m_q_t_t", "mqtt")
	if !(contig > gapped && gapped > 0) {
		t.Errorf("contiguous should outrank gapped: %v vs %v", contig, gapped)
	}
}

func TestTruncateTitle(t *testing.T) {
	if got := truncateTitle("Short"); got != "Short" {
		t.Errorf("short title altered: %q", got)
	}
	if got := truncateTitle("Backups## 1. Rest"); got != "Backups" {
		t.Errorf("glued subheading not cut: %q", got)
	}
	long := strings.Repeat("a", 200)
	if got := truncateTitle(long); len([]rune(got)) != maxTitleRunes+1 {
		t.Errorf("long title not capped: %d runes", len([]rune(got)))
	}
}

func TestRenderMathMermaidAnchors(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "index.md", "# Hello World\n\nInline $E=mc^2$ and:\n\n$$\n\\int_0^1 x dx\n$$\n\n```mermaid\ngraph TD; A-->B;\n```\n")
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, html, err := root.RenderDoc("index.md", "github-dark")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `id="hello-world"`) {
		t.Errorf("heading anchor missing: %s", html)
	}
	if !strings.Contains(html, "language-mermaid") {
		t.Errorf("mermaid fence not preserved: %s", html)
	}
	if !strings.Contains(html, "math") {
		t.Errorf("math markup missing: %s", html)
	}
}

// U+0130 lowercases to two runes (three bytes): byte indices from a
// case-folded copy do not line up with the original. This used to slice
// mid-rune (mojibake) or panic outright, killing the whole /api/search
// handler — every query after that showed "Keine Treffer".
func TestHighlightMarkUnicodeExpansion(t *testing.T) {
	text := escapeHTML("xİroll and more text here to fill the snippet window nicely")
	out := highlightMark(text, escapeHTML("roll"))
	if !strings.Contains(out, "<mark>roll</mark>") {
		t.Errorf("mark missing or misaligned: %q", out)
	}
	if strings.Contains(out, "\ufffd") {
		t.Errorf("replacement char leaked: %q", out)
	}
}

func TestFormatTitleUmlaut(t *testing.T) {
	if got := FormatTitle("Übersicht"); got != "Übersicht" {
		t.Errorf("umlaut initial corrupted: %q", got)
	}
}

// A match glued to a long spaceless run once overshot the window trim
// (slice bounds [5416:5295]) and panicked the whole /api/search handler:
// every later query died, the UI showed "Keine Treffer" for everything.
func TestSnippetWindowNeverInverts(t *testing.T) {
	body := "prefix text here roll" + strings.Repeat("x", 200) + " tail end words here"
	snip := snippetExact(body+" ", "roll")
	if !strings.Contains(snip, "<mark>roll</mark>") {
		t.Errorf("mark missing: %q", snip)
	}
	dir := t.TempDir()
	writeFixture(t, dir, "long.md", "# Lang\n\nText "+strings.Repeat("y", 300)+" roll Ende.\n")
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	hits, err := root.Search("roll")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("want 1 hit, got %+v", hits)
	}
}

func TestRenderDirIndexAndVirtual(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "withidx/index.md", "# Hallo Index\n\nIntro.\n")
	writeFixture(t, dir, "withidx/a.md", "# A\n")
	writeFixture(t, dir, "plain/b.md", "# B\n\nSiehe [a](../withidx/a.md).\n")
	writeFixture(t, dir, "plain/sub/c.md", "# C\n")
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Directory with index.md renders the index page.
	title, html, err := root.RenderDir("withidx", "github-dark")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Hallo Index" || !strings.Contains(html, "Intro.") {
		t.Errorf("index not rendered: %q %q", title, html)
	}

	// Directory without index.md renders a virtual listing.
	title, html, err = root.RenderDir("plain", "github-dark")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Plain" {
		t.Errorf("title = %q", title)
	}
	for _, want := range []string{`data-md="plain/b.md"`, `data-md="plain/sub"`, ">B<", ">Sub<"} {
		if !strings.Contains(html, want) {
			t.Errorf("listing lacks %q: %s", want, html)
		}
	}

	// Missing and escaping directories fail.
	if _, _, err := root.RenderDir("nope", "github-dark"); err == nil {
		t.Error("missing dir accepted")
	}
	if _, _, err := root.RenderDir("../..", "github-dark"); err == nil {
		t.Error("escape accepted")
	}
	// Files are not directories.
	if _, _, err := root.RenderDir("plain/b.md", "github-dark"); err == nil {
		t.Error("file accepted as dir")
	}
}

func TestExternalLinksMarkedAndTabbed(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "index.md", "# S\n\n[web](https://example.com/x) and [mail](mailto:a@b.c) and [local](good.md).\n")
	writeFixture(t, dir, "good.md", "# G\n")
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, html, err := root.RenderDoc("index.md", "github-dark")
	if err != nil {
		t.Fatal(err)
	}
	// External: marked, new window, opener severed.
	for _, want := range []string{`class="external-link"`, `target="_blank"`, `noopener`, `noreferrer`, `href="https://example.com/x"`} {
		if !strings.Contains(html, want) {
			t.Errorf("external link lacks %q: %s", want, html)
		}
	}
	// mailto keeps native behavior (no new window, no confirm hook).
	if strings.Contains(html, "mailto") && strings.Contains(html, "external-link mailto") {
		t.Errorf("mailto misclassified: %s", html)
	}
	// Internal link still navigates in-app.
	if !strings.Contains(html, `data-md="good.md"`) {
		t.Errorf("internal link lost data-md: %s", html)
	}
}

func TestTOCEntryJSONKeys(t *testing.T) {
	raw, _ := json.Marshal(TOCEntry{Path: "a/b.md", Title: "B", IsDir: false})
	s := string(raw)
	for _, k := range []string{`"path"`, `"title"`, `"children"`, `"isDir"`} {
		if !strings.Contains(s, k) {
			t.Errorf("missing lowercase key %s in %s", k, s)
		}
	}
}

func writeFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRootResolveRejectsEscape(t *testing.T) {
	root, err := NewRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := root.Resolve("../outside.md"); err == nil {
		t.Error("escape accepted, want rejection")
	}
	if _, err := root.Resolve("/etc/passwd"); err == nil {
		t.Error("absolute escape accepted, want rejection")
	}
}

func TestRewriteAndSearch(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "index.md", "# Start\n\nSee [good](good.md), [gone](missing.md) and [out](../x.md).\n")
	writeFixture(t, dir, "good.md", "# Gute Seite\n\nEin Text über Snapshot und Deltas.\n")
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}

	title, html, err := root.RenderDoc("index.md", "github-dark")
	if err != nil {
		t.Fatal(err)
	}
	if title != "Start" {
		t.Errorf("title = %q", title)
	}
	if !strings.Contains(html, `data-md="good.md"`) {
		t.Errorf("valid link not marked: %s", html)
	}
	if strings.Count(html, "broken-link") != 2 {
		t.Errorf("want 2 broken links, got: %s", html)
	}

	hits, err := root.Search("snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Path != "good.md" {
		t.Fatalf("search hits = %+v", hits)
	}
	if !strings.Contains(hits[0].Snippet, "<mark>") {
		t.Errorf("snippet lacks mark: %q", hits[0].Snippet)
	}

	if hits, _ := root.Search(""); len(hits) != 0 {
		t.Errorf("empty query should yield nothing, got %d", len(hits))
	}
}
