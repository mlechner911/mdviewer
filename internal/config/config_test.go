package config

import (
	"os"
	"path/filepath"
	"strings"
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

// The Sensor Monitor address of MarkSafe 1.8 survives a save of the
// standard build, so the optional module can still take it over.
func TestSaveKeepsLegacySensorMonitorURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"sensor_monitor_url":"https://sm.example"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	m := &ConfigManager{path: path}
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), `"sensor_monitor_url": "https://sm.example"`) {
		t.Fatalf("saved config: %s", b)
	}
}
