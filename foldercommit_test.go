package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The test binary doubles as the check program: with MARKSAFE_FAKE_CHECK
// set it prints that JSON and exits 0, or 1 when it holds findings.
func TestMain(m *testing.M) {
	if out := os.Getenv("MARKSAFE_FAKE_CHECK"); out != "" {
		fmt.Println(out)
		if strings.Contains(out, `"severity"`) {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// commitApp: an allowed folder that is a git repo with a trusted check.
func commitApp(t *testing.T) (*App, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	a, dir := folderCheckApp(t)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "none"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	gitIn(t, dir, "init", "-q")
	gitIn(t, dir, "config", "user.name", "Test")
	gitIn(t, dir, "config", "user.email", "test@example.invalid")
	gitIn(t, dir, "config", "commit.gpgsign", "false")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".marksafe.yaml"), fmt.Sprintf("version: 1\ncheck:\n  command: [%q]\ncommit: true\n", exe))
	doc := filepath.Join(dir, "sop.md")
	writeFile(t, doc, "---\ndoc_id: SOP-QM-001\nversion: 3\n---\n# SOP\n")
	info, err := a.GetFolderCheck(doc)
	if err != nil || info.Spec == nil {
		t.Fatalf("info = %+v, err = %v", info, err)
	}
	if err := a.TrustFolderCheck(doc, info.Spec.Hash); err != nil {
		t.Fatal(err)
	}
	return a, doc
}

func TestCommitNeedsCleanCheckOfThisContent(t *testing.T) {
	a, doc := commitApp(t)

	if err := a.CommitFolderFile(doc, "x"); !errors.Is(err, errCommitNotChecked) {
		t.Fatalf("commit before any check: %v", err)
	}

	t.Setenv("MARKSAFE_FAKE_CHECK", `{"ok":false,"findings":[{"file":"sop.md","line":2,"severity":"error","message":"bad"}]}`)
	if _, err := a.RunFolderCheck(doc); err != nil {
		t.Fatal(err)
	}
	if err := a.CommitFolderFile(doc, "x"); !errors.Is(err, errCommitNotChecked) {
		t.Fatalf("commit after error findings: %v", err)
	}

	t.Setenv("MARKSAFE_FAKE_CHECK", `{"ok":true,"findings":[]}`)
	if _, err := a.RunFolderCheck(doc); err != nil {
		t.Fatal(err)
	}
	writeFile(t, doc, "---\ndoc_id: SOP-QM-001\nversion: 3\n---\n# SOP changed\n")
	if err := a.CommitFolderFile(doc, "x"); !errors.Is(err, errCommitNotChecked) {
		t.Fatalf("commit of content changed after the check: %v", err)
	}

	if _, err := a.RunFolderCheck(doc); err != nil {
		t.Fatal(err)
	}
	if err := a.CommitFolderFile(doc, "SOP-QM-001 v3: neu"); err != nil {
		t.Fatalf("clean check, same content: %v", err)
	}
}

func TestCommitOnlyWhenEnabled(t *testing.T) {
	a, doc := commitApp(t)
	dir := filepath.Dir(doc)
	exe, _ := os.Executable()
	writeFile(t, filepath.Join(dir, ".marksafe.yaml"), fmt.Sprintf("version: 1\ncheck:\n  command: [%q]\n", exe))
	if err := a.CommitFolderFile(doc, "x"); !errors.Is(err, errCommitNotEnabled) {
		t.Fatalf("err = %v", err)
	}
}

func TestCommitSuggestion(t *testing.T) {
	dir := t.TempDir()
	tests := []struct{ content, want string }{
		{"---\ndoc_id: SOP-QM-001\nversion: 3\n---\n# x\n", "SOP-QM-001 v3: "},
		{"---\ndoc_id: SOP-QM-001\n---\n# x\n", "SOP-QM-001: "},
		{"# nur Text\n", "doc.md: "},
	}
	for _, tt := range tests {
		p := filepath.Join(dir, "doc.md")
		writeFile(t, p, tt.content)
		if got := (&App{}).CommitSuggestion(p); got != tt.want {
			t.Errorf("CommitSuggestion = %q, want %q", got, tt.want)
		}
	}
}
