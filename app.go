package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"marksafe/internal/config"
	"marksafe/internal/filesystem"
	"marksafe/internal/markdown"
	"marksafe/internal/mdbrowse"

	"github.com/mlechner911/mdviewer/mlcext"
	"github.com/mlechner911/mdviewer/mlcext/mlcdoc"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// FileResult holds the outcome of a file open operation.
type FileResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// versionFile is the repository's VERSION file, embedded so every build
// (Taskfile, CI, remote macOS) reports the same version without -ldflags.
//
//go:embed VERSION
var versionFile string

var appVersion = strings.TrimSpace(versionFile)

// GetVersion returns the application version to the frontend.
func (a *App) GetVersion() string {
	return appVersion
}

// App struct defines the main application state and dependencies.
type App struct {
	window      *application.WebviewWindow
	renderer    *markdown.Renderer
	config      *config.ConfigManager
	initialFile string
}

// emit broadcasts an event to the frontend (v2: runtime.EventsEmit).
func emit(name string, data ...any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data...)
	}
}

// logger returns the Wails application logger, or the default logger before startup.
func logger() *slog.Logger {
	if app := application.Get(); app != nil && app.Logger != nil {
		return app.Logger
	}
	return slog.Default()
}

// label returns the translation for key, or fallback if it is missing.
func label(t map[string]string, key, fallback string) string {
	if s := t[key]; s != "" {
		return s
	}
	return fallback
}

// NewApp creates a new App application struct.
func NewApp() *App {
	cfg, _ := config.NewConfigManager()
	return &App{
		renderer: markdown.NewRenderer(),
		config:   cfg,
	}
}

// SetInitialFile stores the file path provided via CLI.
func (a *App) SetInitialFile(path string) {
	if abs, err := filepath.Abs(path); err == nil {
		a.initialFile = abs
	} else {
		a.initialFile = path
	}
}

