package main

// Folder check (plan P-MARKS-20261112-01): a folder with a .marksafe.yaml
// names a command MarkSafe runs after saving, and may ask for a commit.
// Nothing runs before the user allowed that exact command for the folder.

import (
	"context"
	"errors"
	"path/filepath"

	"marksafe/internal/config"
	"marksafe/internal/foldercheck"
)

// FolderCheckInfo describes the .marksafe.yaml that applies to a file.
type FolderCheckInfo struct {
	// Spec is nil when no .marksafe.yaml applies.
	Spec *foldercheck.Spec `json:"spec"`
	// Trusted: the user allowed this exact check for the folder.
	Trusted bool `json:"trusted"`
}

// errFolderCheckNotTrusted guards RunFolderCheck against a frontend that
// skipped the trust dialog.
var errFolderCheckNotTrusted = errors.New("folder check not allowed by the user")

// GetFolderCheck returns the .marksafe.yaml that applies to path, if any.
// A broken file is reported as an error so the user sees why nothing runs.
func (a *App) GetFolderCheck(path string) (*FolderCheckInfo, error) {
	spec, err := a.findFolderCheck(path)
	if err != nil || spec == nil {
		return &FolderCheckInfo{}, err
	}
	return &FolderCheckInfo{Spec: spec, Trusted: a.config.IsFolderCheckTrusted(spec.Dir, spec.Hash)}, nil
}

// TrustFolderCheck stores the user's permission for the check with this
// hash in dir. The hash must still match the file on disk, so a file
// changed while the dialog was open is not trusted by accident.
func (a *App) TrustFolderCheck(path, hash string) error {
	spec, err := a.findFolderCheck(path)
	if err != nil {
		return err
	}
	if spec == nil || spec.Hash != hash {
		return errors.New(".marksafe.yaml changed, please confirm again")
	}
	return a.config.TrustFolderCheck(spec.Dir, spec.Hash)
}

// RunFolderCheck runs the check that applies to path. It refuses to run a
// check the user has not allowed.
func (a *App) RunFolderCheck(path string) (*foldercheck.Result, error) {
	spec, err := a.findFolderCheck(path)
	if err != nil {
		return nil, err
	}
	if spec == nil || spec.Check == nil {
		return nil, errors.New("no check configured for this folder")
	}
	if !a.config.IsFolderCheckTrusted(spec.Dir, spec.Hash) {
		return nil, errFolderCheckNotTrusted
	}
	res := foldercheck.Run(context.Background(), spec.Check)
	res.Resolve(spec.Dir)
	return res, nil
}

// findFolderCheck looks for .marksafe.yaml above path, bounded by the
// document folder or, outside it, by the whitelisted folder holding path.
func (a *App) findFolderCheck(path string) (*foldercheck.Spec, error) {
	if path == "" || !a.config.IsPathAllowed(path) {
		return nil, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return foldercheck.Find(abs, folderCheckBoundary(a.config, a.documentRoot(), abs))
}

func folderCheckBoundary(cfg *config.ConfigManager, docRoot, abs string) string {
	if docRoot != "" {
		if rel, err := filepath.Rel(docRoot, abs); err == nil && filepath.IsLocal(rel) {
			return docRoot
		}
	}
	return cfg.WhitelistedRootOf(abs)
}
