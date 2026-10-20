package mlcext

import (
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// LayoutOptions is what the caller adds to the layout: the parts that
// belong to printing, not to the document.
type LayoutOptions struct {
	// Watermark text across every page; empty for none. The Sensor Monitor
	// sets it by the version's status (ENTWURF, ERSETZT, none when
	// approved), MarkSafe by its setting. mlcdoc takes 1–60 characters;
	// longer text is cut.
	Watermark string
	// Organization goes into the middle of the header, Logo (a path
	// relative to the document or a data: URL) to its left.
	Organization string
	Logo         string
	// Lang of mlcdoc's built-in texts; default "de".
	Lang string
}

type layoutHeader struct {
	Logo   string `yaml:"logo,omitempty"`
	Left   string `yaml:"left,omitempty"`
	Center string `yaml:"center,omitempty"`
	Right  string `yaml:"right,omitempty"`
}

type layoutFile struct {
	Lang      string        `yaml:"lang,omitempty"`
	Title     string        `yaml:"title,omitempty"`
	Watermark string        `yaml:"watermark,omitempty"`
	Header    *layoutHeader `yaml:"header,omitempty"`
}

// LayoutYAML returns the layout file for "mlcdoc md-to-pdf --layout" for
// md: the title, a header with the kind on the left and "INFO-QM-001 ·
// Version 1.0" on the right, and what opt adds. mlcdoc lays the layout
// file over the document's own front matter, so the document (and its
// signature) stays unchanged; mlcdoc needs nothing new. Without doc_id and
// title the header stays mlcdoc's default.
func LayoutYAML(md string, opt LayoutOptions) string {
	m := DocumentMeta(md)
	l := layoutFile{Lang: opt.Lang, Title: m.Title, Watermark: watermarkText(opt.Watermark)}
	if l.Lang == "" {
		l.Lang = "de"
	}
	h := layoutHeader{Logo: opt.Logo, Center: opt.Organization}
	if m.Kind != "" {
		h.Left = kindShort(m.Kind)
	}
	if m.DocID != "" {
		h.Right = m.DocID
		if m.Version != "" {
			h.Right += " · Version " + m.Version
		}
	}
	if h != (layoutHeader{}) {
		if h.Left == "" {
			h.Left = m.Title // replaces "MLC DOCUMENTATION"
		}
		l.Header = &h
	}
	b, err := yaml.Marshal(l)
	if err != nil {
		return ""
	}
	return string(b)
}

// kindShort is the kind's name for the page header.
func kindShort(kind string) string {
	label := KindLabel(kind)
	if i := strings.Index(label, " ("); i > 0 {
		return label[:i]
	}
	return label
}

func watermarkText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) > 60 {
		s = string([]rune(s)[:60])
	}
	return s
}
