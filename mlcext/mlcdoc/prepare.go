package mlcdoc

import (
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Keys are the front matter keys mlcdoc 0.16.0 accepts; it rejects any
// other key. Taken from mlcdoc itself (the "expected one of" list of its
// error); TestKeysMatchBinary compares it with the installed binary.
var Keys = []string{
	"title", "author", "subject", "keywords", "creator", "label", "header", "footer", "lang",
	"signatures", "signature", "date", "watermark", "timezone", "pdf-standard", "font",
	"mono-font", "font-size", "heading-sizes", "code-font-size", "table-font-size",
	"header-font-size", "page-size", "landscape", "slides", "toc", "bookmarks",
	"slide-template", "margin", "accent", "code-theme", "table-style", "task-style",
	"table-header-repeat", "line-numbers", "no-header", "page-break-before", "no-external",
	"allow-http", "meta",
}

// Prepare makes md acceptable for mlcdoc: front matter keys mlcdoc does
// not know are removed and returned. That covers documents of other tools
// (tags, aliases …) as well as the Sensor Monitor's upload keys
// (event_kinds, version, description), which the server strips on upload
// too; title stays as the PDF title. A document without front matter, or
// with invalid YAML (mlcdoc reports that), is returned unchanged.
func Prepare(md string) (string, []string) {
	text := strings.TrimPrefix(md, "\xEF\xBB\xBF")
	lines := strings.Split(text, "\n")
	if len(lines) < 2 || strings.TrimRight(lines[0], "\r") != "---" {
		return md, nil
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if t := strings.TrimRight(lines[i], "\r"); t == "---" || t == "..." {
			end = i
			break
		}
	}
	if end < 0 {
		return md, nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:end], "\n")), &doc); err != nil || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return md, nil
	}
	m := doc.Content[0]
	var ignored []string
	kept := m.Content[:0:0]
	for i := 0; i+1 < len(m.Content); i += 2 {
		if slices.Contains(Keys, m.Content[i].Value) {
			kept = append(kept, m.Content[i], m.Content[i+1])
		} else {
			ignored = append(ignored, m.Content[i].Value)
		}
	}
	if len(ignored) == 0 {
		return md, nil
	}
	body := strings.Join(lines[end+1:], "\n")
	if len(kept) == 0 {
		return strings.TrimLeft(body, "\r\n"), ignored
	}
	m.Content = kept
	out, err := yaml.Marshal(&doc)
	if err != nil {
		return md, nil
	}
	return "---\n" + string(out) + "---\n" + body, ignored
}
