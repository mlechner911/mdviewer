// Package markdown provides tools for converting Markdown text to sanitized HTML
// with support for syntax highlighting, math (KaTeX), diagrams (Mermaid), and Emojis.
package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/microcosm-cc/bluemonday"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Renderer handles the lifecycle of markdown-to-html conversion.
type Renderer struct {
	p *bluemonday.Policy
}

// NewRenderer initializes a new Renderer with a standard security policy.
func NewRenderer() *Renderer {
	p := bluemonday.UGCPolicy()
	// Allow classes/styles for Chroma, Mermaid, and Alerts
	p.AllowAttrs("class").OnElements("span", "code", "pre", "div", "blockquote", "p", "h1", "h2", "h3", "h4", "h5", "h6")
	p.AllowAttrs("style").OnElements("span", "code", "pre")
	p.AllowAttrs("id").OnElements("div")

	return &Renderer{
		p: p,
	}
}

// GitHubAlertTransformer handles GitHub-flavored Alerts: > [!NOTE]
type GitHubAlertTransformer struct{}

func (g *GitHubAlertTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindBlockquote {
			return ast.WalkContinue, nil
		}

		bq := n.(*ast.Blockquote)
		if bq.ChildCount() == 0 || bq.FirstChild().Kind() != ast.KindParagraph {
			return ast.WalkContinue, nil
		}

		p := bq.FirstChild().(*ast.Paragraph)
		if p.ChildCount() == 0 || p.FirstChild().Kind() != ast.KindText {
			return ast.WalkContinue, nil
		}

		t := p.FirstChild().(*ast.Text)
		content := string(t.Segment.Value(reader.Source()))
		trimmed := strings.TrimSpace(content)

		alerts := []string{"[!NOTE]", "[!TIP]", "[!IMPORTANT]", "[!WARNING]", "[!CAUTION]"}
		for _, alert := range alerts {
			if strings.HasPrefix(trimmed, alert) {
				typeStr := strings.ToLower(strings.Trim(alert, "[!]"))
				bq.SetAttributeString("class", []byte("markdown-alert markdown-alert-"+typeStr))
				
				// Calculate removal
				idx := strings.Index(content, alert)
				if idx != -1 {
					offset := idx + len(alert)
					// Skip potential newline or space after the marker
					if offset < len(content) && (content[offset] == '\n' || content[offset] == ' ') {
						offset++
					}
					
					if offset < len(content) {
						t.Segment = text.NewSegment(t.Segment.Start+offset, t.Segment.Stop)
					} else {
						t.Segment = text.NewSegment(t.Segment.Stop, t.Segment.Stop)
					}
				}
				
				fmt.Printf("[DEBUG] Detected Alert: %s on blockquote\n", typeStr)
				return ast.WalkContinue, nil
			}
		}

		return ast.WalkContinue, nil
	})
}

// GetStyleCSS returns the CSS definitions for a given Chroma style name.
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

// Render parses the input markdown and converts it to a sanitized HTML string.
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
			emoji.Emoji,
			mathjax.MathJax,
			highlighting.NewHighlighting(
				highlighting.WithStyle(chromaStyle),
				highlighting.WithFormatOptions(
					html.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithASTTransformers(
				util.Prioritized(&GitHubAlertTransformer{}, 100),
			),
		),
	)

	if err := md.Convert([]byte(input), &buf); err != nil {
		return "", err
	}

	return r.p.Sanitize(buf.String()), nil
}