//go:build !windows

package mlcdoc

import "os/exec"

func hideConsole(*exec.Cmd) {}

func defaultLocations() []string {
	return []string{"/usr/local/bin/mlcdoc", "/opt/homebrew/bin/mlcdoc"}
}
