package main

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"marksafe/internal/config"
	"marksafe/internal/filesystem"
	"marksafe/internal/markdown"

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
		a.ShowAbout(t["aboutTitle"], t["aboutBody"])
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

	// Submenu: Language
	langMenu := viewMenu.AddSubmenu(t["menuLanguage"])
	langMenu.Add("English").OnClick(func(*application.Context) { emit("set-locale", "en") })
	langMenu.Add("Deutsch").OnClick(func(*application.Context) { emit("set-locale", "de") })
	langMenu.Add("Español").OnClick(func(*application.Context) { emit("set-locale", "es") })
	langMenu.Add("Français").OnClick(func(*application.Context) { emit("set-locale", "fr") })

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

// ShowAbout displays a native message box with product information.
func (a *App) ShowAbout(title, message string) {
	if title == "" {
		title = "Über MarkSafe"
	}
	if message == "" {
		message = "MarkSafe v" + a.GetVersion() + "\n\nMarkdown-Betrachter und -Editor\n\nCopyright (c) 2026 Michael Lechner\nLizenziert unter MIT."
	}
	application.Get().Dialog.Info().
		SetTitle(title).
		SetMessage(message).
		Show()
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
	if defaultFilename == "" {
		defaultFilename = "document.md"
	} else {
		defaultFilename = filepath.Base(defaultFilename)
	}
	selectedPath, err := application.Get().Dialog.SaveFile().
		SetMessage("Save Markdown File").
		SetFilename(defaultFilename).
		AddFilter("Markdown Files (*.md)", "*.md").
		AddFilter("All Files (*.*)", "*.*").
		PromptForSingleSelection()
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

	katexCSS := a.renderer.GetKatexCSS()

	// Create a standalone HTML document
	fullHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>MarkSafe Export</title>
    <!-- Optional: Uncomment the following line to use KaTeX fonts from CDN if you have internet access -->
    <!-- <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.16.21/dist/katex.min.css"> -->
    <style>
        :root {
            --bg-color: #ffffff;
            --text-color: #24292e;
            --link-color: #0969da;
            --border-color: #dfe2e5;
            --code-bg: #f6f8fa;
            --alert-bg: rgba(0, 0, 0, 0.03);
        }

        /* KaTeX Embedded */
        %s
        
        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif; 
            padding: 2rem; 
            max-width: 900px; 
            margin: 0 auto; 
            line-height: 1.6; 
            background-color: var(--bg-color);
            color: var(--text-color); 
        }
        
        .markdown-body { box-sizing: border-box; min-width: 200px; max-width: 980px; margin: 0 auto; }
        
        /* Syntax Highlighting and Theme Overrides */
        %s
        
        a { color: var(--link-color); text-decoration: none; }
        a:hover { text-decoration: underline; }

        .markdown-alert { 
            padding: 0.75rem 1rem; 
            margin-bottom: 1rem; 
            border-left: 0.25rem solid; 
            border-radius: 0 0.375rem 0.375rem 0; 
            background: var(--alert-bg); 
        }
        .markdown-alert-note { border-color: #0969da; }
        .markdown-alert-tip { border-color: #1a7f37; }
        .markdown-alert-important { border-color: #8250df; }
        .markdown-alert-warning { border-color: #9a6700; }
        .markdown-alert-caution { border-color: #cf222e; }
        
        pre { background: var(--code-bg); padding: 1rem; border-radius: 6px; overflow: auto; }
        code { font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace; font-size: 85%%; }
        
        table { border-collapse: collapse; width: 100%%; margin: 1rem 0; display: block; overflow: auto; }
        th { font-weight: 600; background-color: var(--code-bg); }
        th, td { border: 1px solid var(--border-color); padding: 6px 13px; }
        tr:nth-child(2n) { background-color: var(--code-bg); }
        
        img { max-width: 100%%; box-sizing: content-box; background-color: #fff; }
        blockquote { padding: 0 1em; color: #6a737d; border-left: 0.25em solid var(--border-color); margin: 0 0 1rem 0; }

        /* Mermaid Placeholder Styling */
        .mermaid {
            background: var(--code-bg);
            padding: 1rem;
            border-radius: 0.5rem;
            margin: 1.5rem 0;
            text-align: center;
            font-family: sans-serif;
            border: 1px dashed var(--border-color);
        }
    </style>
</head>
<body>
    <article class="markdown-body">
        %s
    </article>
</body>
</html>`, katexCSS, cssContent, htmlContent)

	err = filesystem.WriteFile(path, fullHTML)
	if err != nil {
		return "", err
	}
	return path, nil
}
