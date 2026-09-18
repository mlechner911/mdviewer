package browse

import (
	"fmt"
	"os"
	"path/filepath"
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
