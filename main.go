// Package main is the entry point for the MD Viewer application.
// It initializes the Wails application and sets up the asset server and bindings.
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// main is the application's entry point.
func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	// The configuration here handles window size, asset serving, and Go-to-JS bindings.
	err := wails.Run(&options.App{
		Title:  "md-viewer",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
