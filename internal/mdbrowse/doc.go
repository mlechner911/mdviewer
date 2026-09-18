package mdbrowse

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// TOCEntry represents a single entry in the table of contents.
// JSON tags are lowercase to match the Svelte frontend contract
// (Sidebar.svelte, App.svelte, lib/backend.ts).
type TOCEntry struct {
	Path     string     `json:"path"`     // relative to root
	Title    string     `json:"title"`    // from frontmatter or first H1
	Children []TOCEntry `json:"children"` // sub-entries (for directory nesting)
	IsDir    bool       `json:"isDir"`
}

// PageData holds the data for rendering a single page.
type PageData struct {
	Path     string // relative to root
	Title    string
	Content  string // raw markdown content
	FullPath string // absolute path
}

// ExtractTitle reads a markdown file and extracts its title.
// Priority: YAML frontmatter "title:" > first H1 heading > filename (without extension).
func ExtractTitle(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	inFrontmatter := false
	frontmatterTitle := ""

	titleRegex := regexp.MustCompile(`^#\s+(.+)$`)
	fmTitleRegex := regexp.MustCompile(`^title:\s*(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// Track frontmatter
		if line == "---" {
			if !inFrontmatter && lineCount == 1 {
				inFrontmatter = true
				continue
			}
			if inFrontmatter && lineCount > 1 {
				inFrontmatter = false
				continue
			}
		}

		// Inside frontmatter, look for title
		if inFrontmatter {
			matches := fmTitleRegex.FindStringSubmatch(line)
			if len(matches) == 2 {
				frontmatterTitle = strings.TrimSpace(matches[1])
				continue
			}
		}

		// First H1 after frontmatter
		matches := titleRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			return truncateTitle(matches[1]), nil
		}

		// Stop after first 30 lines if no H1 found and we have a frontmatter title
		if lineCount > 30 && frontmatterTitle != "" {
			return truncateTitle(frontmatterTitle), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if frontmatterTitle != "" {
		return truncateTitle(frontmatterTitle), nil
	}

	// Fallback: use filename
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return truncateTitle(strings.ReplaceAll(name, "-", " ")), nil
}

// maxTitleRunes caps sidebar titles: a source line that glues whole
// paragraphs into one "heading" must not blow up the TOC.
const maxTitleRunes = 120

// truncateTitle shortens overlong titles rune-aware, appending "…".
func truncateTitle(s string) string {
	// A glued "## subheading" inside one source line never belongs to
	// the title ("Backups## 1. ..."): cut it before truncating.
	if i := strings.Index(s, "##"); i != -1 {
		s = s[:i]
	}
	r := []rune(strings.TrimSpace(s))
	if len(r) == 0 {
		return strings.TrimSpace(s)
	}
	if len(r) <= maxTitleRunes {
		return string(r)
	}
	return string(r[:maxTitleRunes]) + "\u2026"
}

// FormatTitle converts a filename or heading into a display title.
func FormatTitle(title string) string {
	title = strings.ReplaceAll(title, "-", " ")
	words := strings.Fields(title)
	for i, w := range words {
		if w == "" {
			continue
		}
		// Rune-aware: w[:1] would split multi-byte initials (Übersicht).
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

// formatSlug creates a URL-safe slug from a title.
func formatSlug(title string) string {
	slug := strings.ToLower(strings.TrimSpace(title))
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

// SanitizeFilename creates a safe filename from a title.
func SanitizeFilename(title string) string {
	return formatSlug(title) + ".md"
}

// RelativePath returns a path relative to root.
func RelativePath(root, fullPath string) (string, error) {
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// String returns a human-readable path representation.
func (e *TOCEntry) String() string {
	prefix := ""
	if e.IsDir {
		prefix = "📁 "
	}
	return fmt.Sprintf("%s%s (%s)", prefix, e.Title, e.Path)
}
