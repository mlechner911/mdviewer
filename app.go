package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"md-viewer/internal/config"
	"md-viewer/internal/filesystem"
	"md-viewer/internal/markdown"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// FileResult holds the outcome of a file open operation.
type FileResult struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// App struct defines the main application state and dependencies.
type App struct {
	ctx         context.Context
	renderer    *markdown.Renderer
	config      *config.ConfigManager
	initialFile string
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

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// UpdateMenu dynamically updates the application menu with translated strings.
func (a *App) UpdateMenu(t map[string]string) {
	appMenu := menu.NewMenu()

	// File Menu
	fileMenu := appMenu.AddSubmenu(t["menuFile"])
	fileMenu.AddText(t["menuNewTab"], keys.CmdOrCtrl("t"), func(_ *menu.CallbackData) {
		a.MenuNewTab()
	})
	fileMenu.AddText(t["menuOpen"], keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		a.MenuOpenFile()
	})
	fileMenu.AddText(t["menuSave"], keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) {
		a.MenuSaveFile()
	})
	
	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.AppMenu())
		appMenu.Append(menu.EditMenu())
	} else {
		editMenu := appMenu.AddSubmenu(t["menuEdit"])
		editMenu.AddText(t["menuUndo"], keys.CmdOrCtrl("z"), func(_ *menu.CallbackData) {})
		editMenu.AddText(t["menuRedo"], keys.CmdOrCtrl("y"), func(_ *menu.CallbackData) {})
		editMenu.AddSeparator()
		editMenu.AddText(t["menuCut"], keys.CmdOrCtrl("x"), func(_ *menu.CallbackData) {})
		editMenu.AddText(t["menuCopy"], keys.CmdOrCtrl("c"), func(_ *menu.CallbackData) {})
		editMenu.AddText(t["menuPaste"], keys.CmdOrCtrl("v"), func(_ *menu.CallbackData) {})
	}

	runtime.MenuSetApplicationMenu(a.ctx, appMenu)
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
	runtime.LogInfof(a.ctx, "ReadFile: Reading file %s", path)
	return filesystem.ReadFile(path)
}

// GetInitialContent is called by the frontend on mount to check for CLI files.
func (a *App) GetInitialContent() *FileResult {
	if a.initialFile != "" {
		content, err := filesystem.ReadFile(a.initialFile)
		if err == nil {
			return &FileResult{
				Path:    a.initialFile,
				Content: content,
			}
		}
	}
	return nil
}

// RenderMarkdown converts markdown string to sanitized HTML.
func (a *App) RenderMarkdown(input string, theme string) string {
	runtime.LogDebugf(a.ctx, "Request: RenderMarkdown (Theme: %s)", theme)
	html, err := a.renderer.Render(input, theme)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to render markdown: %v", err)
		return fmt.Sprintf("<p>Error rendering markdown: %v</p>", err)
	}
	return html
}

// GetStyleCSS returns the raw CSS for a specific syntax highlighting style.
func (a *App) GetStyleCSS(style string) string {
	css, err := a.renderer.GetStyleCSS(style)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to get CSS for style %s: %v", style, err)
		return ""
	}
	return css
}

// MenuOpenFile is called from the native application menu.
func (a *App) MenuOpenFile() {
	runtime.EventsEmit(a.ctx, "menu-open-file")
}

// MenuSaveFile is called from the native application menu.
func (a *App) MenuSaveFile() {
	runtime.EventsEmit(a.ctx, "menu-save-file")
}

// MenuNewTab is called from the native application menu.
func (a *App) MenuNewTab() {
	runtime.EventsEmit(a.ctx, "menu-new-tab")
}

// OpenFile opens a native file dialog and returns the path and content.
func (a *App) OpenFile() (*FileResult, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Markdown File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown Files (*.md)", Pattern: "*.md"},
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
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
	return &FileResult{
		Path:    absPath,
		Content: content,
	}, nil
}

// GetFileTitle extracts the base name from a file path.
func (a *App) GetFileTitle(path string) string {
	return filepath.Base(path)
}

// SaveFile opens a native save dialog and saves content.
func (a *App) SaveFile(content string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title: "Save Markdown File",
		DefaultFilename: "document.md",
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown Files (*.md)", Pattern: "*.md"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	absPath, _ := filepath.Abs(path)
	err = filesystem.WriteFile(absPath, content)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

// ExportHTML saves the rendered markdown as a standalone HTML file.
func (a *App) ExportHTML(htmlContent string, cssContent string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export to HTML",
		DefaultFilename: "exported.html",
		Filters: []runtime.FileFilter{
			{DisplayName: "HTML Files (*.html)", Pattern: "*.html"},
		},
	})
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
    <title>MD Viewer Export</title>
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
        %%s
        
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
        %%s
        
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
        code { font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace; font-size: 85%%%%; }
        
        table { border-collapse: collapse; width: 100%%%%; margin: 1rem 0; display: block; overflow: auto; }
        th { font-weight: 600; background-color: var(--code-bg); }
        th, td { border: 1px solid var(--border-color); padding: 6px 13px; }
        tr:nth-child(2n) { background-color: var(--code-bg); }
        
        img { max-width: 100%%%%; box-sizing: content-box; background-color: #fff; }
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
        %%s
    </article>
</body>
</html>`, katexCSS, cssContent, htmlContent)

	err = filesystem.WriteFile(path, fullHTML)
	if err != nil {
		return "", err
	}
	return path, nil
}
