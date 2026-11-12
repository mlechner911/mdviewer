package foldercheck

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Commit errors the UI explains in its own words.
var (
	ErrNoGit           = errors.New("git is not installed")
	ErrNoRepo          = errors.New("the file is not in a git repository")
	ErrNoIdentity      = errors.New("git user.name or user.email is not set on this computer")
	ErrNothingToCommit = errors.New("the file has no changes to commit")
	ErrEmptyMessage    = errors.New("the commit message is empty")
)

// gitTimeout bounds every git call; commit hooks of the repository run too.
const gitTimeout = 60 * time.Second

// CommitFile commits only file, with message, as the git identity of the
// computer. Other staged changes stay as they are. It never pushes.
func CommitFile(ctx context.Context, file, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return ErrEmptyMessage
	}
	if _, err := exec.LookPath("git"); err != nil {
		return ErrNoGit
	}
	dir := filepath.Dir(file)
	if _, err := git(ctx, dir, "rev-parse", "--show-toplevel"); err != nil {
		return ErrNoRepo
	}
	for _, key := range []string{"user.name", "user.email"} {
		if v, _ := git(ctx, dir, "config", "--get", key); v == "" {
			return ErrNoIdentity
		}
	}
	name := filepath.Base(file)
	status, err := git(ctx, dir, "status", "--porcelain", "--", name)
	if err != nil {
		return err
	}
	if status == "" {
		return ErrNothingToCommit
	}
	if _, err := git(ctx, dir, "add", "--", name); err != nil {
		return err
	}
	_, err = git(ctx, dir, "commit", "--only", "--no-edit", "-m", message, "--", name)
	return err
}

// git runs one git command in dir and returns its trimmed stdout. Git never
// prompts: there is no terminal behind MarkSafe.
func git(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_EDITOR=true")
	cmd.WaitDelay = waitDelay
	hideConsole(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("git %s: %s: %w", args[0], msg, err)
	}
	return strings.TrimSpace(stdout.String()), nil
}
