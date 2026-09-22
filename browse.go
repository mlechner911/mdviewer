package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"marksafe/internal/mdbrowse"
)

// Browse mode backend: directory browsing inside the Wails app, without
// an HTTP server. All logic lives in the shared internal/mdbrowse
// library (same implementation as the marksafe_browse tool); these
// methods are thin Wails bindings over the Root API.
//
// Security note: every method is confined to rootDir by the library's
// hard boundary (Resolve rejects escapes). Image/file delivery keeps
// using /local-resource with the regular whitelist — the frontend adds
// the browse root on mode entry (Phase 3 wiring).

// BrowseDoc is a rendered document for browse mode.
type BrowseDoc struct {
	Title    string `json:"title"`
	HTML     string `json:"html"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// BrowseTree returns the table of contents for rootDir.
func (a *App) BrowseTree(rootDir string) ([]mdbrowse.TOCEntry, error) {
	logger().Debug("Request: BrowseTree", "root", rootDir)
	root, err := mdbrowse.NewRoot(rootDir)
	if err != nil {
		return nil, err
	}
	return root.Scan()
}

// BrowseRender renders relPath under rootDir. Directories serve their
// index.md or a virtual listing, like the browse tool. chromaStyle
// follows the editor convention ("github-dark"/"github").
func (a *App) BrowseRender(rootDir, relPath, chromaStyle string) (BrowseDoc, error) {
	logger().Debug("Request: BrowseRender", "root", rootDir, "path", relPath)
	root, err := mdbrowse.NewRoot(rootDir)
	if err != nil {
		return BrowseDoc{}, err
	}
	abs, err := root.Resolve(relPath)
	if err != nil {
		return BrowseDoc{}, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return BrowseDoc{}, err
	}
	var title, html string
	var size int64
	var modified string
	switch {
	case info.IsDir():
		title, html, err = root.RenderDir(relPath, chromaStyle)
	case !strings.HasSuffix(strings.ToLower(relPath), ".md"):
		return BrowseDoc{}, fmt.Errorf("not a markdown file: %s", relPath)
	default:
		title, html, err = root.RenderDoc(relPath, chromaStyle)
		size = info.Size()
		modified = info.ModTime().UTC().Format(time.RFC3339)
	}
	if err != nil {
		logger().Error("Failed to render", "error", err)
		return BrowseDoc{}, err
	}
	return BrowseDoc{Title: title, HTML: html, Size: size, Modified: modified}, nil
}

// BrowseSearch runs a fuzzy full-text search over rootDir.
func (a *App) BrowseSearch(rootDir, query string) ([]mdbrowse.SearchResult, error) {
	logger().Debug("Request: BrowseSearch", "root", rootDir, "query", query)
	root, err := mdbrowse.NewRoot(rootDir)
	if err != nil {
		return nil, err
	}
	return root.Search(query)
}

// OpenBrowseWindow opens the standalone browse window for rootDir,
// starting at relPath. At most one browse window ever exists: an
// existing one is focused and navigated instead of opening another.
func (a *App) OpenBrowseWindow(rootDir, relPath string) {
	logger().Debug("Request: OpenBrowseWindow", "root", rootDir, "path", relPath)
	app := application.Get()
	if app == nil {
		return
	}
	target := "/?browse=" + url.QueryEscape(rootDir) + "&path=" + url.QueryEscape(relPath)
	title := "MarkSafe Browse \u2014 " + rootDir
	if w, ok := app.Window.GetByName("browse"); ok {
		w.SetURL(target)
		w.SetTitle(title)
		w.Focus()
		return
	}
	w := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "browse",
		Title:            "MarkSafe Browse \u2014 " + rootDir,
		Width:            1280,
		Height:           860,
		MinWidth:         900,
		MinHeight:        600,
		URL:              target,
		BackgroundColour: application.NewRGB(27, 38, 54),
		Windows: application.WindowsWindow{
			Theme: application.Dark,
		},
	})
	w.Show()
	w.Focus()
}
