//go:build sensormonitor

package main

// Built in with the tag "sensormonitor" (task … SENSOR=1): the connection
// to the MLC Sensor Monitor, the private submodule frontend/src/sensormonitor.

import (
	"os"
	"path/filepath"

	sensormonitor "marksafe/frontend/src/sensormonitor/backend"
)

func init() {
	productName = "MarkSafePlus"
	webviewDataPath = func() string {
		dir, err := os.UserConfigDir()
		if err != nil {
			return ""
		}
		return filepath.Join(dir, productName, "webview")
	}
	extensionFactories = append(extensionFactories, func(version string) extension {
		return sensormonitor.NewExtension(version)
	})
}
