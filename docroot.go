package main

// The document folder: the base folder of the document collection
// (templates, RefNo proposals, Save As start folder, browse window).

import (
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

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