// UpdateMenu dynamically updates the application menu with translated strings.
func (a *App) UpdateMenu(t map[string]string) {
	app := application.Get()
	if app == nil {
		return
	}
	appMenu := application.NewMenu()

	if runtime.GOOS == "darwin" {
		appMenu.AddRole(application.AppMenu)
	}

	// File Menu
	fileMenu := appMenu.AddSubmenu(t["menuFile"])
	fileMenu.Add(t["menuNewTab"]).SetAccelerator("CmdOrCtrl+t").OnClick(func(*application.Context) {
		a.MenuNewTab()
	})
	fileMenu.Add(t["menuNewFromTemplate"]).SetAccelerator("CmdOrCtrl+Shift+n").OnClick(func(*application.Context) {
		emit("menu-new-from-template")
	})
	fileMenu.Add(label(t, "menuOpenDocumentRoot", "Open Document Folder")).SetAccelerator("CmdOrCtrl+Shift+d").OnClick(func(*application.Context) {
		go a.OpenDocumentRoot()
	})
	fileMenu.Add(label(t, "menuChooseDocumentRoot", "Choose Document Folder…")).OnClick(func(*application.Context) {
		go func() { _, _ = a.ChooseDocumentRoot() }()
	})
	fileMenu.AddSeparator()
	fileMenu.Add(t["menuOpen"]).SetAccelerator("CmdOrCtrl+o").OnClick(func(*application.Context) {
		a.MenuOpenFile()
	})

	// Submenu: Recent Files
	recentMenu := fileMenu.AddSubmenu(t["menuRecentFiles"])
	recentFiles := a.config.GetRecentFiles()
	if len(recentFiles) == 0 {
		recentMenu.Add(t["menuNoRecentFiles"]).SetEnabled(false)
	} else {
		for _, path := range recentFiles {
			p := path // Capture for closure
			recentMenu.Add(filepath.Base(p)).OnClick(func(*application.Context) {
				emit("menu-open-recent", p)
			})
		}
	}

	fileMenu.Add(t["menuSave"]).SetAccelerator("CmdOrCtrl+s").OnClick(func(*application.Context) {
		a.MenuSaveFile()
	})
	fileMenu.Add(label(t, "menuSaveAs", "Save As...")).SetAccelerator("CmdOrCtrl+Shift+s").OnClick(func(*application.Context) {
		a.MenuSaveAsFile()
	})
	fileMenu.AddSeparator()
	fileMenu.Add(t["menuAbout"]).OnClick(func(*application.Context) {
		// Themed in-app dialog (frontend) instead of the native one,
		// so About follows the active Dark/Light mode.
		emit("menu-about")
	})

	// Format Menu
	formatMenu := appMenu.AddSubmenu(t["menuFormat"])
	formatMenu.Add(t["menuBold"]).SetAccelerator("CmdOrCtrl+b").OnClick(func(*application.Context) { emit("format-bold") })
	formatMenu.Add(t["menuItalic"]).SetAccelerator("CmdOrCtrl+i").OnClick(func(*application.Context) { emit("format-italic") })
	formatMenu.AddSeparator()
	formatMenu.Add("H1").SetAccelerator("CmdOrCtrl+1").OnClick(func(*application.Context) { emit("format-h1") })
	formatMenu.Add("H2").SetAccelerator("CmdOrCtrl+2").OnClick(func(*application.Context) { emit("format-h2") })
	formatMenu.Add("H3").SetAccelerator("CmdOrCtrl+3").OnClick(func(*application.Context) { emit("format-h3") })
	formatMenu.AddSeparator()
	formatMenu.Add(t["menuCodeBlock"]).SetAccelerator("CmdOrCtrl+Shift+c").OnClick(func(*application.Context) { emit("format-code") })

	// View Menu (Language & Appearance)
	viewMenu := appMenu.AddSubmenu(t["menuView"])

	// Submenu: Language (checkmark tracks the active locale; the menu
	// rebuilds on every locale change, so exactly one item is checked)
	langMenu := viewMenu.AddSubmenu(t["menuLanguage"])
	activeLocale := t["currentLocale"]
	for _, lang := range []struct{ id, label string }{
		{"en", "English"}, {"de", "Deutsch"}, {"es", "Espa\u00f1ol"}, {"fr", "Fran\u00e7ais"},
	} {
		lang := lang // capture for closure
		langMenu.Add(lang.label).
			OnClick(func(*application.Context) { emit("set-locale", lang.id) }).
			SetChecked(lang.id != "" && lang.id == activeLocale)
	}

	// Submenu: Appearance
	apprMenu := viewMenu.AddSubmenu(t["menuAppearance"])
	apprMenu.Add(t["menuThemeDark"]).OnClick(func(*application.Context) { emit("set-theme", "dark") })
	apprMenu.Add(t["menuThemeLight"]).OnClick(func(*application.Context) { emit("set-theme", "light") })
	apprMenu.Add(t["menuThemeAuto"]).OnClick(func(*application.Context) { emit("set-theme", "auto") })

	// Edit Menu: role items perform the native clipboard/undo actions on every platform.
	editMenu := appMenu.AddSubmenu(label(t, "menuEdit", "Edit"))
	editMenu.Append(application.NewMenuFromItems(
		application.NewUndoMenuItem().SetLabel(label(t, "menuUndo", "Undo")),
		application.NewRedoMenuItem().SetLabel(label(t, "menuRedo", "Redo")),
		application.NewMenuItemSeparator(),
		application.NewCutMenuItem().SetLabel(label(t, "menuCut", "Cut")),
		application.NewCopyMenuItem().SetLabel(label(t, "menuCopy", "Copy")),
		application.NewPasteMenuItem().SetLabel(label(t, "menuPaste", "Paste")),
		application.NewSelectAllMenuItem(),
	))

	// macOS has one application menu; Windows and Linux attach menus per window.
	if runtime.GOOS == "darwin" {
		app.Menu.Set(appMenu)
	} else if a.window != nil {
		a.window.SetMenu(appMenu)
	}
}

