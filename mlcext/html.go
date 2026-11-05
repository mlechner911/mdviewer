package mlcext

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// HTML in Markdown and the PDF: the HTML preview shows it as GitHub does,
// mlcdoc does not. These rules say what becomes of it in the PDF, so the PDF
// path (PDFHTML) and the editor's notes (Options.PDF) agree:
//   - <b>, <strong>, <i>, <em>, <s>, <del>, <strike>, <code> become Markdown;
//   - <br>, <a> and comments work as they are;
//   - other inline tags (<u>, <sub>, <span> …) lose their tags, the text stays;
//   - HTML blocks (<div>, <p>, <details>, <table> …) and <img> do not appear,
//     except a page break.

var (
	reHTMLTag = regexp.MustCompile(`</?([A-Za-z][A-Za-z0-9-]*)(?:\s[^<>]*)?/?>`)
	// CommonMark HTML block kind 6 (and kind 1), the names that start a block.
	htmlBlockNames = map[string]bool{
		"address": true, "article": true, "aside": true, "base": true, "basefont": true, "blockquote": true, "body": true,
		"caption": true, "center": true, "col": true, "colgroup": true, "dd": true, "details": true, "dialog": true,
		"dir": true, "div": true, "dl": true, "dt": true, "fieldset": true, "figcaption": true, "figure": true,
		"footer": true, "form": true, "frame": true, "frameset": true, "h1": true, "h2": true, "h3": true, "h4": true,
		"h5": true, "h6": true, "head": true, "header": true, "hr": true, "html": true, "iframe": true, "legend": true,
		"li": true, "link": true, "main": true, "menu": true, "menuitem": true, "nav": true, "noframes": true, "ol": true,
		"optgroup": true, "option": true, "p": true, "param": true, "search": true, "section": true, "summary": true,
		"table": true, "tbody": true, "td": true, "tfoot": true, "th": true, "thead": true, "title": true, "tr": true,
		"track": true, "ul": true, "script": true, "pre": true, "style": true, "textarea": true,
	}
	htmlMarkdown = map[string]string{"b": "**", "strong": "**", "i": "*", "em": "*", "s": "~~", "del": "~~", "strike": "~~", "code": "`"}
	htmlKept     = map[string]bool{"br": true, "a": true}
)

type htmlKind int

const (
	htmlConvert htmlKind = iota // becomes Markdown
	htmlStrip                   // tag goes, text stays
	htmlBlock                   // does not appear in the PDF
)

// htmlItem is one tag (or one HTML block) the PDF treats differently.
type htmlItem struct {
	start, end int
	kind       htmlKind
	name       string
	marker     string // htmlConvert
	closing    bool
}

// scanHTML finds the HTML of md outside front matter and code.
func scanHTML(md string) []htmlItem {
	_, _, bodyLine := splitFrontMatter(md)
	var items []htmlItem
	lines := strings.SplitAfter(md, "\n")
	offset := 0
	fence, block := "", ""
	for n, line := range lines {
		start := offset
		offset += len(line)
		if n < bodyLine {
			continue
		}
		text := strings.TrimRight(line, "\r\n")
		if fence != "" {
			if strings.HasPrefix(strings.TrimSpace(text), fence) {
				fence = ""
			}
			continue
		}
		if f := codeFence(text); f != "" && block == "" {
			fence = f
			continue
		}
		if block == "comment" {
			if strings.Contains(text, "-->") {
				items, block = items[:len(items)-1], ""
			}
			continue
		}
		if block != "" {
			if strings.TrimSpace(text) != "" {
				items[len(items)-1].end = start + len(text)
			}
			if strings.TrimSpace(text) == "" {
				block = ""
			}
			continue
		}
		if indent := len(text) - len(strings.TrimLeft(text, " ")); indent <= 3 {
			t := text[indent:]
			if strings.HasPrefix(t, "<!--") {
				if !strings.Contains(t, "-->") {
					block = "comment"
					items = append(items, htmlItem{start: start, end: start + len(text), kind: htmlBlock})
				}
				continue
			}
			if name, ok := htmlBlockStart(t); ok {
				items = append(items, htmlItem{start: start + indent, end: start + len(text), kind: htmlBlock, name: name})
				if strings.Contains(strings.ToLower(t), "page-break") {
					items = items[:len(items)-1] // the PDF breaks the page
				} else if strings.TrimSpace(t) != "" {
					block = name
				}
				continue
			}
		}
		items = append(items, inlineHTML(text, start)...)
	}
	return items
}

