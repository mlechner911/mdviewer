package mlcext

import (
	"bytes"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Header holds the document header lines of the MLC Markdown Extension,
// read from the first lines up to the first "---" (at most 40 lines):
//
//	**Dokumenten-ID:** SOP-KAL-001 | **Version:** 1.0
//
// Placeholders such as "[Datum]" count as empty. A trailing "\" (a Markdown
// line break) is not part of the value.
type Header struct {
	// Title from "**Titel:** …" (SOP templates); informational, the Sensor
	// Monitor takes the title from the upload form or the front matter.
	Title   string
	RefNo   string
	Version string
	// ValidFrom is the raw "Gültig ab" value. Validity is set by the
	// approval in the Sensor Monitor; the check reports it as an error.
	ValidFrom string
	// Kind from the RefNo prefix: sop | va | form | info, empty if unknown.
	Kind string

	refLine, versionLine, validLine int
}

var hdrField = regexp.MustCompile(`\*\*\s*(Titel|Dokumenten-ID|Dokument-ID|Version|Gültig ab)\s*:\s*\*\*\s*([^|\n*]*)`)

// ParseHeader reads the header lines of md (front matter is skipped).
func ParseHeader(md string) Header {
	_, body, offset := splitFrontMatter(md)
	ls := strings.Split(body, "\n")
	if len(ls) > 40 {
		ls = ls[:40]
	}
	var h Header
	for i, l := range ls {
		if i > 0 && strings.TrimSpace(l) == "---" && (h.RefNo != "" || h.Version != "") {
			break
		}
		for _, m := range hdrField.FindAllStringSubmatch(l, -1) {
			v := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(m[2]), `\`))
			line := offset + i + 1
			switch m[1] {
			case "Gültig ab":
				if h.validLine == 0 {
					h.ValidFrom, h.validLine = v, line
				}
				continue
			}
			if v == "" || strings.HasPrefix(v, "[") {
				continue // placeholder "[Datum]"
			}
			switch m[1] {
			case "Titel":
				if h.Title == "" {
					h.Title = v
				}
			case "Dokumenten-ID", "Dokument-ID":
				if h.RefNo == "" {
					h.RefNo, h.refLine = v, line
				}
			case "Version":
				if h.Version == "" {
					h.Version, h.versionLine = v, line
				}
			}
		}
	}
	h.Kind = KindFromRef(h.RefNo)
	return h
}

// KindFromRef is the document kind from the RefNo prefix (SOP-…, VA-…,
// FB-…, INFO-…); empty when the prefix says nothing.
func KindFromRef(ref string) string {
	switch strings.ToUpper(strings.SplitN(ref, "-", 2)[0]) {
	case "SOP":
		return "sop"
	case "VA":
		return "va"
	case "FB":
		return "form"
	case "INFO", "IN":
		return "info"
	}
	return ""
}

// FrontMatter is the optional YAML block at the top of a document (after a
// signature block of the old front form, see SplitSignature).
type FrontMatter struct {
	// DocID is the document id ("SOP-QM-001"): the only place for it in
	// new documents; the "**Dokumenten-ID:**" line is the legacy form.
	DocID       string
	Title       string
	Version     string
	Description string
	EventKinds  []string
	// EventKindsSet: "event_kinds" is present, also when empty
	// ("event_kinds: []" means "applies always").
	EventKindsSet bool
	// Present: the document starts with a front matter block.
	Present bool
}

// ParseFrontMatter returns the front matter and the body after it. Without
// a block, an unclosed block or invalid YAML the whole input is the body, as
// in the Sensor Monitor.
func ParseFrontMatter(md string) (FrontMatter, string) {
	fm, body, _, _ := parseFrontMatter(md)
	return fm, body
}

// splitFrontMatter returns the raw YAML, the body and the number of lines
// before the body (0: no front matter).
func splitFrontMatter(md string) (raw, body string, bodyLine int) {
	b := frontMatterBlock(md)
	if !b.present {
		return "", md, 0
	}
	return b.raw, b.body, b.bodyLine
}

// fmBlock is the document's front matter as found in the file.
type fmBlock struct {
	raw, body string
	// present: a closed block; open: the file starts a block it never closes.
	present, open bool
	// yamlLine: 1-based line of the first YAML line; bodyLine: lines before
	// the body. Both count from the start of the file.
	yamlLine, bodyLine int
}

// frontMatterBlock finds the document's front matter. A signature block of
// the old front form is not it: the document's own block follows it.
// Without a block, body is md after such a signature.
func frontMatterBlock(md string) fmBlock {
	skip := skipFrontSignature(md)
	before := strings.Count(md[:skip], "\n")
	text := strings.TrimPrefix(md[skip:], "\xEF\xBB\xBF")
	ls := strings.Split(text, "\n")
	if len(ls) == 0 || strings.TrimRight(ls[0], "\r") != "---" {
		return fmBlock{body: md[skip:], bodyLine: before}
	}
	for i := 1; i < len(ls); i++ {
		if t := strings.TrimRight(ls[i], "\r"); t == "---" || t == "..." {
			rest := strings.Join(ls[i+1:], "\n")
			trimmed := strings.TrimLeft(rest, "\r\n")
			skipped := strings.Count(rest[:len(rest)-len(trimmed)], "\n")
			return fmBlock{
				raw: strings.Join(ls[1:i], "\n"), body: trimmed, present: true,
				yamlLine: before + 2, bodyLine: before + i + 1 + skipped,
			}
		}
	}
	return fmBlock{body: md[skip:], bodyLine: before, open: true}
}

var reEventKindsKey = regexp.MustCompile(`(?m)^event_kinds\s*:`)

// parseFrontMatter also returns the YAML node (nil without a valid block)
// and a YAML error, for the check.
func parseFrontMatter(md string) (FrontMatter, string, *yaml.Node, error) {
	b := frontMatterBlock(md)
	if !b.present {
		return FrontMatter{}, b.body, nil, nil
	}
	raw, body := b.raw, b.body
	var node yaml.Node
	if err := yaml.NewDecoder(bytes.NewReader([]byte(raw))).Decode(&node); err != nil && err.Error() != "EOF" {
		return FrontMatter{Present: true}, md, nil, err
	}
	var v struct {
		DocID       string   `yaml:"doc_id"`
		Title       string   `yaml:"title"`
		Version     string   `yaml:"version"`
		Description string   `yaml:"description"`
		EventKinds  []string `yaml:"event_kinds"`
	}
	fm := FrontMatter{Present: true}
	if err := node.Decode(&v); err == nil {
		fm.DocID, fm.Title, fm.Version, fm.Description, fm.EventKinds = strings.TrimSpace(v.DocID), v.Title, v.Version, v.Description, v.EventKinds
	}
	fm.EventKindsSet = reEventKindsKey.MatchString(raw)
	return fm, body, &node, nil
}

// KindLabel is the German name of a document kind from KindFromRef.
func KindLabel(kind string) string {
	switch kind {
	case "sop":
		return "Standardarbeitsanweisung (SOP)"
	case "va":
		return "Verfahrensanweisung (VA)"
	case "form":
		return "Formblatt (FB)"
	case "info":
		return "Information (INFO)"
	}
	return ""
}

// StoredText is md as it is rendered (HTML, and the text mlcdoc gets):
//   - a signature block (either form, see SplitSignature) is removed;
//   - {{doc_id}}, {{version}} and {{title}} are replaced (ExpandVariables);
//   - an upload front matter (only the keys in FrontMatterKeys) is removed,
//     as on upload; mlcdoc rejects those keys. Front matter with other keys
//     (mlcdoc layout such as header or footer) is kept.
func StoredText(md string) string {
	_, md = SplitSignature(md)
	md = ExpandVariables(md)
	fm, body, node, err := parseFrontMatter(md)
	if !fm.Present || err != nil || node == nil {
		return md
	}
	if len(node.Content) > 0 {
		m := node.Content[0]
		for i := 0; i+1 < len(m.Content); i += 2 {
			if !slices.Contains(FrontMatterKeys, m.Content[i].Value) {
				return md
			}
		}
	}
	return body
}
