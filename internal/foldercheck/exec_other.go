//go:build !windows

package foldercheck

import "os/exec"

func hideConsole(*exec.Cmd) {}
