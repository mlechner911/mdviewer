package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"marksafe/internal/config"
	"marksafe/internal/filesystem"
	"marksafe/internal/markdown"
	"marksafe/internal/mdbrowse"

	"github.com/mlechner911/mdviewer/mlcext"
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
	pdf         pdfTool
	templates   templateSeed
	sensor      sensorState
	closing     closeState
}

// templateSeed copies the built-in templates into the user's template
// folder once per start (mlcext.SeedTemplates).
type templateSeed struct {
	once sync.Once
	ok   bool
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
	// PDF functions need mlcdoc; without it (e.g. the bare Linux binary)
	// they are not offered at all.
	hasPDF := a.MLCDocVersion() != ""
	// Print goes through the PDF (system PDF viewer), like the export.
	// Outside macOS Ctrl+P is handled by the page (as the edit keys, see
	// editItem): through the menu WebView2 would open its own print dialog
	// as well.
	if hasPDF {
		fileMenu.Append(application.NewMenuFromItems(editItem(application.NewMenuItem(label(t, "menuPrint", "Print…")).
			SetAccelerator("CmdOrCtrl+p").OnClick(func(*application.Context) { emit("menu-print") }), "")))
	}
	// Submenu: Export (PDF from mlcdoc, HTML from the preview)
	exportMenu := fileMenu.AddSubmenu(label(t, "menuExport", "Export"))
	if hasPDF {
		exportMenu.Add(label(t, "menuExportPdf", "As PDF…")).SetAccelerator("CmdOrCtrl+Shift+e").OnClick(func(*application.Context) {
			emit("menu-export-pdf")
		})
	}
	exportMenu.Add(label(t, "menuExportHtml", "As HTML…")).OnClick(func(*application.Context) {
		emit("menu-export-html")
	})
	// Submenu: Sensor Monitor, only with an address in the settings: for
	// everyone else MarkSafe shows nothing of it.
	if a.config.GetSensorMonitorURL() != "" {
		smMenu := fileMenu.AddSubmenu(label(t, "menuSensorMonitor", "Sensor Monitor"))
		smMenu.Add(label(t, "menuSensorOpen", "Open…")).OnClick(func(*application.Context) { emit("menu-sm-open") })
		smMenu.Add(label(t, "menuSensorSaveDraft", "Save as Draft…")).OnClick(func(*application.Context) { emit("menu-sm-save-draft") })
		smMenu.AddSeparator()
		if a.sensorLogin() == nil {
			smMenu.Add(label(t, "menuSensorLogin", "Sign In…")).OnClick(func(*application.Context) { emit("menu-sm-login") })
		} else {
			smMenu.Add(label(t, "menuSensorLogout", "Sign Out")).OnClick(func(*application.Context) { a.SensorMonitorLogout() })
		}
	}
	fileMenu.AddSeparator()
	fileMenu.Add(label(t, "menuSettings", "Settings…")).SetAccelerator("CmdOrCtrl+,").OnClick(func(*application.Context) {
		emit("menu-settings")
	})
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
		editItem(application.NewUndoMenuItem(), label(t, "menuUndo", "Undo")),
		editItem(application.NewRedoMenuItem(), label(t, "menuRedo", "Redo")),
		application.NewMenuItemSeparator(),
		editItem(application.NewCutMenuItem(), label(t, "menuCut", "Cut")),
		editItem(application.NewCopyMenuItem(), label(t, "menuCopy", "Copy")),
		editItem(application.NewPasteMenuItem(), label(t, "menuPaste", "Paste")),
		editItem(application.NewSelectAllMenuItem(), ""),
	))

	// macOS has one application menu; Windows and Linux attach menus per window.
	if runtime.GOOS == "darwin" {
		app.Menu.Set(appMenu)
	} else if a.window != nil {
		a.window.SetMenu(appMenu)
	}
}