// SetWindowTitle dynamically updates the native OS application window title.
func (a *App) SetWindowTitle(title string) {
	if a.window != nil {
		a.window.SetTitle(title)
	}
}

// IsPathAllowed checks if a local file path is within a whitelisted directory.
func (a *App) IsPathAllowed(path string) bool {
	return a.config.IsPathAllowed(path)
}

// IsURLAllowed checks if a URL's domain is whitelisted.
func (a *App) IsURLAllowed(url string) bool {
	return a.config.IsURLAllowed(url)
}

// AddPathToWhitelist adds a directory to the whitelist.
func (a *App) AddPathToWhitelist(path string) error {
	return a.config.AddPath(path)
}

// AddURLToWhitelist adds a URL domain to the whitelist.
func (a *App) AddURLToWhitelist(url string) error {
	return a.config.AddURL(url)
}

// GetParentDir returns the absolute parent directory of a path.
func (a *App) GetParentDir(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Dir(path)
	}
	return filepath.Dir(abs)
}

// ResolveRelativePath resolves a relative path against a base directory.
func (a *App) ResolveRelativePath(baseDir string, relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(baseDir, relPath)
}

// ReadFile reads the content of a file given its path.
func (a *App) ReadFile(path string) (string, error) {
	logger().Info("ReadFile: Reading file", "path", path)
	return filesystem.ReadFile(path)
}

// GetInitialContent is called by the frontend on mount to check for CLI files.
func (a *App) GetInitialContent() *FileResult {
	if a.initialFile != "" {
		content, err := filesystem.ReadFile(a.initialFile)
		if err == nil {
			a.config.AddRecentFile(a.initialFile)
			return &FileResult{
				Path:    a.initialFile,
				Content: content,
			}
		} else if os.IsNotExist(err) {
			return &FileResult{
				Path:    a.initialFile,
				Content: "",
			}
		}
	}
	return nil
}

// AnalyzeMLC returns the MLC Markdown Extension analysis of md (a
// non-standard extension, docs/MLC-EXTENSIONS.md): diagnostics, field and
// hint positions and the shared lists, for the editor and the status bar.
// Plain Markdown gets an empty analysis (Detected false).
func (a *App) AnalyzeMLC(md string) mlcext.Analysis {
	return mlcext.Analyze(md)
}

// MLCDocPreview is a page preview rendered by mlcdoc: PNG data URLs per
// page, mlcdoc's warnings, or the message why mlcdoc cannot render it.
type MLCDocPreview struct {
	Pages    []string `json:"pages"`
	Warnings []string `json:"warnings"`
	Error    string   `json:"error"`
}

// MLCDocVersion returns the installed mlcdoc ("mlcdoc 0.16.0 (…)") or ""
// when there is none; the PDF preview is hidden then.
func (a *App) MLCDocVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return mlcdoc.Version(ctx, "")
}

// PreviewMLCDoc renders md with mlcdoc into page images, exactly as the
// PDF will look (MLC Markdown Extension, non-standard). docPath locates
// relative images; they are only read from a whitelisted directory.
func (a *App) PreviewMLCDoc(md string, docPath string) MLCDocPreview {
	baseDir := ""
	if docPath != "" {
		if abs, err := filepath.Abs(docPath); err == nil && a.config.IsPathAllowed(filepath.Dir(abs)) {
			baseDir = filepath.Dir(abs)
		}
	}
	// Render what the Sensor Monitor stores: without the upload front
	// matter, which mlcdoc rejects.
	res, err := mlcdoc.Preview(context.Background(), mlcdoc.Request{Markdown: mlcext.StoredText(md), BaseDir: baseDir})
	if err != nil {
		return MLCDocPreview{Pages: []string{}, Warnings: []string{}, Error: err.Error()}
	}
	out := MLCDocPreview{Pages: make([]string, len(res.Pages)), Warnings: res.Warnings}
	if out.Warnings == nil {
		out.Warnings = []string{}
	}
	for i, p := range res.Pages {
		out.Pages[i] = "data:image/png;base64," + base64.StdEncoding.EncodeToString(p)
	}
	return out
}

