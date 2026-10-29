package main

// The native application menu: built from the frontend's translations
// (UpdateMenu) and the entries that hand actions to the frontend.

import (
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// label returns the translation for key, or fallback if it is missing.
func label(t map[string]string, key, fallback string) string {
	if s := t[key]; s != "" {
		return s
	}
	return fallback
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
	// Entries of built-in extensions (extensions.go), e.g. "Sensor Monitor".
	for _, e := range a.extensions {
		e.FileMenu(fileMenu, func(key, fallback string) string { return label(t, key, fallback) })
	}
	fileMenu.AddSeparator()
	// Settings only exist for extensions so far: without one, no entry.
	if a.hasSettings() {
		fileMenu.Add(label(t, "menuSettings", "Settings…")).SetAccelerator("CmdOrCtrl+,").OnClick(func(*application.Context) {
			emit("menu-settings")
		})
	}
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
