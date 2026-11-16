package mlcext

import (
	"path/filepath"
	"strings"
)

// ExtMLC is the file extension of a document of the MLC Markdown Extension.
// The content is UTF-8 Markdown, byte for byte the same as a ".md" file;
// the extension only tells the system to open it with a reader of the
// extension (MarkSafe) instead of a plain Markdown viewer. A signature
// covers the content, never the file name, so renaming keeps it valid.
const ExtMLC = ".mlcmd"

// markdownExts are the extensions read as Markdown documents.
var markdownExts = []string{".md", ".markdown", ".mdown", ExtMLC}

// IsMarkdownFile reports whether name has a Markdown extension (".md",
// ".markdown", ".mdown" or ".mlcmd"), regardless of case.
func IsMarkdownFile(name string) bool {
	ext := filepath.Ext(name)
	for _, e := range markdownExts {
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

// TrimMarkdownExt is name without its Markdown extension; other names are
// returned unchanged.
func TrimMarkdownExt(name string) string {
	if IsMarkdownFile(name) {
		return name[:len(name)-len(filepath.Ext(name))]
	}
	return name
}

// SaveExt is the extension a new file for md should get: ".mlcmd" for a
// document of the extension (see Detect), else ".md". An existing file
// keeps its extension; this is only the suggestion for a new name.
func SaveExt(md string) string {
	if Detect(md) {
		return ExtMLC
	}
	return ".md"
}
