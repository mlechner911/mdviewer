//go:build sensormonitor

// SPDX-License-Identifier: Elastic-2.0
// Copyright (c) 2026 Michael Lechner. Part of the optional Sensor Monitor
// module, licensed under the Elastic License 2.0: see sensormonitor/LICENSE.

package main

// Built in with the tag "sensormonitor" (task … SENSOR=1): the connection
// to the MLC Sensor Monitor (package sensormonitor).

import "marksafe/sensormonitor"

func init() {
	extensionFactories = append(extensionFactories, func(version string) extension {
		return sensormonitor.NewExtension(version)
	})
}
