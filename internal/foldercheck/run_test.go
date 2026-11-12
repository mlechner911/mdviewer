package foldercheck

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The test binary doubles as the fake check program: run with
// FOLDERCHECK_FAKE set, TestMain prints the canned output and exits.
func TestMain(m *testing.M) {
	if mode := os.Getenv("FOLDERCHECK_FAKE"); mode != "" {
		os.Exit(fake(mode))
	}
	os.Exit(m.Run())
}

func fake(mode string) int {
	switch mode {
	case "clean":
		fmt.Println(`{"ok":true,"findings":[]}`)
		return 0
	case "findings":
		fmt.Println("building…")
		fmt.Println(`{"ok":false,"findings":[{"file":"common/a.md","line":3,"severity":"error","code":"x","message":"bad"},{"file":"b.md","severity":"warning","message":"meh"}]}`)
		return 1
	case "warnings":
		fmt.Println(`{"ok":true,"findings":[{"file":"b.md","severity":"warning","message":"meh"}]}`)
		return 0
	case "notfolder":
		fmt.Println(`{"ok":false,"error":"catalog.yaml missing"}`)
		return 2
	case "crash":
		fmt.Fprintln(os.Stderr, "panic: boom")
		return 3
	case "garbage":
		fmt.Println("not json")
		return 0
	case "slow":
		time.Sleep(10 * time.Second)
	}
	return 0
}

func runFake(t *testing.T, mode string, timeout time.Duration) *Result {
	t.Helper()
	t.Setenv("FOLDERCHECK_FAKE", mode)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return Run(context.Background(), &Check{Command: []string{exe}, Cwd: t.TempDir(), Timeout: timeout})
}

func TestRun(t *testing.T) {
	tests := []struct {
		mode       string
		wantFailed bool
		wantError  bool
		findings   int
		blocks     bool
	}{
		{"clean", false, false, 0, false},
		{"findings", false, false, 2, true},
		{"warnings", false, false, 1, false},
		{"notfolder", false, true, 0, true},
		{"crash", true, false, 0, true},
		{"garbage", true, false, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			r := runFake(t, tt.mode, 10*time.Second)
			if r.Failed != tt.wantFailed || (r.Error != "") != tt.wantError || len(r.Findings) != tt.findings {
				t.Fatalf("result = %+v", r)
			}
			if r.HasErrors() != tt.blocks {
				t.Errorf("HasErrors = %v, want %v", r.HasErrors(), tt.blocks)
			}
		})
	}
}

func TestRunCrashKeepsStderr(t *testing.T) {
	r := runFake(t, "crash", 10*time.Second)
	if !strings.Contains(r.Stderr, "panic: boom") || !strings.Contains(r.Stderr, "exit code 3") {
		t.Errorf("stderr = %q", r.Stderr)
	}
}

func TestRunTimeout(t *testing.T) {
	start := time.Now()
	r := runFake(t, "slow", 300*time.Millisecond)
	if !r.Failed || !strings.Contains(r.Stderr, "timeout") {
		t.Fatalf("result = %+v", r)
	}
	if time.Since(start) > 5*time.Second {
		t.Error("timeout did not stop the command")
	}
}

func TestRunProgramNotFound(t *testing.T) {
	r := Run(context.Background(), &Check{Command: []string{"bin/missing"}, Cwd: t.TempDir(), Timeout: time.Second})
	if !r.Failed || !strings.Contains(r.Stderr, "not found") {
		t.Fatalf("result = %+v", r)
	}
}

func TestInterpretExit1WithoutFindings(t *testing.T) {
	if r := interpret(1, `{"ok":false,"findings":[]}`, ""); !r.Failed {
		t.Errorf("exit 1 without findings must fail closed: %+v", r)
	}
}

func TestResolve(t *testing.T) {
	r := &Result{Findings: []Finding{{File: "common/a.md"}, {File: ""}}}
	r.Resolve(filepath.Join("base", "seeds"))
	if want := filepath.Join("base", "seeds", "common", "a.md"); r.Findings[0].Path != want {
		t.Errorf("path = %q, want %q", r.Findings[0].Path, want)
	}
	if r.Findings[1].Path != "" {
		t.Errorf("a finding without file has no path: %q", r.Findings[1].Path)
	}
}
