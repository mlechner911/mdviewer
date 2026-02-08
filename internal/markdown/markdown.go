package markdown

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

// Renderer handles markdown to HTML conversion
type Renderer struct {
	md goldmark.Markdown
	p  *bluemonday.Policy
}

// NewRenderer creates a new markdown renderer with GFM and highlighting
func NewRenderer() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.Typographer,
			mathjax.Mathjax,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github-dark"),
			),
		),
	)

	// Security-Policy: Erlaube Klassen und Styles für Chroma-Highlighting und Mermaid
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").OnElements("span", "code", "pre", "div")
	p.AllowAttrs("style").OnElements("span", "code", "pre")
	p.AllowAttrs("id").OnElements("div")

	return &Renderer{
		md: md,
		p:  p,
	}
}

// Render converts markdown string to sanitized HTML
func (r *Renderer) Render(input string) (string, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(input), &buf); err != nil {
		return "", err
	}

	return r.p.Sanitize(buf.String()), nil
}
