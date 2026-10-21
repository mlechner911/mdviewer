package main

// Closing with unsaved changes: the window's X and quitting (Cmd+Q) are
// held back while the frontend reports unsaved tabs; it asks the user
// (save / don't save / cancel) and then calls CloseApp.

import (
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type closeState struct {
	unsaved atomic.Bool
	allowed atomic.Bool // CloseApp was called: close without asking
}

// SetUnsavedChanges is reported by the frontend whenever a tab becomes
// dirty or is saved.
func (a *App) SetUnsavedChanges(unsaved bool) {
	a.closing.unsaved.Store(unsaved)
}

// shouldClose: close now, or hold back and let the frontend ask.
func (a *App) shouldClose() bool {
	if a.closing.allowed.Load() || !a.closing.unsaved.Load() {
		return true
	}
	emit("app-close-requested")
	return false
}

// CloseApp quits after the user answered (changes saved or discarded).
func (a *App) CloseApp() {
	a.closing.allowed.Store(true)
	if app := application.Get(); app != nil {
		app.Quit()
	}
}
