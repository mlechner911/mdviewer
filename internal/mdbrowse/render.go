package mdbrowse

import (
	"path/filepath"
	"strings"

	"marksafe/internal/markdown"
)

// sharedRenderer is the single Markdown implementation for all frontends
// (browse server, Wails preview): Chroma highlighting, MathJax spans,
// GitHub alerts, sanitized HTML. No second implementation.
var sharedRenderer = markdown.NewRenderer()

// RenderDoc reads a root-relative document and returns its title plus
// rendered HTML (frontmatter stripped, pasted-markdown defects repaired,
// internal links validated). chromaStyle selects code colors, e.g.
// "github-dark" or "github" (empty = renderer default).
func (r *Root) RenderDoc(relPath, chromaStyle string) (title, html string, err error) {
	content, absPath, err := r.ReadDoc(relPath)
	if err != nil {
		return "", "", err
	}
	title, _ = ExtractTitle(absPath)
	html = renderMarkdownToHTML(content, chromaStyle)
	html = r.rewriteMarkdownLinks(html, filepath.ToSlash(relPath))
	return title, html, nil
}

// renderMarkdownToHTML converts markdown to sanitized HTML.
func renderMarkdownToHTML(content, chromaStyle string) string {
	content = stripFrontmatter(content)
	// Repair glued block markers from pasted/AI-generated sources
	// (normalize.go); well-formed documents pass through unchanged.
	content = normalizePastedMarkdown(content)
	html, err := sharedRenderer.Render(content, chromaStyle)
	if err != nil {
		return "<div class=\"markdown\"><p>Error rendering document</p></div>"
	}
	return "<div class=\"markdown\">" + html + "</div>"
}

// stripFrontmatter removes YAML frontmatter delimiters from markdown content.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	endIdx := strings.Index(content[3:], "---")
	if endIdx == -1 {
		return content
	}
	rest := content[3+endIdx+3:]
	return strings.TrimLeft(rest, "\n")
}

// escapeHTML provides basic HTML escaping.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
