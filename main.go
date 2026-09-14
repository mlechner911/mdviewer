package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Check for command line arguments (first argument after program name)
	if len(os.Args) > 1 {
		app.SetInitialFile(os.Args[1])
	}

	// Create application with options
	wailsApp := application.New(application.Options{
		Name:        "MarkSafe",
		Description: "Markdown viewer and editor",
		Services: []application.Service{
			application.NewService(app),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		FileAssociations: []string{".md", ".markdown", ".mdown"},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// macOS hands "Open with" over as an application event instead of argv.
	wailsApp.Event.OnApplicationEvent(events.Common.ApplicationOpenedWithFile, func(e *application.ApplicationEvent) {
		if app.initialFile == "" {
			app.SetInitialFile(e.Context().Filename())
		}
	})

	app.window = wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "MarkSafe",
		Width:            1200,
		Height:           800,
		BackgroundColour: application.NewRGB(27, 38, 54),
		// Enable Drag and Drop support; only elements with data-file-drop-target accept drops
		EnableFileDrop: true,
		Windows: application.WindowsWindow{
			// Dark title bar on Windows
			Theme: application.Dark,
		},
		URL: "/",
	})

	// Forward dropped files to the frontend (see frontend/src/lib/wails.ts).
	app.window.OnWindowEvent(events.Common.WindowFilesDropped, func(e *application.WindowEvent) {
		if paths := e.Context().DroppedFiles(); len(paths) > 0 {
			wailsApp.Event.Emit("files-dropped", paths)
		}
	})

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}
