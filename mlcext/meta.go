package mlcext

import (
	"regexp"
	"strings"
)

// Meta is what identifies a document: id, version and title. The front
// matter is the only place for them in new documents (doc_id, version,
// title, like Hugo and Jekyll); the header lines "**Dokumenten-ID:** … |
// **Version:** …" in the text are the legacy form and only fill in what the
// front matter leaves open.
type Meta struct {
	DocID   string `json:"docId"`
	Version string `json:"version"`
	Title   string `json:"title"`
	// Kind from the DocID prefix: sop | va | form | info, empty if unknown.
	Kind string `json:"kind"`
	// Source of DocID and Version: "frontmatter", "header" or "" (missing).
	DocIDSource   string `json:"docIdSource"`
	VersionSource string `json:"versionSource"`
}

// DocumentMeta returns the id, version and title of md: front matter first,
// header lines as a fallback.
func DocumentMeta(md string) Meta {
	fm, _ := ParseFrontMatter(md)
	h := ParseHeader(md)
	m := Meta{Title: fm.Title}
	switch {
	case fm.DocID != "":
		m.DocID, m.DocIDSource = fm.DocID, "frontmatter"
	case h.RefNo != "":
		m.DocID, m.DocIDSource = h.RefNo, "header"
	}
	switch {
	case strings.TrimSpace(fm.Version) != "":
		m.Version, m.VersionSource = strings.TrimSpace(fm.Version), "frontmatter"
	case h.Version != "":
		m.Version, m.VersionSource = h.Version, "header"
	}
	if m.Title == "" {
		m.Title = h.Title
	}
	m.Kind = KindFromRef(m.DocID)
	return m
}

// Variables are the only names "{{…}}" may use in a document. They stay in
// the file and are replaced when it is rendered (StoredText), so they stay
// right after a new version or a changed id. No logic, no loops, no own
// variables.
var Variables = []string{"doc_id", "version", "title"}

// Value of variable name in m ("" when unknown or empty).
func (m Meta) Value(name string) string {
	switch name {
	case "doc_id":
		return m.DocID
	case "version":
		return m.Version
	case "title":
		return m.Title
	}
	return ""
}

// VarRef is one "{{…}}" in the text.
type VarRef struct {
	// Name as written, trimmed ("doc_id", "kennung|yaml").
	Name string
	// Start and End: byte offsets of "{{" and after "}}".
	Start, End int
}

var reVar = regexp.MustCompile(`\{\{([^{}\n]*)\}\}`)

// varRefs finds the "{{…}}" in md outside the front matter, code blocks,
// inline code and a signature block.
func varRefs(md string) []VarRef {
	b := frontMatterBlock(md)
	start := len(md) - len(b.body)
	end := endSignatureAt(md)
	if end < start {
		end = start
	}
	var refs []VarRef
	inFence, fence := false, ""
	off := start
	for _, l := range strings.SplitAfter(md[start:end], "\n") {
		if f := codeFence(l); f != "" && (!inFence || strings.HasPrefix(strings.TrimSpace(l), fence)) {
			if inFence {
				inFence = false
			} else {
				inFence, fence = true, f
			}
		} else if !inFence {
			code := codeSpans(l)
			for _, m := range reVar.FindAllStringSubmatchIndex(l, -1) {
				if inSpans(code, m[0]) {
					continue
				}
				refs = append(refs, VarRef{Name: strings.TrimSpace(l[m[2]:m[3]]), Start: off + m[0], End: off + m[1]})
			}
		}
		off += len(l)
	}
	return refs
}

// codeSpans are the [start, end) ranges of inline code in one line.
func codeSpans(l string) [][2]int {
	var spans [][2]int
	for i := 0; i < len(l); {
		if l[i] != '`' {
			i++
			continue
		}
		n := 0
		for i+n < len(l) && l[i+n] == '`' {
			n++
		}
		if j := strings.Index(l[i+n:], l[i:i+n]); j >= 0 {
			spans = append(spans, [2]int{i, i + n + j + n})
			i += n + j + n
		} else {
			i += n
		}
	}
	return spans
}

func inSpans(spans [][2]int, at int) bool {
	for _, s := range spans {
		if at >= s[0] && at < s[1] {
			return true
		}
	}
	return false
}

func isVariable(name string) bool {
	for _, v := range Variables {
		if v == name {
			return true
		}
	}
	return false
}

// ExpandVariables replaces {{doc_id}}, {{version}} and {{title}} in md with
// the values from its front matter (header lines as a fallback). Unknown
// names and variables without a value stay as written, so the problem is
// visible; Check reports them.
func ExpandVariables(md string) string {
	refs := varRefs(md)
	if len(refs) == 0 {
		return md
	}
	m := DocumentMeta(md)
	var b strings.Builder
	last := 0
	for _, r := range refs {
		v := m.Value(r.Name)
		if !isVariable(r.Name) || v == "" {
			continue
		}
		b.WriteString(md[last:r.Start])
		b.WriteString(v)
		last = r.End
	}
	b.WriteString(md[last:])
	return b.String()
}

// SuggestFileName proposes a file name from the document id: "INFO-QM-001"
// plus ext (".md"), with the version for exports that are passed on
// ("INFO-QM-001_v1.0.pdf"). Empty without a doc_id: the caller keeps its
// usual name.
func SuggestFileName(md, ext string, withVersion bool) string {
	m := DocumentMeta(md)
	if m.DocID == "" {
		return ""
	}
	name := m.DocID
	if withVersion && m.Version != "" {
		name += "_v" + strings.TrimPrefix(strings.TrimPrefix(m.Version, "v"), "V")
	}
	name = strings.TrimSpace(badFileChars.ReplaceAllString(name, "_"))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return name + ext
}
