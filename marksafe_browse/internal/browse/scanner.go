package browse

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Scanner scans a directory and builds a TOC structure.
type Scanner struct {
	Root string // absolute path of the root directory
}

// NewScanner creates a new Scanner for the given root directory.
func NewScanner(root string) (*Scanner, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path %s: %w", root, err)
	}

	// Verify root exists and is a directory
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("root directory does not exist: %s", absRoot)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory: %s", absRoot)
	}

	return &Scanner{Root: absRoot}, nil
}

// Scan recursively scans the directory and returns the TOC structure.
func (s *Scanner) Scan() ([]TOCEntry, error) {
	entries, err := s.scanDir(s.Root)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// scanDir recursively scans a directory and returns TOC entries.
func (s *Scanner) scanDir(dir string) ([]TOCEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var entries []TOCEntry

	// Separate files and subdirectories
	var mdFiles []string
	var subdirs []string

	for _, f := range files {
		name := f.Name()
		// Skip hidden files and directories
		if strings.HasPrefix(name, ".") {
			continue
		}
		fullPath := filepath.Join(dir, name)
		if f.IsDir() {
			subdirs = append(subdirs, fullPath)
		} else if strings.HasSuffix(name, ".md") {
			mdFiles = append(mdFiles, fullPath)
		}
	}

	// Sort directories first, then files (both alphabetically)
	sort.Strings(subdirs)
	sort.Strings(mdFiles)

	// Process subdirectories recursively
	for _, subdir := range subdirs {
		subEntries, err := s.scanDir(subdir)
		if err != nil {
			return nil, err
		}
		if len(subEntries) > 0 {
			// Get directory name as title
			dirName := filepath.Base(subdir)
			entries = append(entries, TOCEntry{
				Path:     dirName,
				Title:    FormatTitle(dirName),
				Children: subEntries,
				IsDir:    true,
			})
		}
	}

	// Process markdown files (skip index.md at this level — handled separately)
	for _, mdPath := range mdFiles {
		name := filepath.Base(mdPath)
		if name == "index.md" {
			continue // index.md handled as landing page
		}
		title, err := ExtractTitle(mdPath)
		if err != nil {
			title = FormatTitle(strings.TrimSuffix(name, ".md"))
		}
		relPath, _ := RelativePath(s.Root, mdPath)

		entries = append(entries, TOCEntry{
			Path:  relPath,
			Title: title,
			IsDir: false,
		})
	}

	return entries, nil
}

// FindIndex checks if an index.md exists at the root.
// Returns its absolute path if found, nil otherwise.
func (s *Scanner) FindIndex() (string, bool) {
	indexPath := filepath.Join(s.Root, "index.md")
	if _, err := os.Stat(indexPath); err == nil {
		return indexPath, true
	}
	return "", false
}

// GetMarkdownFiles returns all .md files in the directory tree.
func (s *Scanner) GetMarkdownFiles() ([]string, error) {
	var files []string
	err := filepath.WalkDir(s.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// CountMarkdownFiles returns the total number of .md files (excluding index.md).
func (s *Scanner) CountMarkdownFiles() (int, error) {
	files, err := s.GetMarkdownFiles()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, f := range files {
		if filepath.Base(f) != "index.md" {
			count++
		}
	}
	return count, nil
}
