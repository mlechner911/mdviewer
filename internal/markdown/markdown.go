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
	p *bluemonday.Policy
}

// NewRenderer creates a new markdown renderer
func NewRenderer() *Renderer {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").OnElements("span", "code", "pre", "div")
	p.AllowAttrs("style").OnElements("span", "code", "pre")
	p.AllowAttrs("id").OnElements("div")

	return &Renderer{
		p: p,
	}
}

// Render converts markdown string to sanitized HTML using a specific Chroma style
func (r *Renderer) Render(input string, chromaStyle string) (string, error) {
	var buf bytes.Buffer
	
	// Default to github-dark if empty
	if chromaStyle == "" {
		chromaStyle = "github-dark"
	}
	
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.Typographer,
			mathjax.Mathjax,
			highlighting.NewHighlighting(
				highlighting.WithStyle(chromaStyle),
			),
		),
	)

	if err := md.Convert([]byte(input), &buf); err != nil {
		return "", err
	}

	return r.p.Sanitize(buf.String()), nil
}
