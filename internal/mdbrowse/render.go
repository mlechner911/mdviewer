package mdbrowse

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// RenderDoc reads a root-relative document and returns its title plus
// rendered HTML (frontmatter stripped, pasted-markdown defects repaired,
// internal links validated). The HTML fragment is wrapped in
// <div class="markdown"> for styling.
func (r *Root) RenderDoc(relPath string) (title, html string, err error) {
	content, absPath, err := r.ReadDoc(relPath)
	if err != nil {
		return "", "", err
	}
	title, _ = ExtractTitle(absPath)
	html = renderMarkdownToHTML(content)
	html = r.rewriteMarkdownLinks(html, filepath.ToSlash(relPath))
	return title, html, nil
}

// renderMarkdownToHTML converts markdown to HTML using goldmark.
func renderMarkdownToHTML(content string) string {
	content = stripFrontmatter(content)
	// Repair glued block markers from pasted/AI-generated sources
	// (normalize.go); well-formed documents pass through unchanged.
	content = normalizePastedMarkdown(content)
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Table, extension.Strikethrough, extension.TaskList),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return "<div class=\"markdown\"><p>Error: " + escapeHTML(err.Error()) + "</p></div>"
	}
	return "<div class=\"markdown\">" + buf.String() + "</div>"
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
