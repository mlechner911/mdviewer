package main

// Saving = commit (plan P-MARKS-20261112-01 Phase 3). The backend checks the
// rules itself, fail closed: a commit needs a .marksafe.yaml with commit:
// true that the user allowed, and a clean check of exactly this content.

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mlechner911/mdviewer/mlcext"

	"marksafe/internal/foldercheck"
)

// cleanRun remembers the last check per folder: when it started, the file
// whose save started it and that file's content, and whether it was clean.
// The check covers the whole folder, so any file unchanged since the start
// may be committed after a clean run.
type cleanRun struct {
	hash    string
	file    string
	sum     [32]byte
	started time.Time
	clean   bool
}

// unchangedSince reports whether abs is as the run checked it: the same
// content for the file that started the run, else not modified after the
// run started.
func (r cleanRun) unchangedSince(abs string) bool {
	if abs == r.file {
		sum, err := fileSum(abs)
		return err == nil && sum == r.sum
	}
	st, err := os.Stat(abs)
	return err == nil && !st.ModTime().After(r.started)
}

type folderRunLog struct {
	mu   sync.Mutex
	runs map[string]cleanRun
}

func (l *folderRunLog) set(dir string, r cleanRun) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.runs == nil {
		l.runs = map[string]cleanRun{}
	}
	l.runs[dir] = r
}

func (l *folderRunLog) get(dir string) (cleanRun, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.runs[dir]
	return r, ok
}

func fileSum(path string) ([32]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(data), nil
}

// Errors of CommitFolderFile the UI explains.
var (
	errCommitNotEnabled = errors.New("this folder does not commit on save")
	errCommitNotChecked = errors.New("the file changed since the last check, or the check found errors")
)

// CommitSuggestion is the prefilled commit message: "<doc_id> v<version>: "
// from the front matter, else "<file name>: ".
func (a *App) CommitSuggestion(path string) string {
	data, err := os.ReadFile(path)
	if err == nil {
		if m := mlcext.DocumentMeta(string(data)); m.DocID != "" {
			if m.Version != "" {
				return m.DocID + " v" + m.Version + ": "
			}
			return m.DocID + ": "
		}
	}
	return filepath.Base(path) + ": "
}

// CommitFolderFile commits only path with message, after a clean check of
// its current content. It never pushes.
func (a *App) CommitFolderFile(path, message string) error {
	spec, err := a.findFolderCheck(path)
	if err != nil {
		return err
	}
	if spec == nil || !spec.Commit {
		return errCommitNotEnabled
	}
	if !a.config.IsFolderCheckTrusted(spec.Dir, spec.Hash) {
		return errFolderCheckNotTrusted
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	run, ok := a.folderRuns.get(spec.Dir)
	if !ok || !run.clean || run.hash != spec.Hash || !run.unchangedSince(abs) {
		return errCommitNotChecked
	}
	return foldercheck.CommitFile(context.Background(), abs, message)
}
