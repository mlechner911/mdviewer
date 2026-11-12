//go:build windows

package foldercheck

import (
	"os/exec"
	"syscall"
)

// hideConsole keeps the console window of the check program and of git from
// flashing up when MarkSafe (a GUI program) runs them.
func hideConsole(cmd *exec.Cmd) {
	const createNoWindow = 0x08000000
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
