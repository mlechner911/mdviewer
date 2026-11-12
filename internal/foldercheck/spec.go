// Package foldercheck reads a folder's .marksafe.yaml and runs the check
// command it names. It knows nothing about any particular project: a folder
// opts in by carrying the file (plan P-MARKS-20261112-01).
package foldercheck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// FileName is the name of the per-folder settings file.
const FileName = ".marksafe.yaml"

// DefaultTimeout applies when check.timeout is not set.
const DefaultTimeout = 60 * time.Second

// Spec is a parsed .marksafe.yaml.
type Spec struct {
	// Dir is the folder holding the file; finding paths are relative to it.
	Dir string `json:"dir"`
	// Version of the file format; only 1 is known.
	Version int `json:"version"`
	// Check is nil when the file names no check command.
	Check *Check `json:"check,omitempty"`
	// Commit: saving a file also commits it (only without error findings).
	Commit bool `json:"commit"`
	// Hash identifies check and commit; trust is granted for this value.
	Hash string `json:"hash"`
}

// Check is the command run after saving.
type Check struct {
	// Command is the program and its arguments, run without a shell.
	Command []string `json:"command"`
	// Cwd is the absolute working directory.
	Cwd string `json:"cwd"`
	// Timeout ends the run; the result then counts as failed.
	Timeout time.Duration `json:"timeout"`
}

type rawSpec struct {
	Version int `yaml:"version"`
	Check   *struct {
		Command []string `yaml:"command"`
		Cwd     string   `yaml:"cwd"`
		Timeout string   `yaml:"timeout"`
	} `yaml:"check"`
	Commit bool `yaml:"commit"`
}

// Find looks for .marksafe.yaml from the folder of file upwards, never above
// boundary. It returns nil without error when there is none. A file outside
// boundary is never looked up.
func Find(file, boundary string) (*Spec, error) {
	dir, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		return nil, fmt.Errorf("foldercheck: %w", err)
	}
	boundary, err = filepath.Abs(boundary)
	if err != nil || boundary == "" {
		return nil, nil
	}
	if !within(dir, boundary) {
		return nil, nil
	}
	for {
		path := filepath.Join(dir, FileName)
		if _, err := os.Stat(path); err == nil {
			return Load(path)
		}
		if samePath(dir, boundary) {
			return nil, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, nil
		}
		dir = parent
	}
}

// Load reads and validates one .marksafe.yaml.
func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("foldercheck: %w", err)
	}
	var raw rawSpec
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	dir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("foldercheck: %w", err)
	}
	spec, err := build(raw, dir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return spec, nil
}

func build(raw rawSpec, dir string) (*Spec, error) {
	if raw.Version != 1 {
		return nil, fmt.Errorf("version %d is not supported (this MarkSafe knows version 1)", raw.Version)
	}
	spec := &Spec{Dir: dir, Version: raw.Version, Commit: raw.Commit}
	if raw.Check != nil {
		check, err := buildCheck(raw, dir)
		if err != nil {
			return nil, err
		}
		spec.Check = check
	}
	if spec.Commit && spec.Check == nil {
		return nil, errors.New("commit needs a check command")
	}
	spec.Hash = spec.hash()
	return spec, nil
}

func buildCheck(raw rawSpec, dir string) (*Check, error) {
	c := raw.Check
	if len(c.Command) == 0 || c.Command[0] == "" {
		return nil, errors.New("check.command is empty")
	}
	check := &Check{Command: c.Command, Cwd: dir, Timeout: DefaultTimeout}
	if c.Cwd != "" {
		cwd := c.Cwd
		if !filepath.IsAbs(cwd) {
			cwd = filepath.Join(dir, cwd)
		}
		check.Cwd = filepath.Clean(cwd)
	}
	if c.Timeout != "" {
		d, err := time.ParseDuration(c.Timeout)
		if err != nil || d <= 0 {
			return nil, fmt.Errorf("check.timeout %q is not a duration like 120s", c.Timeout)
		}
		check.Timeout = d
	}
	return check, nil
}

// hash covers everything that decides what runs and what gets committed.
func (s *Spec) hash() string {
	data, _ := json.Marshal(struct {
		Dir    string
		Check  *Check
		Commit bool
	}{s.Dir, s.Check, s.Commit})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
