// Package filesystem provides internal utilities for basic file I/O operations.
// These are decoupled from the UI-specific runtime calls used in the app layer.
package filesystem

import (
	"os"
)

// ReadFile reads the entire content of a file at the given path and returns it as a string.
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile writes the provided string content to a file at the given path.
// It uses 0644 permissions (standard readable/writable for owner, readable for others).
func WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}