// MLCTemplates lists the templates for new documents of the MLC Markdown
// Extension: the built-in ones (SOP, VA, form, info), then the user's own
// from <document folder>/_vorlagen.
func (a *App) MLCTemplates() []mlcext.Template {
	list := mlcext.Templates()
	if dir := a.userTemplatesDir(); dir != "" {
		list = append(list, mlcext.TemplatesInDir(dir)...)
	}
	return list
}

func (a *App) userTemplatesDir() string {
	if root := a.documentRoot(); root != "" {
		return filepath.Join(root, mdbrowse.TemplatesDir)
	}
	return ""
}

// NewMLCDocument creates the Markdown of a new document from a template.
// A user template must lie in the document folder's _vorlagen.
func (a *App) NewMLCDocument(p mlcext.NewParams) (string, error) {
	if p.Template != "" {
		dir := a.userTemplatesDir()
		abs, err := filepath.Abs(p.Template)
		if dir == "" || err != nil || filepath.Dir(abs) != filepath.Clean(dir) {
			return "", fmt.Errorf("Vorlage liegt nicht im Vorlagenordner: %s", p.Template)
		}
	}
	return mlcext.NewDocument(p)
}

// SuggestMLCRefNo proposes the next RefNo for kind and area from the
// document folder, or else the documents next to docPath (only whitelisted
// directories are read). It is a proposal: only the Sensor Monitor knows
// which RefNos are taken.
func (a *App) SuggestMLCRefNo(kind, area, docPath string) string {
	dir := a.documentRoot()
	if dir == "" && docPath != "" {
		if abs, err := filepath.Abs(docPath); err == nil && a.config.IsPathAllowed(filepath.Dir(abs)) {
			dir = filepath.Dir(abs)
		}
	}
	var existing []string
	if dir != "" {
		existing = mlcext.RefNosInDir(dir, 5000)
	}
	return mlcext.SuggestRefNo(kind, area, existing)
}

// documentRoot is the configured document folder if it still exists and
// is whitelisted, else "".
func (a *App) documentRoot() string {
	dir := a.config.GetDocumentRoot()
	if dir == "" || !a.config.IsPathAllowed(dir) {
		return ""
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return ""
	}
	return dir
}

// GetDocumentRoot returns the document folder (the base folder of the
// document collection), "" when none is set.
func (a *App) GetDocumentRoot() string {
	return a.documentRoot()
}

