//go:build sensormonitor

package main

// Built in with the tag "sensormonitor" (task … SENSOR=1): the connection
// to the MLC Sensor Monitor (package sensormonitor).

import "marksafe/sensormonitor"

func init() {
	extensionFactories = append(extensionFactories, func(version string) extension {
		return sensormonitor.NewExtension(version)
	})
}
