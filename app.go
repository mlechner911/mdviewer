package main

import (
	"context"
	"fmt"

	"md-viewer/internal/filesystem"
	"md-viewer/internal/markdown"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct defines the main application state and dependencies.
type App struct {
	ctx         context.Context
	renderer    *markdown.Renderer
	initialFile string
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{
		renderer: markdown.NewRenderer(),
	}
}

// SetInitialFile stores the file path provided via CLI.
func (a *App) SetInitialFile(path string) {
	a.initialFile = path
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ReadFile reads the content of a file given its path.
// This is used for Drag and Drop files received by the frontend.
func (a *App) ReadFile(path string) (string, error) {
	runtime.LogInfof(a.ctx, "ReadFile: Reading file %s", path)
	return filesystem.ReadFile(path)
}

// GetInitialContent is called by the frontend on mount to check for CLI files.
func (a *App) GetInitialContent() string {
	if a.initialFile != "" {
		content, err := filesystem.ReadFile(a.initialFile)
		if err == nil {
			return content
		}
	}
	return ""
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

// OpenFile opens a native file dialog.
func (a *App) OpenFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Markdown File",
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown Files (*.md)", Pattern: "*.md"},
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	return filesystem.ReadFile(path)
}

// SaveFile opens a native save dialog and saves the provided content to the selected path.
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

	err = filesystem.WriteFile(path, content)
	if err != nil {
		return "", err
	}

	return path, nil
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

	// Create a standalone HTML document
	fullHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>MD Viewer Export</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.16.0/dist/katex.min.css">
    <style>
        body { font-family: sans-serif; padding: 2rem; max-width: 900px; margin: 0 auto; line-height: 1.6; }
        %s
        .markdown-alert { padding: 0.75rem 1rem; margin-bottom: 1rem; border-left: 0.25rem solid; background: #f8f9fa; }
        .markdown-alert-note { border-color: #0969da; }
        .markdown-alert-tip { border-color: #1a7f37; }
        .markdown-alert-important { border-color: #8250df; }
        .markdown-alert-warning { border-color: #9a6700; }
        .markdown-alert-caution { border-color: #cf222e; }
        pre { background: #f6f8fa; padding: 1rem; border-radius: 6px; overflow: auto; }
        table { border-collapse: collapse; width: 100%%; margin: 1rem 0; }
        th, td { border: 1px solid #dfe2e5; padding: 6px 13px; }
        tr:nth-child(2n) { background-color: #f6f8fa; }
    </style>
</head>
<body>
    <article class="markdown-body">
        %s
    </article>
</body>
</html>`, cssContent, htmlContent)

	err = filesystem.WriteFile(path, fullHTML)
	if err != nil {
		return "", err
	}
	return path, nil
}
