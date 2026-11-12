package foldercheck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Finding is one result of the check. File is relative to Spec.Dir; Line
// and Col are 0 when the rule knows no position.
type Finding struct {
	File     string `json:"file"`
	Line     int    `json:"line,omitempty"`
	Col      int    `json:"col,omitempty"`
	Severity string `json:"severity"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message"`
	// Path is the absolute path of File, filled in by Resolve.
	Path string `json:"path,omitempty"`
}

// Result of one check run.
type Result struct {
	Findings []Finding `json:"findings"`
	// Error is the message of exit code 2 (not a folder the check knows).
	Error string `json:"error,omitempty"`
	// Failed: the command did not run, crashed, timed out or printed no
	// result. Stderr holds the end of its error output.
	Failed bool   `json:"failed,omitempty"`
	Stderr string `json:"stderr,omitempty"`
}

// Resolve fills in the absolute Path of every finding, relative to dir
// (the folder of the .marksafe.yaml).
func (r *Result) Resolve(dir string) {
	for i := range r.Findings {
		if f := r.Findings[i].File; f != "" {
			r.Findings[i].Path = filepath.Join(dir, filepath.FromSlash(f))
		}
	}
}

// HasErrors reports whether the result blocks a commit (fail closed).
func (r *Result) HasErrors() bool {
	if r.Failed || r.Error != "" {
		return true
	}
	for _, f := range r.Findings {
		if f.Severity != "warning" {
			return true
		}
	}
	return false
}

// stderrTail is how much of the error output a failed run keeps.
const stderrTail = 2000

// waitDelay bounds the wait for children that keep the pipes open after
// the command itself was killed (go run starts the real binary as a child).
const waitDelay = 2 * time.Second

// Run executes the check command. It never uses a shell.
func Run(ctx context.Context, check *Check) *Result {
	ctx, cancel := context.WithTimeout(ctx, check.Timeout)
	defer cancel()

	prog, err := resolveProgram(check.Command[0], check.Cwd)
	if err != nil {
		return &Result{Failed: true, Stderr: err.Error()}
	}
	cmd := exec.CommandContext(ctx, prog, check.Command[1:]...)
	cmd.Dir = check.Cwd
	cmd.WaitDelay = waitDelay
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return failed(fmt.Sprintf("timeout after %s", check.Timeout), stderr.String())
	}
	exit := 0
	if runErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) {
			return failed(runErr.Error(), stderr.String())
		}
		exit = exitErr.ExitCode()
	}
	return interpret(exit, stdout.String(), stderr.String())
}

// interpret maps exit code and output to a result: 0 and 1 carry findings,
// 2 carries an error message, everything else failed.
func interpret(exit int, stdout, stderr string) *Result {
	if exit < 0 || exit > 2 {
		return failed(fmt.Sprintf("exit code %d", exit), stderr)
	}
	var out struct {
		OK       bool      `json:"ok"`
		Findings []Finding `json:"findings"`
		Error    string    `json:"error"`
	}
	if err := json.Unmarshal([]byte(lastLine(stdout)), &out); err != nil {
		return failed("no JSON result on stdout", stderr)
	}
	if exit == 2 {
		msg := out.Error
		if msg == "" {
			msg = "the check does not apply to this folder"
		}
		return &Result{Error: msg, Findings: []Finding{}}
	}
	if exit == 1 && len(out.Findings) == 0 {
		return failed("exit code 1 without findings", stderr)
	}
	if out.Findings == nil {
		out.Findings = []Finding{}
	}
	return &Result{Findings: out.Findings}
}

func failed(reason, stderr string) *Result {
	tail := strings.TrimSpace(stderr)
	if len(tail) > stderrTail {
		tail = "…" + tail[len(tail)-stderrTail:]
	}
	if tail == "" {
		tail = reason
	} else {
		tail = reason + "\n" + tail
	}
	return &Result{Failed: true, Stderr: tail, Findings: []Finding{}}
}

// lastLine is the last non-empty line; tools may log before the result.
func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// resolveProgram: a bare name ("go") is looked up in PATH; a path
// ("bin/check") is relative to cwd. On Windows ".exe" is tried as well.
func resolveProgram(name, cwd string) (string, error) {
	if !strings.ContainsAny(name, `/\`) {
		p, err := exec.LookPath(name)
		if err != nil {
			return "", fmt.Errorf("program %q not found: %w", name, err)
		}
		return p, nil
	}
	p := name
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}
	// On Windows "bin/check" means bin/check.exe first: a file without an
	// extension next to it (a macOS or Linux build) cannot run there.
	candidates := []string{p}
	if runtime.GOOS == "windows" && filepath.Ext(p) == "" {
		candidates = []string{p + ".exe", p}
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("program %q not found", p)
}
