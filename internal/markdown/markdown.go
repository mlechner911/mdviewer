package markdown

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
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

// GetStyleCSS returns the CSS for a given Chroma style
func (r *Renderer) GetStyleCSS(styleName string) (string, error) {
	s := styles.Get(styleName)
	if s == nil {
		s = styles.Fallback
	}
	
	formatter := html.New(html.WithClasses(true))
	var buf bytes.Buffer
	err := formatter.WriteCSS(&buf, s)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Render converts markdown string to sanitized HTML using a specific Chroma style
func (r *Renderer) Render(input string, chromaStyle string) (string, error) {
	var buf bytes.Buffer
	
	if chromaStyle == "" {
		chromaStyle = "github-dark"
	}
	
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.Typographer,
			mathjax.MathJax,
			highlighting.NewHighlighting(
				highlighting.WithStyle(chromaStyle),
				highlighting.WithFormatOptions(
					html.WithClasses(true),
				),
			),
		),
	)

	if err := md.Convert([]byte(input), &buf); err != nil {
		return "", err
	}

	return r.p.Sanitize(buf.String()), nil
}