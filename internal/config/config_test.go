package config

import (
	"path/filepath"
	"testing"
)

func TestIsWithin(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "home", "user", "docs")

	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{"same directory", root, true},
		{"file inside", filepath.Join(root, "readme.md"), true},
		{"nested file", filepath.Join(root, "screenshots", "shot.png"), true},
		{"sibling with shared prefix", root + "-private", false},
		{"sibling file with shared prefix", filepath.Join(root+"-private", "secret.png"), false},
		{"parent directory", filepath.Dir(root), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWithin(tt.target, root); got != tt.want {
				t.Errorf("isWithin(%q, %q) = %v, want %v", tt.target, root, got, tt.want)
			}
		})
	}
}

func TestIsWithinRootWithTrailingSeparator(t *testing.T) {
	root := string(filepath.Separator)
	if !isWithin(filepath.Join(root, "any", "file.png"), root) {
		t.Errorf("expected every absolute path to be within the filesystem root")
	}
}

func TestIsURLAllowed(t *testing.T) {
	m := &ConfigManager{config: AppConfig{WhitelistedURLs: []string{"example.com", "https://Docs.MLCgo.eu/path"}}}

	tests := []struct {
		target string
		want   bool
	}{
		{"example.com", true},
		{"sub.example.com", true},
		{"https://example.com/img.png", true},
		{"https://example.com:8443/x", true},
		{"evil-example.com", false},
		{"example.com.evil.net", false},
		{"docs.mlcgo.eu", true},
		{"mlcgo.eu", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := m.IsURLAllowed(tt.target); got != tt.want {
			t.Errorf("IsURLAllowed(%q) = %v, want %v", tt.target, got, tt.want)
		}
	}
}