// editItem labels an edit role item. Outside macOS the keyboard shortcut
// stays with the webview: Wails (v3 beta.17, Windows) runs the menu action
// for Ctrl+V and then hands the key on to WebView2 as well, so every paste
// (and undo, cut …) happened twice. The menu still shows the shortcut and
// works by click.
func editItem(item *application.MenuItem, text string) *application.MenuItem {
	if text != "" {
		item.SetLabel(text)
	}
	if runtime.GOOS == "darwin" {
		return item
	}
	acc := item.GetAccelerator()
	item.RemoveAccelerator()
	if runtime.GOOS == "windows" && acc != "" {
		item.SetLabel(item.Label() + "\t" + acc)
	}
	return item
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

// MLCMoveHeader is the quick fix "Kopf ins Front Matter übernehmen": the
// legacy header lines become doc_id, version and title. "" when there is
// nothing to move.
func (a *App) MLCMoveHeader(md string) string {
	out, ok := mlcext.MoveHeaderToFrontMatter(md)
	if !ok {
		return ""
	}
	return out
}

// MLCFitTable sets the column widths of the table at line (1-based):
// mode "auto" (mlcdoc sizes by content) or "content" (full width, shares
// after the content). "" when line is in no table.
func (a *App) MLCFitTable(md string, line int, mode string) string {
	out, ok := mlcext.FitTableColumns(md, line, mode)
	if !ok {
		return ""
	}
	return out
}

// MLCSignatureState tells whether md is signed and still matches its
// checksum (an edit voids the signature).
func (a *App) MLCSignatureState(md string) mlcext.SignatureState {
	return mlcext.CheckSignature(md)
}

// MLCRemoveSignature returns md without its signature block.
func (a *App) MLCRemoveSignature(md string) string {
	return mlcext.RemoveSignature(md)
}

// MLCTemplates lists the templates for new documents of the MLC Markdown
// Extension: the built-in starting set, replaced or extended by the
// schemas (*.json + Markdown body) in <document folder>/_vorlagen. New
// kinds or fields need a new file there, not a new program.
func (a *App) MLCTemplates() []mlcext.TemplateInfo {
	docs := a.documentRefs()
	list := a.loadTemplates()
	for i := range list {
		list[i] = list[i].WithDocuments(docs)
	}
	return list
}

// documentRefs are the documents of the document folder that docref
// fields can point to.
func (a *App) documentRefs() []mlcext.DocumentRef {
	root := a.documentRoot()
	if root == "" {
		return nil
	}
	r, err := mdbrowse.NewRoot(root)
	if err != nil {
		return nil
	}
	docs, err := r.Documents()
	if err != nil {
		return nil
	}
	refs := make([]mlcext.DocumentRef, 0, len(docs))
	for _, d := range docs {
		refs = append(refs, mlcext.DocumentRef{RefNo: d.RefNo, Title: d.Title})
	}
	return refs
}

// appTemplatesDir is the user's own template folder
// (<user config>/marksafe/vorlagen): the built-in templates are copied
// there once and can be changed without a new release.
func (a *App) appTemplatesDir() string {
	if a.config == nil {
		return ""
	}
	return filepath.Join(a.config.Dir(), "vorlagen")
}

// loadTemplates: the user's template folder (seeded from the built-in
// set; without it the built-in set itself), then _vorlagen in the document
// folder. A later template with the same id replaces an earlier one.
func (a *App) loadTemplates() []mlcext.TemplateInfo {
	dir := a.appTemplatesDir()
	a.templates.once.Do(func() {
		if dir == "" {
			return
		}
		written, err := mlcext.SeedTemplates(dir)
		if err != nil {
			logger().Error("Seeding templates failed", "dir", dir, "error", err)
			return
		}
		if len(written) > 0 {
			logger().Info("Templates copied to the user folder", "dir", dir, "files", written)
		}
		a.templates.ok = true
	})
	if !a.templates.ok {
		return mlcext.LoadTemplates(a.userTemplatesDir())
	}
	return mlcext.LoadTemplatesFrom(false, dir, a.userTemplatesDir())
}

// OpenTemplatesFolder shows the user's template folder in the file
// manager (created and seeded if needed).
func (a *App) OpenTemplatesFolder() error {
	a.loadTemplates()
	dir := a.appTemplatesDir()
	if dir == "" {
		return fmt.Errorf("no template folder")
	}
	return application.Get().Env.OpenFileManager(dir, false)
}

// TemplatesFolder is the path of the user's template folder, for display.
func (a *App) TemplatesFolder() string {
	return a.appTemplatesDir()
}

func (a *App) userTemplatesDir() string {
	if root := a.documentRoot(); root != "" {
		return filepath.Join(root, mdbrowse.TemplatesDir)
	}
	return ""
}

// RenderMLCTemplate fills template id with the form values and checks the
// result (field errors, diagnostics of the new document).
func (a *App) RenderMLCTemplate(id string, values map[string]any) mlcext.RenderResult {
	for _, t := range a.loadTemplates() {
		if t.ID == id {
			return t.WithDocuments(a.documentRefs()).Render(values)
		}
	}
	return mlcext.RenderResult{FieldErrors: map[string]string{"": "Vorlage nicht gefunden: " + id}, Diagnostics: []mlcext.Diagnostic{}}
}

// SuggestMLCRefNo proposes the next RefNo for prefix and area from the
// document folder, or else the documents next to docPath (only whitelisted
// directories are read). It is a proposal: only the Sensor Monitor knows
// which RefNos are taken.
func (a *App) SuggestMLCRefNo(prefix, area, docPath string) string {
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
	return mlcext.SuggestRefNo(prefix, area, existing)
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
	// A document with a doc_id is named after it (INFO-QM-001.md).
	if name := mlcext.SuggestFileName(content, ".md", false); name != "" {
		defaultFilename = name
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

// ExportHTML saves the rendered markdown as a standalone HTML file. md (the
// source) only names the file: <doc_id>_v<version>.html when it has a
// doc_id.
func (a *App) ExportHTML(htmlContent string, cssContent string, md string) (string, error) {
	name := "exported.html"
	if s := mlcext.SuggestFileName(md, ".html", true); s != "" {
		name = s
	}
	path, err := application.Get().Dialog.SaveFile().
		SetMessage("Export to HTML").
		SetFilename(name).
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