// ChooseDocumentRoot lets the user pick the document folder; it is
// whitelisted and stored. Returns "" when cancelled.
func (a *App) ChooseDocumentRoot() (string, error) {
	dlg := application.Get().Dialog.OpenFile().
		SetTitle("Dokumentenordner wählen").
		CanChooseDirectories(true).
		CanChooseFiles(false)
	if cur := a.documentRoot(); cur != "" {
		dlg.SetDirectory(cur)
	}
	dir, err := dlg.PromptForSingleSelection()
	if err != nil || dir == "" {
		return "", err
	}
	if err := a.config.AddPath(dir); err != nil {
		return "", err
	}
	if err := a.config.SetDocumentRoot(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// OpenDocumentRoot opens the document folder in the browse window, asking
// for it first when none is set.
func (a *App) OpenDocumentRoot() {
	dir := a.documentRoot()
	if dir == "" {
		var err error
		if dir, err = a.ChooseDocumentRoot(); err != nil || dir == "" {
			return
		}
	}
	a.OpenBrowseWindow(dir, "")
}

// RenderMarkdown converts markdown string to sanitized HTML.
func (a *App) RenderMarkdown(input string, theme string) string {
	logger().Debug("Request: RenderMarkdown", "theme", theme)
	html, err := a.renderer.Render(input, theme)
	if err != nil {
		logger().Error("Failed to render markdown", "error", err)
		return fmt.Sprintf("<p>Error rendering markdown: %v</p>", err)
	}
	return html
}

// GetStyleCSS returns the raw CSS for a specific syntax highlighting style.
func (a *App) GetStyleCSS(style string) string {
	css, err := a.renderer.GetStyleCSS(style)
	if err != nil {
		logger().Error("Failed to get CSS for style", "style", style, "error", err)
		return ""
	}
	return css
}

// MenuOpenFile is called from the native application menu.
func (a *App) MenuOpenFile() {
	emit("menu-open-file")
}

// MenuSaveFile is called from the native application menu.
func (a *App) MenuSaveFile() {
	emit("menu-save-file")
}

// MenuSaveAsFile is called from the native application menu.
func (a *App) MenuSaveAsFile() {
	emit("menu-save-file-as")
}

// MenuNewTab is called from the native application menu.
func (a *App) MenuNewTab() {
	emit("menu-new-tab")
}

// OpenFile opens a native file dialog and returns the path and content.
func (a *App) OpenFile() (*FileResult, error) {
	path, err := application.Get().Dialog.OpenFile().
		SetTitle("Open Markdown File").
		AddFilter("Markdown Files (*.md)", "*.md").
		AddFilter("Text Files (*.txt)", "*.txt").
		AddFilter("All Files (*.*)", "*.*").
		PromptForSingleSelection()
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("user cancelled selection")
	}
	absPath, _ := filepath.Abs(path)
	content, err := filesystem.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	a.config.AddRecentFile(absPath)
	return &FileResult{
		Path:    absPath,
		Content: content,
	}, nil
}

// GetFileTitle extracts the base name from a file path.
func (a *App) GetFileTitle(path string) string {
	return filepath.Base(path)
}

// SaveFile saves content directly to the given path, or prompts if path is empty.
func (a *App) SaveFile(path string, content string) (string, error) {
	if path == "" {
		return a.SaveFileAs("document.md", content)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err := filesystem.WriteFile(absPath, content); err != nil {
		return "", err
	}
	a.config.AddRecentFile(absPath)
	return absPath, nil
}

// SaveFileAs opens a native save dialog and saves content to the chosen path.
func (a *App) SaveFileAs(defaultFilename string, content string) (string, error) {
	// A document without a folder yet (e.g. new from a template) starts
	// in the document folder.
	startDir := ""
	if filepath.Base(defaultFilename) == defaultFilename {
		startDir = a.documentRoot()
	}
	if defaultFilename == "" {
		defaultFilename = "document.md"
	} else {
		defaultFilename = filepath.Base(defaultFilename)
	}
	if filepath.Ext(defaultFilename) == "" {
		defaultFilename += ".md"
	}
	dlg := application.Get().Dialog.SaveFile().
		SetMessage("Save Markdown File").
		SetFilename(defaultFilename).
		AddFilter("Markdown Files (*.md)", "*.md").
		AddFilter("All Files (*.*)", "*.*")
	if startDir != "" {
		dlg.SetDirectory(startDir)
	}
	selectedPath, err := dlg.PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if selectedPath == "" {
		return "", nil
	}
	absPath, _ := filepath.Abs(selectedPath)
	if err := filesystem.WriteFile(absPath, content); err != nil {
		return "", err
	}
	a.config.AddRecentFile(absPath)
	return absPath, nil
}

// ExportHTML saves the rendered markdown as a standalone HTML file.
func (a *App) ExportHTML(htmlContent string, cssContent string) (string, error) {
	path, err := application.Get().Dialog.SaveFile().
		SetMessage("Export to HTML").
		SetFilename("exported.html").
		AddFilter("HTML Files (*.html)", "*.html").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}

	fullHTML := a.renderer.StandaloneHTML(htmlContent, cssContent)

	err = filesystem.WriteFile(path, fullHTML)
	if err != nil {
		return "", err
	}
	return path, nil
}
