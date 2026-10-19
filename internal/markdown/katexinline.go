package markdown

import (
	"embed"
	"encoding/base64"
	"regexp"
	"strings"
	"sync"
)

// KaTeX fonts (woff2 only) for offline HTML exports. Copied from
// frontend/node_modules/katex/dist/fonts; keep in sync with
// katex.min.css.txt when KaTeX is upgraded.
//
//go:embed katexfonts/*.woff2
var katexFonts embed.FS

// fontSrcRe matches a whole @font-face src list whose first entry is a
// woff2 file, e.g. src:url(fonts/KaTeX_AMS-Regular.woff2) format("woff2"),…
var fontSrcRe = regexp.MustCompile(`src:url\(fonts/([A-Za-z0-9_-]+\.woff2)\) format\("woff2"\)[^;}]*`)

var (
	katexInlineOnce sync.Once
	katexInlineCSS  string
)

// GetKatexCSSInline returns the KaTeX CSS with every font embedded as a
// woff2 data: URI, so formulas render in a standalone file without
// network or font files next to it (~400 KB; only include it when the
// document contains math).
func (r *Renderer) GetKatexCSSInline() string {
	katexInlineOnce.Do(func() {
		katexInlineCSS = fontSrcRe.ReplaceAllStringFunc(katexCSS, func(m string) string {
			name := fontSrcRe.FindStringSubmatch(m)[1]
			data, err := katexFonts.ReadFile("katexfonts/" + name)
			if err != nil {
				return m
			}
			return `src:url(data:font/woff2;base64,` + base64.StdEncoding.EncodeToString(data) + `) format("woff2")`
		})
	})
	return katexInlineCSS
}

// StandaloneHTML wraps a rendered document (the preview's <article
// class="md-doc">) and its stylesheet into a self-contained HTML file.
// Formulas get the KaTeX stylesheet with embedded fonts, but only when
// the document actually contains math.
func (r *Renderer) StandaloneHTML(body, css string) string {
	katex := ""
	if strings.Contains(body, `class="katex`) {
		katex = "<style>\n" + r.GetKatexCSSInline() + "\n</style>\n"
	}
	return "<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"UTF-8\">\n" +
		"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n" +
		"<title>MarkSafe Export</title>\n" + katex +
		"<style>\n" + css + "\n</style>\n</head>\n<body>\n" + body + "\n</body>\n</html>\n"
}
