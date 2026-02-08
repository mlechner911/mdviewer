package main

import (
	"context"
	"fmt"

	"md-viewer/internal/filesystem"
	"md-viewer/internal/markdown"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct defines the main application state and dependencies.
// It acts as the bridge between the Svelte frontend and the Go backend.
type App struct {
	ctx      context.Context
	renderer *markdown.Renderer
}

// NewApp creates a new App application struct and initializes the markdown renderer.
func NewApp() *App {
	return &App{
		renderer: markdown.NewRenderer(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods (like logging and dialogs).
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name (Legacy/Example function).
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// RenderMarkdown converts markdown string to sanitized HTML using the internal renderer.
// It also takes a chromaStyle parameter to apply specific syntax highlighting colors.
func (a *App) RenderMarkdown(input string, theme string) string {
	runtime.LogDebugf(a.ctx, "Request: RenderMarkdown (Theme: %s)", theme)
	html, err := a.renderer.Render(input, theme)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to render markdown: %v", err)
		return fmt.Sprintf("<p>Error rendering markdown: %v</p>", err)
	}
	// Log a snippet of the HTML to verify attributes/classes
	if len(html) > 200 {
		runtime.LogDebugf(a.ctx, "HTML Snippet: %s...", html[:200])
	} else {
		runtime.LogDebugf(a.ctx, "HTML: %s", html)
	}
	return html
}

// GetStyleCSS returns the raw CSS for a specific syntax highlighting style from Chroma.
// This allows the frontend to inject highlighting styles dynamically.
func (a *App) GetStyleCSS(style string) string {
	css, err := a.renderer.GetStyleCSS(style)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to get CSS for style %s: %v", style, err)
		return ""
	}
	return css
}

// OpenFile opens a native file dialog, reads the selected file, and returns its content.
func (a *App) OpenFile() (string, error) {
	runtime.LogInfo(a.ctx, "Request: OpenFile")
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
		runtime.LogInfo(a.ctx, "OpenFile: User cancelled dialog")
		return "", nil
	}

	runtime.LogInfof(a.ctx, "OpenFile: Reading file %s", path)
	return filesystem.ReadFile(path)
}

// SaveFile opens a native save dialog and saves the provided content to the selected path.
func (a *App) SaveFile(content string) (string, error) {
	runtime.LogInfo(a.ctx, "Request: SaveFile")
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
		runtime.LogInfo(a.ctx, "SaveFile: User cancelled dialog")
		return "", nil
	}

	runtime.LogInfof(a.ctx, "SaveFile: Writing to file %s", path)
	err = filesystem.WriteFile(path, content)
	if err != nil {
		return "", err
	}

	return path, nil
}
