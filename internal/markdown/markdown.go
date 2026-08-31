// Package markdown provides tools for converting Markdown text to sanitized HTML
// with support for syntax highlighting, math (KaTeX), diagrams (Mermaid), Emojis, and YAML Front Matter.
package markdown

import (
	"bytes"
	_ "embed"
	"fmt"
	htmlpkg "html"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//go:embed katex.min.css.txt
var katexCSS string

// Renderer handles the lifecycle of markdown-to-html conversion.
type Renderer struct {
	p *bluemonday.Policy
}

// NewRenderer initializes a new Renderer with a standard security policy.
func NewRenderer() *Renderer {
	p := bluemonday.UGCPolicy()
	// Allow classes/styles for Chroma, Mermaid, Alerts, and Front Matter
	p.AllowAttrs("class").OnElements("span", "code", "pre", "div", "blockquote", "p", "h1", "h2", "h3", "h4", "h5", "h6", "li", "ul", "ol", "table", "tbody", "thead", "tr", "th", "td", "details", "summary", "aside")
	p.AllowAttrs("style").OnElements("span", "code", "pre", "div")
	p.AllowAttrs("id").OnElements("div")
	p.AllowAttrs("open").OnElements("details")

	// Allow Task Lists (Checkboxes)
	p.AllowElements("input")
	p.AllowAttrs("type", "checked", "disabled").OnElements("input")
	p.AllowAttrs("class").OnElements("input")

	// Allow Details / Summary for collapsible metadata
	p.AllowElements("details", "summary")

	return &Renderer{
		p: p,
	}
}

// GetKatexCSS returns the embedded KaTeX CSS.
func (r *Renderer) GetKatexCSS() string {
	return katexCSS
}

// GitHubAlertTransformer handles GitHub-flavored Alerts: > [!NOTE]
type GitHubAlertTransformer struct{}

// Transform walks the AST and applies alert classes to matching blockquotes.
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

		var fullContent strings.Builder
		for c := p.FirstChild(); c != nil; c = c.NextSibling() {
			if t, ok := c.(*ast.Text); ok {
				fullContent.Write(t.Segment.Value(reader.Source()))
			} else {
				break
			}
			if fullContent.Len() > 20 {
				break
			}
		}

		content := fullContent.String()
		trimmed := strings.TrimSpace(content)

		alerts := []string{"[!NOTE]", "[!TIP]", "[!IMPORTANT]", "[!WARNING]", "[!CAUTION]"}
		for _, alert := range alerts {
			if strings.HasPrefix(trimmed, alert) {
				typeStr := strings.ToLower(strings.Trim(alert, "[!]"))
				bq.SetAttributeString("class", []byte("markdown-alert markdown-alert-"+typeStr))

				remaining := len(alert) + strings.Index(content, alert)
				if remaining < len(content) && (content[remaining] == '\n' || content[remaining] == ' ') {
					remaining++
				}

				for c := p.FirstChild(); c != nil && remaining > 0; {
					next := c.NextSibling()
					if t, ok := c.(*ast.Text); ok {
						valLen := len(t.Segment.Value(reader.Source()))
						if valLen <= remaining {
							remaining -= valLen
							p.RemoveChild(p, c)
						} else {
							t.Segment = text.NewSegment(t.Segment.Start+remaining, t.Segment.Stop)
							remaining = 0
						}
					} else {
						break
					}
					c = next
				}

				return ast.WalkContinue, nil
			}
		}

		return ast.WalkContinue, nil
	})
}

// formatMetaValue formats frontmatter values into safe HTML representations.
func formatMetaValue(val any) string {
	switch v := val.(type) {
	case string:
		return htmlpkg.EscapeString(v)
	case []any:
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf(`<span class="frontmatter-tag">%s</span>`, htmlpkg.EscapeString(fmt.Sprint(item))))
		}
		return strings.Join(items, " ")
	case []string:
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf(`<span class="frontmatter-tag">%s</span>`, htmlpkg.EscapeString(item)))
		}
		return strings.Join(items, " ")
	case bool:
		if v {
			return `<span class="text-green-600 dark:text-green-400 font-semibold">true</span>`
		}
		return `<span class="text-red-600 dark:text-red-400 font-semibold">false</span>`
	case nil:
		return `<span class="opacity-50 italic">null</span>`
	default:
		return htmlpkg.EscapeString(fmt.Sprintf("%v", v))
	}
}

// renderFrontMatter converts extracted YAML metadata into a structured HTML block.
func renderFrontMatter(metaData map[string]any) string {
	if len(metaData) == 0 {
		return ""
	}

	keys := make([]string, 0, len(metaData))
	for k := range metaData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(`<details class="frontmatter-container not-prose mb-6 rounded-lg border text-xs overflow-hidden" open>`)
	sb.WriteString(`<summary class="cursor-pointer font-semibold px-4 py-2 select-none opacity-80 hover:opacity-100 flex items-center justify-between border-b">`)
	sb.WriteString(`<span>Metadata &amp; Properties</span>`)
	sb.WriteString(`<span class="text-[10px] opacity-60 font-mono">YAML</span>`)
	sb.WriteString(`</summary>`)
	sb.WriteString(`<div class="p-3 overflow-x-auto">`)
	sb.WriteString(`<table class="frontmatter-table w-full border-collapse">`)
	sb.WriteString(`<tbody>`)

	for _, k := range keys {
		v := metaData[k]
		sb.WriteString(`<tr>`)
		sb.WriteString(fmt.Sprintf(`<td class="font-medium text-slate-500 dark:text-slate-400 pr-4 py-1 align-top whitespace-nowrap w-24">%s</td>`, htmlpkg.EscapeString(k)))
		sb.WriteString(fmt.Sprintf(`<td class="text-slate-900 dark:text-slate-100 py-1">%s</td>`, formatMetaValue(v)))
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</tbody></table></div></details>`)
	return sb.String()
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
		return "", fmt.Errorf("failed to write chroma CSS: %w", err)
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
			meta.Meta,
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

	context := parser.NewContext()
	if err := md.Convert([]byte(input), &buf, parser.WithContext(context)); err != nil {
		return "", fmt.Errorf("goldmark conversion error: %w", err)
	}

	metaData := meta.Get(context)
	frontMatterHTML := renderFrontMatter(metaData)

	rawHTML := frontMatterHTML + buf.String()
	return r.p.Sanitize(rawHTML), nil
}