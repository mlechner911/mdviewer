package main

import (
	"context"
	"fmt"

	"md-viewer/internal/filesystem"
	"md-viewer/internal/markdown"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	renderer *markdown.Renderer
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		renderer: markdown.NewRenderer(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// RenderMarkdown converts markdown to HTML using the internal renderer
func (a *App) RenderMarkdown(input string) string {
	html, err := a.renderer.Render(input)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to render markdown: %v", err)
		return fmt.Sprintf("<p>Error rendering markdown: %v</p>", err)
	}
	return html
}

// OpenFile opens a file dialog and returns the content using internal filesystem package
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

// SaveFile opens a save dialog and saves the content using internal filesystem package
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