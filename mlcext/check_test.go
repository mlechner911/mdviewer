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

func TestAnalyzePositions(t *testing.T) {
	md := "# Ä\n\nOrt: [@#ort ___] <!--* Raum -->\n<!--! Tipp -->\n[@#vorfall ___\n___]\n"
	a := Analyze(md)
	if !a.Detected || len(a.Fields) != 2 || len(a.Hints) != 2 {
		t.Fatalf("analysis: %+v", a)
	}
	f := a.Fields[0]
	if f.Line != 3 || f.Col != 6 || f.EndLine != 3 || f.EndCol != 17 || !f.Prefill || f.Kind != "text" {
		t.Errorf("field 0: %+v", f)
	}
	if m := a.Fields[1]; m.Line != 5 || m.EndLine != 6 || m.EndCol != 5 || m.Kind != "multiline" || !m.Prefill {
		t.Errorf("field 1: %+v", m)
	}
	if h := a.Hints[0]; h.Kind != "tooltip" || h.Line != 3 || h.Col != 18 {
		t.Errorf("hint 0: %+v", h)
	}
	if h := a.Hints[1]; h.Kind != "fill-hint" || h.Line != 4 || h.Col != 1 || h.EndCol != 15 {
		t.Errorf("hint 1: %+v", h)
	}
	if plain := Analyze("# README\n"); plain.Detected || len(plain.Lists.EventKinds) == 0 {
		t.Errorf("plain: %+v", plain)
	}
}

// Every prefill marker must itself be a valid field of the right name.
func TestPrefillMarkersAreValidFields(t *testing.T) {
	for _, p := range Prefill {
		fs, errs := ExtractFields(p.Marker)
		if len(errs) > 0 || len(fs) != 1 || fs[0].Name != p.Name {
			t.Errorf("%s: %q gives %+v %v", p.Name, p.Marker, fs, errs)
		}
	}
}

func TestStoredText(t *testing.T) {
	upload := "---\ntitle: X\nevent_kinds: []\n---\n# A\n"
	if got := StoredText(upload); got != "# A\n" {
		t.Errorf("upload front matter not removed: %q", got)
	}
	layout := "---\ntitle: X\nheader: {left: A}\n---\n# A\n"
	if got := StoredText(layout); got != layout {
		t.Errorf("layout front matter must stay: %q", got)
	}
	if got := StoredText("# A\n"); got != "# A\n" {
		t.Errorf("no front matter: %q", got)
	}
}
