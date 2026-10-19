package mlcext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func codes(ds []Diagnostic, sev Severity) []string {
	var out []string
	for _, d := range ds {
		if d.Severity == sev {
			out = append(out, d.Code)
		}
	}
	return out
}

func has(ds []Diagnostic, code string) *Diagnostic {
	for i := range ds {
		if ds[i].Code == code {
			return &ds[i]
		}
	}
	return nil
}

const validSOP = `---
title: Kalibrierung
event_kinds: [CALIBRATION_DUE, LOW_BATTERY]
---
# Standardarbeitsanweisung (SOP)

**Dokumenten-ID:** SOP-KAL-001 | **Version:** 1.1

---

## 1. Ablauf

Ort: [@#ort ________] <!--* z. B. Kühlraum A -->
Geprüft: [@#geprueft]
<!--! So arbeiten Sie mit dem Blatt. -->
`

func TestValidDocumentHasNoErrors(t *testing.T) {
	ds := Check(validSOP, Options{RequireHeader: true})
	if errs := codes(ds, Error); len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", ds)
	}
	d := has(ds, "field-prefill")
	if d == nil || d.Line != 13 || d.Col != 6 {
		t.Errorf("prefill info at 13:6 expected, got %v", ds)
	}
}

func TestFrontMatterRules(t *testing.T) {
	md := "---\ntitle: X\nvalid_from: 2026-10-01\nowner: QMB\nevent_kinds: [NO_DATA, FOO]\n---\n# X\n"
	ds := Check(md, Options{})
	for code, line := range map[string]int{
		"frontmatter-server-owned": 3,
		"frontmatter-unknown-key":  4,
		"frontmatter-event-kind":   5,
	} {
		d := has(ds, code)
		if d == nil || d.Line != line || d.Severity != Error {
			t.Errorf("%s at line %d expected, got %v", code, line, ds)
		}
	}
	if d := has(Check("---\nevent_kinds: NO_DATA\n---\n", Options{}), "frontmatter-event-kinds"); d == nil {
		t.Errorf("event_kinds as scalar must be an error")
	}
	if d := has(Check("---\ntitle: [unclosed\n---\n", Options{}), "frontmatter-yaml"); d == nil {
		t.Errorf("invalid YAML must be an error")
	}
	if d := has(Check("---\ntitle: X\n# kein Ende\n", Options{}), "frontmatter-unclosed"); d == nil || d.Severity != Warning {
		t.Errorf("unclosed front matter must be a warning")
	}
	if ds := Check("---\nevent_kinds: []\n---\n", Options{}); len(ds) != 0 {
		t.Errorf("empty event_kinds is valid, got %v", ds)
	}
}

func TestHeaderRules(t *testing.T) {
	ds := Check("# Formblatt\n\nText\n", Options{RequireHeader: true})
	for _, code := range []string{"header-id-missing", "header-version-missing"} {
		if d := has(ds, code); d == nil || d.Severity != Warning {
			t.Errorf("%s must be a warning (the upload form can supply it), got %v", code, ds)
		}
	}
	ds = Check("# X\n\n**Dokumenten-ID:** XY-1 | **Version:** 1.0\n**Gültig ab:** [Datum]\n\n---\n", Options{RequireHeader: true})
	if d := has(ds, "header-id-prefix"); d == nil || d.Severity != Warning || d.Line != 3 {
		t.Errorf("unknown prefix warning on line 3 expected, got %v", ds)
	}
	if d := has(ds, "header-valid-from"); d == nil || d.Line != 4 {
		t.Errorf("Gültig ab (also a placeholder) is an error on line 4, got %v", ds)
	}
	// Without RequireHeader nothing about the header is reported.
	if ds := Check("# X\n\n**Gültig ab:** 1.1.2026\n", Options{}); len(ds) != 0 {
		t.Errorf("plain Markdown: %v", ds)
	}
}

func TestHeaderParsing(t *testing.T) {
	h := ParseHeader("---\ntitle: T\n---\n# T\n\n**Dokumenten-ID:** FB-QM-002 | **Version:** 2.1\\\n\n---\n**Version:** 9\n")
	if h.RefNo != "FB-QM-002" || h.Version != "2.1" || h.Kind != "form" {
		t.Errorf("header: %+v", h)
	}
	if h.refLine != 6 {
		t.Errorf("refLine = %d, want 6", h.refLine)
	}
}

func TestFieldErrorsBecomeDiagnostics(t *testing.T) {
	ds := Check("# A\n\nOrt: [@#ort ____\n", Options{})
	if d := has(ds, "field-unclosed"); d == nil || d.Line != 3 || d.Col != 6 {
		t.Errorf("unclosed field at 3:6 expected, got %v", ds)
	}
}

func TestUnclosedHints(t *testing.T) {
	ds := Check("Text\n<!--! Tipp ohne Ende\n", Options{})
	if d := has(ds, "hint-unclosed"); d == nil || d.Line != 2 {
		t.Errorf("unclosed hint expected, got %v", ds)
	}
	if ds := Check("```\n<!--! im Code\n```\n", Options{}); len(ds) != 0 {
		t.Errorf("code must be ignored, got %v", ds)
	}
}

// The Sensor Monitor's shipped templates must all be valid. They live in a
// private repository, so the test reads them from MLCEXT_SEEDS (a directory
// of .md files, e.g. mlcsensormonitor/internal/sop/seeds) and is skipped
// without it.
func TestSensorMonitorSeedsAreValid(t *testing.T) {
	dir := os.Getenv("MLCEXT_SEEDS")
	if dir == "" {
		t.Skip("MLCEXT_SEEDS not set")
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	if len(files) == 0 {
		t.Fatalf("no .md files in %s", dir)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range Check(string(b), Options{RequireHeader: true}) {
			if d.Severity == Error {
				t.Errorf("%s:%d:%d %s: %s", filepath.Base(f), d.Line, d.Col, d.Code, d.Message)
			}
		}
		if testing.Verbose() {
			fs, _ := ExtractFields(string(b))
			t.Logf("%s: %d fields, header %+v", filepath.Base(f), len(fs), strings.TrimSpace(ParseHeader(string(b)).RefNo))
		}
	}
}

func TestDetect(t *testing.T) {
	for md, want := range map[string]bool{
		"# README\n\n**Version:** 2\n":    false,
		"---\ntags: [a]\n---\n# Doc\n":    false,
		"- [ ] task\n\n`[@code]`\n":       false,
		"**Dokumenten-ID:** SOP-QM-001\n": true,
		"---\nevent_kinds: []\n---\n":     true,
		"Ort: [@#ort ____]\n":             true,
		"Ort: [@#ort ____\n":              true,
	} {
		if got := Detect(md); got != want {
			t.Errorf("Detect(%q) = %v, want %v", md, got, want)
		}
	}
}
