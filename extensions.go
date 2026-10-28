package main

import "github.com/wailsapp/wails/v3/pkg/application"

// extension is an optional part of MarkSafe, built in with a build tag
// (sensormonitor_on.go). Without the tag the list stays empty and none of
// its code is compiled.
type extension interface {
	// Service is the extension's Wails service (its bindings).
	Service() application.Service
	// FileMenu adds the extension's entries to the File menu; label
	// translates a menu key (fallback when the frontend sent none).
	FileMenu(menu *application.Menu, label func(key, fallback string) string)
	// HasSettings: the extension has a section in File › Settings….
	HasSettings() bool
}

// extensionFactories register themselves in init (see sensormonitor_on.go);
// main creates them with the version and registers each as a Wails service.
var extensionFactories []func(version string) extension

// hasSettings: some extension shows a section in File › Settings…; without
// one the menu entry is left out.
func (a *App) hasSettings() bool {
	for _, e := range a.extensions {
		if e.HasSettings() {
			return true
		}
	}
	return false
}
