package main

import (
	"embed"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
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

	// Create application menu
	appMenu := menu.NewMenu()
	
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("New Tab", keys.CmdOrCtrl("t"), func(_ *menu.CallbackData) {
		app.MenuNewTab()
	})
	fileMenu.AddText("Open...", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		app.MenuOpenFile()
	})
	fileMenu.AddText("Save", keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) {
		app.MenuSaveFile()
	})
	fileMenu.AddSeparator()
	
	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.EditMenu())
	} else {
		editMenu := appMenu.AddSubmenu("Edit")
		editMenu.AddText("Undo", keys.CmdOrCtrl("z"), func(_ *menu.CallbackData) {})
		editMenu.AddText("Redo", keys.CmdOrCtrl("y"), func(_ *menu.CallbackData) {})
		editMenu.AddSeparator()
		editMenu.AddText("Cut", keys.CmdOrCtrl("x"), func(_ *menu.CallbackData) {})
		editMenu.AddText("Copy", keys.CmdOrCtrl("c"), func(_ *menu.CallbackData) {})
		editMenu.AddText("Paste", keys.CmdOrCtrl("v"), func(_ *menu.CallbackData) {})
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "MD Viewer",
		Width:  1200,
		Height: 800,
		Menu:   appMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		// Enable Drag and Drop support
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
