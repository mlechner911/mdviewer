//go:build windows

package mlcdoc

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// hideConsole keeps mlcdoc's console window from flashing up when a GUI
// program (MarkSafe) runs it.
func hideConsole(cmd *exec.Cmd) {
	const createNoWindow = 0x08000000
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}

// defaultLocations: where the mlcdoc installer puts it. A program started
// before the installation does not see the PATH the installer extended.
func defaultLocations() []string {
	var out []string
	for _, env := range []string{"ProgramFiles", "ProgramW6432"} {
		if dir := os.Getenv(env); dir != "" {
			out = append(out, filepath.Join(dir, "MLC", "mlcdoc", "mlcdoc.exe"))
		}
	}
	return out
}
