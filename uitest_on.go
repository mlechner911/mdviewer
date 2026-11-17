//go:build uitest

package main

import (
	"os"
	"strconv"
)

// Test build only (tag uitest, task handtest:prepare): WebView2 opens its
// DevTools protocol on 127.0.0.1, so Playwright can drive the interface
// (connectOverCDP). Whoever reaches the port controls the app completely,
// the folder check's permission and command included, so a release never
// has this file: task release refuses UITEST=1 and checks the binaries.
// Wails empties WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS; only the app can
// open the port.

// uiTestMark goes before every window title, so nobody works with it.
const uiTestMark = "[UI-TEST] "

// uiTestBrowserArgs: the port from MARKSAFE_CDP_PORT, default 9333.
func uiTestBrowserArgs() []string {
	port := 9333
	if p, err := strconv.Atoi(os.Getenv("MARKSAFE_CDP_PORT")); err == nil && p > 0 && p < 65536 {
		port = p
	}
	return []string{"--remote-debugging-port=" + strconv.Itoa(port)}
}
