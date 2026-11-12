package foldercheck

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// newRepo makes a repository with one committed file and an identity.
func newRepo(t *testing.T, identity bool) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	// Keep the user's global config (identity, hooks, signing) out.
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "none"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	dir := t.TempDir()
	mustGit(t, dir, "init", "-q")
	mustGit(t, dir, "config", "commit.gpgsign", "false")
	if identity {
		mustGit(t, dir, "config", "user.name", "Test")
		mustGit(t, dir, "config", "user.email", "test@example.invalid")
	}
	write(t, filepath.Join(dir, "a.md"), "a\n")
	write(t, filepath.Join(dir, "b.md"), "b\n")
	if identity {
		mustGit(t, dir, "add", ".")
		mustGit(t, dir, "commit", "-q", "-m", "init")
	}
	return dir
}

func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := git(context.Background(), dir, args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestCommitFileOnlyThatFile(t *testing.T) {
	dir := newRepo(t, true)
	write(t, filepath.Join(dir, "a.md"), "a changed\n")
	write(t, filepath.Join(dir, "b.md"), "b changed\n")
	mustGit(t, dir, "add", "b.md") // staged by someone else: must stay staged

	if err := CommitFile(context.Background(), filepath.Join(dir, "a.md"), "  SOP-1 v2: Text  "); err != nil {
		t.Fatal(err)
	}
	if got := mustGit(t, dir, "log", "-1", "--format=%s"); got != "SOP-1 v2: Text" {
		t.Errorf("message = %q", got)
	}
	if got := mustGit(t, dir, "show", "--name-only", "--format=", "HEAD"); got != "a.md" {
		t.Errorf("committed files = %q, want only a.md", got)
	}
	if got := mustGit(t, dir, "diff", "--cached", "--name-only"); got != "b.md" {
		t.Errorf("b.md must stay staged, got %q", got)
	}
}

func TestCommitFileNewFile(t *testing.T) {
	dir := newRepo(t, true)
	write(t, filepath.Join(dir, "sub", "new.md"), "new\n")
	if err := CommitFile(context.Background(), filepath.Join(dir, "sub", "new.md"), "add"); err != nil {
		t.Fatal(err)
	}
	if got := mustGit(t, dir, "show", "--name-only", "--format=", "HEAD"); got != "sub/new.md" {
		t.Errorf("committed files = %q", got)
	}
}

func TestCommitFileErrors(t *testing.T) {
	t.Run("nothing to commit", func(t *testing.T) {
		dir := newRepo(t, true)
		err := CommitFile(context.Background(), filepath.Join(dir, "a.md"), "x")
		if !errors.Is(err, ErrNothingToCommit) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("no identity", func(t *testing.T) {
		dir := newRepo(t, false)
		err := CommitFile(context.Background(), filepath.Join(dir, "a.md"), "x")
		if !errors.Is(err, ErrNoIdentity) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("no repo", func(t *testing.T) {
		t.Setenv("GIT_CEILING_DIRECTORIES", os.TempDir())
		dir := t.TempDir()
		write(t, filepath.Join(dir, "a.md"), "a")
		err := CommitFile(context.Background(), filepath.Join(dir, "a.md"), "x")
		if !errors.Is(err, ErrNoRepo) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("empty message", func(t *testing.T) {
		err := CommitFile(context.Background(), "/x/a.md", "  ")
		if !errors.Is(err, ErrEmptyMessage) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("failing hook is reported", func(t *testing.T) {
		dir := newRepo(t, true)
		hook := filepath.Join(dir, ".git", "hooks", "pre-commit")
		write(t, hook, "#!/bin/sh\necho 'hook says no' >&2\nexit 1\n")
		if err := os.Chmod(hook, 0o755); err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(dir, "a.md"), "changed\n")
		err := CommitFile(context.Background(), filepath.Join(dir, "a.md"), "x")
		if err == nil || !strings.Contains(err.Error(), "hook says no") {
			t.Fatalf("err = %v", err)
		}
	})
}