// htmlBlockStart reports whether line starts an HTML block (CommonMark kinds
// 1, 6 and 7) and its tag name.
func htmlBlockStart(line string) (string, bool) {
	m := reHTMLTag.FindStringSubmatchIndex(line)
	if m == nil || m[0] != 0 {
		return "", false
	}
	name := strings.ToLower(line[m[2]:m[3]])
	if htmlBlockNames[name] {
		return name, true
	}
	// Kind 7: a complete tag alone on its line (<img …>, <span>).
	if strings.TrimSpace(line[m[1]:]) == "" {
		return name, true
	}
	return "", false
}

// inlineHTML are the tags of one line outside code spans. Tags with a
// Markdown form are converted only in pairs on the line; a lone one is
// stripped.
func inlineHTML(line string, offset int) []htmlItem {
	var items []htmlItem
	var open []int
	code := codeSpans(line)
	for _, m := range reHTMLTag.FindAllStringSubmatchIndex(line, -1) {
		if inSpans(code, m[0]) || (m[0] > 0 && line[m[0]-1] == '\\') {
			continue
		}
		name := strings.ToLower(line[m[2]:m[3]])
		if htmlKept[name] {
			continue
		}
		it := htmlItem{start: offset + m[0], end: offset + m[1], name: name, closing: line[m[0]+1] == '/', kind: htmlStrip}
		if name == "img" {
			it.kind = htmlBlock
		}
		items = append(items, it)
		if marker, ok := htmlMarkdown[name]; ok {
			i := len(items) - 1
			if !it.closing {
				open = append(open, i)
				continue
			}
			for k := len(open) - 1; k >= 0; k-- {
				if items[open[k]].name == name {
					items[open[k]].kind, items[open[k]].marker = htmlConvert, marker
					items[i].kind, items[i].marker = htmlConvert, marker
					open = append(open[:k], open[k+1:]...)
					break
				}
			}
		}
	}
	return items
}

// PDFHTML is md as the PDF should get it: HTML with a Markdown form
// converted, other inline tags removed, HTML blocks left out. notes say what
// does not appear as written, for the export's warnings.
func PDFHTML(md string) (string, []string) {
	items := scanHTML(md)
	if len(items) == 0 {
		return md, nil
	}
	var out []byte
	last := 0
	blocks, stripped := map[string]int{}, map[string]int{}
	for _, it := range items {
		out = append(out, md[last:it.start]...)
		last = it.end
		switch it.kind {
		case htmlBlock:
			blocks["<"+it.name+">"]++
		case htmlStrip:
			if !it.closing {
				stripped["<"+it.name+">"]++
			}
		case htmlConvert:
			if it.closing {
				// "<b>fett </b>" must end before the blank: "**fett** ".
				trimmed := strings.TrimRight(string(out), " \t")
				spaces := string(out[len(trimmed):])
				out = append(append([]byte(trimmed), it.marker...), spaces...)
			} else {
				rest := md[it.end:]
				lead := len(rest) - len(strings.TrimLeft(rest, " \t"))
				out = append(append(out, rest[:lead]...), it.marker...)
				last = it.end + lead
			}
		}
	}
	out = append(out, md[last:]...)
	var notes []string
	if s := countList(blocks); s != "" {
		notes = append(notes, "HTML erscheint nicht im PDF: "+s)
	}
	if s := countList(stripped); s != "" {
		notes = append(notes, "HTML im PDF ohne Formatierung: "+s)
	}
	return string(out), notes
}

func countList(m map[string]int) string {
	var parts []string
	for k, n := range m {
		if n > 1 {
			k += fmt.Sprintf(" (%d×)", n)
		}
		parts = append(parts, k)
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// htmlDiagnostics are the notes of PDFHTML at their place in md.
func htmlDiagnostics(md string) []Diagnostic {
	var ds []Diagnostic
	for _, it := range scanHTML(md) {
		line, col := position(md, it.start)
		switch {
		case it.kind == htmlBlock && it.name == "img":
			ds = append(ds, Diagnostic{line, col, Warning, "html-pdf-missing", "Ein Bild per `<img>` erscheint nicht im PDF; besser `![Beschreibung](datei.png)`."})
		case it.kind == htmlBlock:
			ds = append(ds, Diagnostic{line, col, Warning, "html-pdf-missing", fmt.Sprintf("HTML-Block `<%s>` erscheint nicht im PDF (in der HTML-Vorschau schon).", it.name)})
		case it.kind == htmlStrip && !it.closing:
			ds = append(ds, Diagnostic{line, col, Info, "html-pdf-plain", fmt.Sprintf("`<%s>` erscheint im PDF ohne Formatierung, nur der Text.", it.name)})
		}
	}
	return ds
}
