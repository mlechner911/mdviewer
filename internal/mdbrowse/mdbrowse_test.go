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

	title, html, err := root.RenderDoc("index.md")
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
