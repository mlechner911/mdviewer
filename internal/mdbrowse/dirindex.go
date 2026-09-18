package mdbrowse

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// RenderDir renders a directory: if it holds an index.md that page is
// served, otherwise a virtual directory listing in tree style. chromaStyle
// selects code colors, like RenderDoc.
func (r *Root) RenderDir(relPath, chromaStyle string) (title, html string, err error) {
	relSlash := filepath.ToSlash(relPath)
	abs, err := r.Resolve(relSlash)
	if err != nil {
		return "", "", err
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return "", "", fmt.Errorf("not a directory: %s", relSlash)
	}

	// Prefer a real index page when present.
	if indexRel := path.Join(relSlash, "index.md"); fileExists(abs, "index.md") {
		return r.RenderDoc(indexRel, chromaStyle)
	}

	title = dirDisplayTitle(relSlash)
	entries, err := r.scanDir(abs)
	if err != nil {
		return "", "", err
	}

	var b strings.Builder
	b.WriteString(`<div class="markdown">`)
	b.WriteString("<h1>" + escapeHTML(title) + "</h1>")
	if len(entries) == 0 {
		b.WriteString("<p>Keine Markdown-Dateien in diesem Verzeichnis.</p>")
	} else {
		b.WriteString(renderVDir(entries))
	}
	b.WriteString("</div>")
	return title, b.String(), nil
}

// renderVDir renders entries as a nested list mirroring the sidebar tree.
// Files and directories both carry data-md, so the client navigates
// in-app (file -> page, directory -> index or deeper listing).
func renderVDir(entries []TOCEntry) string {
	var b strings.Builder
	b.WriteString(`<ul class="vdir">`)
	for _, e := range entries {
		if e.IsDir {
			b.WriteString(`<li class="vdir-dir"><details open><summary><a data-md="`)
			b.WriteString(escapeHTML(e.Path))
			b.WriteString(`">`)
			b.WriteString(escapeHTML(e.Title))
			b.WriteString("</a></summary>")
			b.WriteString(renderVDir(e.Children))
			b.WriteString("</details></li>")
		} else {
			b.WriteString(`<li class="vdir-file"><a data-md="`)
			b.WriteString(escapeHTML(e.Path))
			b.WriteString(`">`)
			b.WriteString(escapeHTML(e.Title))
			b.WriteString("</a></li>")
		}
	}
	b.WriteString("</ul>")
	return b.String()
}

// dirDisplayTitle derives a title from a root-relative directory path.
func dirDisplayTitle(relSlash string) string {
	relSlash = strings.Trim(relSlash, "/")
	if relSlash == "" || relSlash == "." {
		return "Übersicht"
	}
	return FormatTitle(path.Base(relSlash))
}

// fileExists reports whether name exists directly inside dir.
func fileExists(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}
