//go:build sensormonitor

package main

// Built in with the tag "sensormonitor" (task … SENSOR=1): the connection
// to the MLC Sensor Monitor, the private submodule frontend/src/sensormonitor.

import sensormonitor "marksafe/frontend/src/sensormonitor/backend"

func init() {
	extensionFactories = append(extensionFactories, func(version string) extension {
		return sensormonitor.NewExtension(version)
	})
}
