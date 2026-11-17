//go:build !uitest

package main

// Every normal build: no debugging port, no mark (uitest_on.go is the
// test build).

const uiTestMark = ""

func uiTestBrowserArgs() []string { return nil }
