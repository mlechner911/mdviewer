package browse

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config holds the configuration for the browse server.
type Config struct {
	Root     string // absolute path to the root directory
	Port     int    // server port
	Bind     string // bind address
	Theme    string // dark, light, auto
	ShowTree bool   // show directory tree
}

// ValidateConfig validates the configuration and resolves paths.
func ValidateConfig(root string, port int, bind string, theme string, showTree bool) (*Config, error) {
	// Resolve root to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path: %w", err)
	}

	// Verify root exists
	if _, err := os.Stat(absRoot); err != nil {
		return nil, fmt.Errorf("root directory does not exist: %s", absRoot)
	}

	// Validate theme
	theme = strings.ToLower(theme)
	if theme != "dark" && theme != "light" && theme != "auto" {
		return nil, fmt.Errorf("invalid theme %q, must be 'dark', 'light', or 'auto'", theme)
	}

	// Default bind
	if bind == "" {
		bind = "127.0.0.1"
	}
	if port <= 0 || port > 65535 {
		port = 8080
	}

	return &Config{
		Root:     absRoot,
		Port:     port,
		Bind:     bind,
		Theme:    theme,
		ShowTree: showTree,
	}, nil
}

// ResolvePath validates that the requested path is within the root directory.
// This is the "harte Grenze" (hard boundary) security check.
func (c *Config) ResolvePath(requestedPath string) (string, error) {
	absRequested, err := filepath.Abs(filepath.Join(c.Root, requestedPath))
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Ensure the resolved path is within root
	rel, err := filepath.Rel(c.Root, absRequested)
	if err != nil {
		return "", fmt.Errorf("path escape detected: %s is outside %s", requestedPath, c.Root)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("security violation: %s escapes root %s", requestedPath, c.Root)
	}

	// On Windows, do case-insensitive comparison
	if runtime.GOOS == "windows" {
		absRequested = strings.ToLower(absRequested)
		c.Root = strings.ToLower(c.Root)
	}

	if !strings.HasPrefix(absRequested, c.Root) {
		return "", fmt.Errorf("security violation: %s is outside root %s", requestedPath, c.Root)
	}

	return absRequested, nil
}

// IsWithinRoot checks if the given absolute path is within the root directory.
func (c *Config) IsWithinRoot(absPath string) bool {
	rel, err := filepath.Rel(c.Root, absPath)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		rel = strings.ToLower(rel)
		c.Root = strings.ToLower(c.Root)
	}
	return !strings.HasPrefix(rel, "..")
}

// ConfigString returns a human-readable config summary.
func (c *Config) ConfigString() string {
	return fmt.Sprintf(
		"MarkSafe Browse\n"+
			"  Root:  %s\n"+
			"  Serve: %s:%d\n"+
			"  Theme: %s\n"+
			"  Tree:  %v\n",
		c.Root, c.Bind, c.Port, c.Theme, c.ShowTree,
	)
}
