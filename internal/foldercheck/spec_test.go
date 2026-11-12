package foldercheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const sample = `version: 1
check:
  command: ["go", "run", "./cmd/x", "-json"]
  cwd: ../..
  timeout: 120s
commit: true
`

func TestFind(t *testing.T) {
	root := t.TempDir()
	seeds := filepath.Join(root, "repo", "internal", "seeds")
	write(t, filepath.Join(seeds, FileName), sample)
	file := filepath.Join(seeds, "common", "sop.md")
	write(t, file, "# x")

	tests := []struct {
		name     string
		file     string
		boundary string
		wantDir  string
	}{
		{"found upwards", file, root, seeds},
		{"boundary is the folder itself", file, seeds, seeds},
		{"boundary below the file stops search", file, filepath.Join(seeds, "common"), ""},
		{"file outside boundary", file, filepath.Join(root, "other"), ""},
		{"no file anywhere", filepath.Join(root, "repo", "a.md"), root, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := Find(tt.file, tt.boundary)
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantDir == "" {
				if spec != nil {
					t.Fatalf("want none, got %s", spec.Dir)
				}
				return
			}
			if spec == nil || spec.Dir != tt.wantDir {
				t.Fatalf("want %s, got %+v", tt.wantDir, spec)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	write(t, path, sample)
	spec, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !spec.Commit || spec.Check == nil {
		t.Fatalf("spec = %+v", spec)
	}
	if want := filepath.Clean(filepath.Join(dir, "../..")); spec.Check.Cwd != want {
		t.Errorf("cwd = %s, want %s", spec.Check.Cwd, want)
	}
	if spec.Check.Timeout != 120*time.Second {
		t.Errorf("timeout = %s", spec.Check.Timeout)
	}
	if len(spec.Hash) != 64 {
		t.Errorf("hash = %q", spec.Hash)
	}
}

func TestLoadDefaultsAndErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{"defaults", "version: 1\ncheck:\n  command: [\"x\"]\n", ""},
		{"no check, no commit", "version: 1\n", ""},
		{"unknown version", "version: 2\n", "version 2"},
		{"missing version", "check:\n  command: [\"x\"]\n", "version 0"},
		{"empty command", "version: 1\ncheck:\n  command: []\n", "check.command"},
		{"bad timeout", "version: 1\ncheck:\n  command: [\"x\"]\n  timeout: soon\n", "check.timeout"},
		{"commit without check", "version: 1\ncommit: true\n", "commit needs"},
		{"not yaml", "version: [", "yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), FileName)
			write(t, path, tt.content)
			spec, err := Load(path)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				if spec.Check != nil && (spec.Check.Timeout != DefaultTimeout || spec.Check.Cwd != filepath.Dir(path)) {
					t.Errorf("defaults not applied: %+v", spec.Check)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestHashChangesWithCommand(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	write(t, path, sample)
	a, _ := Load(path)
	write(t, path, strings.Replace(sample, "-json", "-JSON", 1))
	b, _ := Load(path)
	write(t, path, strings.Replace(sample, "commit: true", "commit: false", 1))
	c, _ := Load(path)
	if a.Hash == b.Hash || a.Hash == c.Hash {
		t.Error("hash must change when command or commit changes")
	}
	write(t, path, "# comment\n"+sample)
	d, _ := Load(path)
	if a.Hash != d.Hash {
		t.Error("a comment must not change the hash")
	}
}
