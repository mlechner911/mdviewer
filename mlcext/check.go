package mlcext

import (
	"fmt"
	"slices"
	"strings"
	"sync"

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
	// RequireHeader: check id and version of documents for the Sensor
	// Monitor. A missing doc_id or version (front matter, or the legacy
	// header lines) is a warning (the upload form can supply them),
	// "**Gültig ab:**" an error (the approval sets it). Off for other
	// Markdown.
	RequireHeader bool
}

// Check validates md against the MLC Markdown Extension and returns the
// findings ordered by position.
func Check(md string, opt Options) []Diagnostic {
	var ds []Diagnostic
	add := func(line, col int, sev Severity, code, msg string) {
		ds = append(ds, Diagnostic{Line: line, Col: col, Severity: sev, Code: code, Message: msg})
	}

	keyLines := checkFrontMatter(md, add)

	h := ParseHeader(md)
	fm, _ := ParseFrontMatter(md)
	meta := DocumentMeta(md)
	idLine := 1
	if l := keyLines["doc_id"]; l > 0 && meta.DocIDSource == "frontmatter" {
		idLine = l
	} else if meta.DocIDSource == "header" {
		idLine = h.refLine
	}
	if opt.RequireHeader {
		if meta.DocID == "" {
			add(1, 1, Warning, "header-id-missing", "Kennung fehlt: `doc_id` ins Front Matter schreiben (z. B. `doc_id: SOP-QM-001`); sie muss sonst beim Hochladen eingegeben werden.")
		} else if meta.Kind == "" {
			add(idLine, 1, Warning, "header-id-prefix", fmt.Sprintf("Kennung %q: unbekanntes Präfix; erwartet SOP-, VA-, FB- oder INFO-.", meta.DocID))
		}
		if meta.Version == "" {
			add(1, 1, Warning, "header-version-missing", "Version fehlt: `version` ins Front Matter schreiben (z. B. `version: \"1.0\"`); sie muss sonst beim Hochladen eingegeben werden.")
		}
		if h.validLine > 0 {
			add(h.validLine, 1, Error, "header-valid-from", "„Gültig ab“ gehört nicht ins Dokument: Die Freigabe im Sensor Monitor setzt es.")
		}
	}
	// Legacy header lines: still read, but the front matter is the place.
	if line := firstNonZero(h.refLine, h.versionLine); line > 0 {
		add(line, 1, Warning, "header-in-body", "Kennung und Version gehören ins Front Matter (`doc_id`, `version`); die Kopfzeile im Text ist die alte Form.")
	}
	if fm.DocID != "" && h.RefNo != "" && !strings.EqualFold(fm.DocID, h.RefNo) {
		add(h.refLine, 1, Error, "header-conflict", fmt.Sprintf("Kennung widerspricht sich: Front Matter `%s`, Kopfzeile `%s`.", fm.DocID, h.RefNo))
	}
	if v := strings.TrimSpace(fm.Version); v != "" && h.Version != "" && v != h.Version {
		add(h.versionLine, 1, Error, "header-conflict", fmt.Sprintf("Version widerspricht sich: Front Matter `%s`, Kopfzeile `%s`.", v, h.Version))
	}

	checkVariables(md, meta, add)

	if sig, _ := SplitSignature(md); sig != nil {
		msg := "Dokument ist signiert"
		if by := sig.SignedBy(); by != "" {
			msg += " von " + by
		}
		if at := sig.SignedAt(); at != "" {
			msg += " am " + at
		}
		msg += "; wer es ändert, macht die Signatur ungültig (neue Fassung)."
		add(sig.Line, 1, Info, "signed", msg)
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
	checkGlyphs(md, add)
	checkTables(md, add)

	slices.SortStableFunc(ds, func(a, b Diagnostic) int {
		if a.Line != b.Line {
			return a.Line - b.Line
		}
		return a.Col - b.Col
	})
	return ds
}

// checkFrontMatter checks the front matter and returns the line of each key.
func checkFrontMatter(md string, add func(int, int, Severity, string, string)) map[string]int {
	keyLines := map[string]int{}
	b := frontMatterBlock(md)
	if b.open {
		add(b.bodyLine+1, 1, Warning, "frontmatter-unclosed", "Front Matter ohne abschließendes `---`: Der Block wird als Text gelesen.")
		return keyLines
	}
	if !b.present {
		return keyLines
	}
	_, _, node, err := parseFrontMatter(md)
	if err != nil {
		add(b.yamlLine-1, 1, Error, "frontmatter-yaml", "Front Matter ist kein gültiges YAML: "+err.Error())
		return keyLines
	}
	if node == nil || len(node.Content) == 0 {
		return keyLines
	}
	m := node.Content[0]
	if m.Kind != yaml.MappingNode {
		add(b.yamlLine, 1, Error, "frontmatter-yaml", "Front Matter muss aus `schlüssel: wert`-Zeilen bestehen.")
		return keyLines
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		line := k.Line + b.yamlLine - 1
		keyLines[k.Value] = line
		switch {
		case strings.EqualFold(k.Value, "watermark"):
			add(line, k.Column, Error, "frontmatter-server-owned", "`watermark` gehört nicht ins Dokument: Das Wasserzeichen setzt der Sensor Monitor beim Drucken nach dem Stand der Fassung, MarkSafe nach seiner Einstellung.")
		case serverOwnedKeys[strings.ToLower(k.Value)]:
			add(line, k.Column, Error, "frontmatter-server-owned", fmt.Sprintf("`%s` gehört nicht ins Dokument: Gültigkeit, Prüftermin und Status setzt die Freigabe im Sensor Monitor.", k.Value))
		case !slices.Contains(FrontMatterKeys, k.Value):
			add(line, k.Column, Error, "frontmatter-unknown-key", fmt.Sprintf("Unbekannter Schlüssel `%s`; erlaubt sind %s.", k.Value, strings.Join(FrontMatterKeys, ", ")))
		case k.Value == "event_kinds":
			if v.Kind != yaml.SequenceNode {
				add(v.Line+b.yamlLine-1, v.Column, Error, "frontmatter-event-kinds", "`event_kinds` muss eine Liste sein, z. B. `[OUT_OF_RANGE, NO_DATA]`.")
				continue
			}
			for _, e := range v.Content {
				if !IsEventKind(e.Value) {
					add(e.Line+b.yamlLine-1, e.Column, Error, "frontmatter-event-kind", fmt.Sprintf("Unbekannte Ereignisart `%s`; bekannt sind %s.", e.Value, strings.Join(EventKinds, ", ")))
				}
			}
		case k.Value == "meta":
			if v.Kind != yaml.MappingNode {
				add(v.Line+b.yamlLine-1, v.Column, Error, "frontmatter-meta", "`meta` muss ein Block aus `schlüssel: wert`-Zeilen sein.")
				continue
			}
			for j := 0; j+1 < len(v.Content); j += 2 {
				if sk, sv := v.Content[j], v.Content[j+1]; sk.Value == "schema" && sv.Value != FormRecordSchema {
					add(sv.Line+b.yamlLine-1, sv.Column, Warning, "frontmatter-meta-schema", fmt.Sprintf("Unbekanntes `meta.schema` `%s`; bekannt ist %s.", sv.Value, FormRecordSchema))
				}
			}
		case k.Value == "doc_id" || k.Value == "version" || k.Value == "title":
			if v.Kind != yaml.ScalarNode {
				add(v.Line+b.yamlLine-1, v.Column, Error, "frontmatter-yaml", fmt.Sprintf("`%s` muss ein einzelner Wert sein.", k.Value))
			}
		}
	}
	return keyLines
}

// checkVariables reports "{{…}}" in the text: unknown names, leftover
// template placeholders and variables without a value.
func checkVariables(md string, meta Meta, add func(int, int, Severity, string, string)) {
	for _, r := range varRefs(md) {
		line, col := position(md, r.Start)
		switch {
		case isVariable(r.Name):
			if meta.Value(r.Name) == "" {
				add(line, col, Warning, "variable-empty", fmt.Sprintf("`{{%s}}` hat keinen Wert: `%s` fehlt im Front Matter.", r.Name, r.Name))
			}
		case strings.Contains(r.Name, "|") || isTemplateFieldName(r.Name):
			add(line, col, Error, "template-placeholder-left", fmt.Sprintf("Platzhalter `{{%s}}` der Vorlage wurde beim Anlegen nicht ersetzt.", r.Name))
		default:
			add(line, col, Error, "unknown-variable", fmt.Sprintf("Unbekannte Variable `{{%s}}`; erlaubt sind {{doc_id}}, {{version}} und {{title}}.", r.Name))
		}
	}
}

// isTemplateFieldName: a placeholder of a built-in template's input field
// (their names are the usual ones; a template from a folder may use others,
// which then count as unknown variables, also an error).
func isTemplateFieldName(name string) bool {
	templateFieldsOnce.Do(func() {
		// Field names of the old template format, still in circulation.
		templateFields = map[string]bool{"kennung": true, "titel": true, "anlass": true, "bereich": true}
		for _, t := range LoadTemplates() {
			for _, f := range t.Fields {
				templateFields[f.ID] = true
			}
		}
	})
	return templateFields[name]
}

var (
	templateFieldsOnce sync.Once
	templateFields     map[string]bool
)

func firstNonZero(xs ...int) int {
	best := 0
	for _, x := range xs {
		if x > 0 && (best == 0 || x < best) {
			best = x
		}
	}
	return best
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
// has a doc_id (front matter or "**Dokumenten-ID:**" line), a signature,
// names "event_kinds" in its front matter, or contains a form field (also a malformed one). Plain Markdown
// (a README with "tags:" in its front matter) is not checked, so it never
// gets MLC diagnostics.
func Detect(md string) bool {
	if DocumentMeta(md).DocID != "" {
		return true
	}
	if sig, _ := SplitSignature(md); sig != nil {
		return true
	}
	if raw, _, n := splitFrontMatter(md); n > 0 && reEventKindsKey.MatchString(raw) {
		return true
	}
	fields, errs := ExtractFields(md)
	return len(fields) > 0 || len(errs) > 0
}
