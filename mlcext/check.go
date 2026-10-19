package mlcext

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Severity of a diagnostic.
type Severity string

const (
	// Error: the Sensor Monitor rejects the document or mlcdoc cannot set it.
	Error Severity = "error"
	// Warning: accepted, but probably not what the author wants.
	Warning Severity = "warning"
	// Info: intended behavior worth knowing (e.g. a field filled from an event).
	Info Severity = "info"
)

// Diagnostic is one finding, at a 1-based line and column (in characters).
type Diagnostic struct {
	Line     int      `json:"line"`
	Col      int      `json:"col"`
	Severity Severity `json:"severity"`
	// Code is stable across releases ("field-unclosed", …); Message is German.
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Options select the checks.
type Options struct {
	// RequireHeader: check the header lines of documents for the Sensor
	// Monitor. A missing "**Dokumenten-ID:**" or "**Version:**" is a warning
	// (the upload form can supply them), "**Gültig ab:**" an error (the
	// approval sets it). Off for other Markdown.
	RequireHeader bool
}

// Check validates md against the MLC Markdown Extension and returns the
// findings ordered by position.
func Check(md string, opt Options) []Diagnostic {
	var ds []Diagnostic
	add := func(line, col int, sev Severity, code, msg string) {
		ds = append(ds, Diagnostic{Line: line, Col: col, Severity: sev, Code: code, Message: msg})
	}

	checkFrontMatter(md, add)

	h := ParseHeader(md)
	if opt.RequireHeader {
		if h.RefNo == "" {
			add(1, 1, Warning, "header-id-missing", "Kopfzeile `**Dokumenten-ID:**` fehlt (z. B. `**Dokumenten-ID:** SOP-QM-001`); die Kennung muss sonst beim Hochladen eingegeben werden.")
		} else if h.Kind == "" {
			add(h.refLine, 1, Warning, "header-id-prefix", fmt.Sprintf("Dokumenten-ID %q: unbekanntes Präfix; erwartet SOP-, VA-, FB- oder INFO-.", h.RefNo))
		}
		if h.Version == "" {
			add(1, 1, Warning, "header-version-missing", "Kopfzeile `**Version:**` fehlt (z. B. `**Version:** 1.0`); die Version muss sonst beim Hochladen eingegeben werden.")
		}
		if h.validLine > 0 {
			add(h.validLine, 1, Error, "header-valid-from", "„Gültig ab“ gehört nicht ins Dokument: Die Freigabe im Sensor Monitor setzt es.")
		}
	}

	fields, errs := ExtractFields(md)
	for _, e := range errs {
		add(e.Line, e.Col, Error, e.Code, e.Message)
	}
	for _, f := range fields {
		if f.Named && IsPrefillField(f.Name) {
			add(f.Line, f.Col, Info, "field-prefill", fmt.Sprintf("Feld `%s` wird aus einem Ereignis vorausgefüllt.", f.Name))
		}
	}

	checkComments(md, add)

	slices.SortStableFunc(ds, func(a, b Diagnostic) int {
		if a.Line != b.Line {
			return a.Line - b.Line
		}
		return a.Col - b.Col
	})
	return ds
}

func checkFrontMatter(md string, add func(int, int, Severity, string, string)) {
	text := strings.TrimPrefix(md, "\xEF\xBB\xBF")
	if strings.TrimRight(strings.SplitN(text, "\n", 2)[0], "\r") != "---" {
		return
	}
	_, _, node, err := parseFrontMatter(md)
	if _, _, bodyLine := splitFrontMatter(md); bodyLine == 0 {
		add(1, 1, Warning, "frontmatter-unclosed", "Front Matter ohne abschließendes `---`: Der Block wird als Text gelesen.")
		return
	}
	if err != nil {
		add(1, 1, Error, "frontmatter-yaml", "Front Matter ist kein gültiges YAML: "+err.Error())
		return
	}
	if node == nil || len(node.Content) == 0 {
		return
	}
	m := node.Content[0]
	if m.Kind != yaml.MappingNode {
		add(2, 1, Error, "frontmatter-yaml", "Front Matter muss aus `schlüssel: wert`-Zeilen bestehen.")
		return
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		line := k.Line + 1 // the YAML starts on line 2
		switch {
		case serverOwnedKeys[strings.ToLower(k.Value)]:
			add(line, k.Column, Error, "frontmatter-server-owned", fmt.Sprintf("`%s` gehört nicht ins Dokument: Gültigkeit, Prüftermin und Status setzt die Freigabe im Sensor Monitor.", k.Value))
		case !slices.Contains(FrontMatterKeys, k.Value):
			add(line, k.Column, Error, "frontmatter-unknown-key", fmt.Sprintf("Unbekannter Schlüssel `%s`; erlaubt sind %s.", k.Value, strings.Join(FrontMatterKeys, ", ")))
		case k.Value == "event_kinds":
			if v.Kind != yaml.SequenceNode {
				add(v.Line+1, v.Column, Error, "frontmatter-event-kinds", "`event_kinds` muss eine Liste sein, z. B. `[OUT_OF_RANGE, NO_DATA]`.")
				continue
			}
			for _, e := range v.Content {
				if !IsEventKind(e.Value) {
					add(e.Line+1, e.Column, Error, "frontmatter-event-kind", fmt.Sprintf("Unbekannte Ereignisart `%s`; bekannt sind %s.", e.Value, strings.Join(EventKinds, ", ")))
				}
			}
		}
	}
}

// checkComments reports unclosed hints "<!--!" and "<!--*" (outside code).
func checkComments(md string, add func(int, int, Severity, string, string)) {
	inFence, fence := false, ""
	offset := 0
	for _, l := range strings.SplitAfter(md, "\n") {
		if f := codeFence(l); f != "" && (!inFence || strings.HasPrefix(strings.TrimSpace(l), fence)) {
			if inFence {
				inFence = false
			} else {
				inFence, fence = true, f
			}
			offset += len(l)
			continue
		}
		if !inFence {
			for _, kind := range []string{"<!--!", "<!--*"} {
				for i := 0; ; {
					j := strings.Index(l[i:], kind)
					if j < 0 {
						break
					}
					at := offset + i + j
					if !strings.Contains(md[at:], "-->") {
						line, col := position(md, at)
						add(line, col, Error, "hint-unclosed", fmt.Sprintf("Hinweis `%s` ohne schließendes `-->`: der Rest des Dokuments verschwindet.", kind))
					}
					i += j + len(kind)
				}
			}
		}
		offset += len(l)
	}
}

// Detect reports whether md is a document of the MLC Markdown Extension: it
// has a "**Dokumenten-ID:**" header line, names "event_kinds" in its front
// matter, or contains a form field (also a malformed one). Plain Markdown
// (a README with "tags:" in its front matter) is not checked, so it never
// gets MLC diagnostics.
func Detect(md string) bool {
	if ParseHeader(md).RefNo != "" {
		return true
	}
	if raw, _, n := splitFrontMatter(md); n > 0 && reEventKindsKey.MatchString(raw) {
		return true
	}
	fields, errs := ExtractFields(md)
	return len(fields) > 0 || len(errs) > 0
}
