// Package mdbrowse provides the shared Markdown directory-browsing library:
// rooted file access with a hard security boundary, TOC scanning, title
// extraction, full-text search, link validation and the render pipeline.
// It is UI-agnostic: both the marksafe_browse web server and (future)
// other frontends build on this Root API.
package mdbrowse

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Root is a rooted view of a Markdown directory. Dir is absolute.
// Nothing outside Dir is reachable through this API (harte Grenze).
type Root struct {
	Dir string
}

// NewRoot resolves dir to an absolute path and verifies it exists.
func NewRoot(dir string) (*Root, error) {
	absRoot, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path %s: %w", dir, err)
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("root directory does not exist: %s", absRoot)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory: %s", absRoot)
	}
	return &Root{Dir: absRoot}, nil
}

// Resolve maps a root-relative request path to an absolute path,
// rejecting anything that escapes the root.
func (r *Root) Resolve(requestedPath string) (string, error) {
	if filepath.IsAbs(requestedPath) {
		return "", fmt.Errorf("absolute paths not allowed: %s", requestedPath)
	}
	absRequested, err := filepath.Abs(filepath.Join(r.Dir, requestedPath))
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	rel, err := filepath.Rel(r.Dir, absRequested)
	if err != nil {
		return "", fmt.Errorf("path escape detected: %s is outside %s", requestedPath, r.Dir)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("security violation: %s escapes root %s", requestedPath, r.Dir)
	}

	// On Windows, do case-insensitive comparison
	if runtime.GOOS == "windows" {
		absRequested = strings.ToLower(absRequested)
		if !strings.HasPrefix(absRequested, strings.ToLower(r.Dir)) {
			return "", fmt.Errorf("security violation: %s is outside root %s", requestedPath, r.Dir)
		}
		return absRequested, nil
	}

	if !strings.HasPrefix(absRequested, r.Dir) {
		return "", fmt.Errorf("security violation: %s is outside root %s", requestedPath, r.Dir)
	}

	return absRequested, nil
}

// IsWithinRoot reports whether absPath lies inside the root directory.
func (r *Root) IsWithinRoot(absPath string) bool {
	rel, err := filepath.Rel(r.Dir, absPath)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		rel = strings.ToLower(rel)
	}
	return !strings.HasPrefix(rel, "..")
}

// ReadDoc reads a root-relative document and returns content + abs path.
func (r *Root) ReadDoc(relPath string) (content, absPath string, err error) {
	absPath, err = r.Resolve(relPath)
	if err != nil {
		return "", "", err
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", "", err
	}
	return string(data), absPath, nil
}

// Scan recursively scans the directory and returns the TOC structure.
func (r *Root) Scan() ([]TOCEntry, error) {
	return r.scanDir(r.Dir)
}

// scanDir recursively scans a directory and returns TOC entries.
func (r *Root) scanDir(dir string) ([]TOCEntry, error) {
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
		subEntries, err := r.scanDir(subdir)
		if err != nil {
			return nil, err
		}
		if len(subEntries) > 0 {
			// Root-relative path (matters for nesting: "docs/api",
			// not just "api"); title from the plain directory name.
			relDir, _ := RelativePath(r.Dir, subdir)
			entries = append(entries, TOCEntry{
				Path:     relDir,
				Title:    FormatTitle(filepath.Base(subdir)),
				Children: subEntries,
				IsDir:    true,
			})
		}
	}

	// Process markdown files (skip index.md at this level — landing page)
	for _, mdPath := range mdFiles {
		name := filepath.Base(mdPath)
		if name == "index.md" {
			continue
		}
		title, err := ExtractTitle(mdPath)
		if err != nil {
			title = FormatTitle(strings.TrimSuffix(name, ".md"))
		}
		relPath, _ := RelativePath(r.Dir, mdPath)

		entries = append(entries, TOCEntry{
			Path:  relPath,
			Title: title,
			IsDir: false,
		})
	}

	return entries, nil
}

// FindIndex checks if an index.md exists at the root.
func (r *Root) FindIndex() (string, bool) {
	indexPath := filepath.Join(r.Dir, "index.md")
	if _, err := os.Stat(indexPath); err == nil {
		return indexPath, true
	}
	return "", false
}

// GetMarkdownFiles returns all .md files in the directory tree.
func (r *Root) GetMarkdownFiles() ([]string, error) {
	var files []string
	err := filepath.WalkDir(r.Dir, func(path string, d os.DirEntry, err error) error {
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
func (r *Root) CountMarkdownFiles() (int, error) {
	files, err := r.GetMarkdownFiles()
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
