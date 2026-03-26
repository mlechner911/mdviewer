package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// AppConfig holds the application settings, primarily the security whitelists.
type AppConfig struct {
	WhitelistedPaths []string `json:"whitelisted_paths"`
	WhitelistedURLs  []string `json:"whitelisted_urls"`
	RecentFiles      []string `json:"recent_files"`
}

// ConfigManager handles concurrent-safe access to the AppConfig.
type ConfigManager struct {
	config AppConfig
	path   string
	mu     sync.RWMutex
}

// NewConfigManager initializes a manager, ensuring the config file exists in the user's app data directory.
func NewConfigManager() (*ConfigManager, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(userConfigDir, "marksafe")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(appDir, "config.json")
	m := &ConfigManager{
		path: configPath,
	}

	if err := m.Load(); err != nil {
		// If file doesn't exist, we start with an empty config
		m.config = AppConfig{
			WhitelistedPaths: []string{},
			WhitelistedURLs:  []string{},
			RecentFiles:      []string{},
		}
		return m, m.Save()
	}

	return m, nil
}

// Load reads the config from disk.
func (m *ConfigManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &m.config)
	if err != nil {
		return err
	}

	// Ensure slices are initialized
	if m.config.WhitelistedPaths == nil {
		m.config.WhitelistedPaths = []string{}
	}
	if m.config.WhitelistedURLs == nil {
		m.config.WhitelistedURLs = []string{}
	}
	if m.config.RecentFiles == nil {
		m.config.RecentFiles = []string{}
	}

	return nil
}

// Save writes the config to disk.
func (m *ConfigManager) Save() error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.config, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}

// GetRecentFiles returns the list of recently opened files.
func (m *ConfigManager) GetRecentFiles() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.RecentFiles
}

// AddRecentFile adds a file path to the list of recent files, keeping only the most recent 10.
func (m *ConfigManager) AddRecentFile(path string) error {
	m.mu.Lock()
	
	// Remove if already exists to move it to the top
	newList := []string{path}
	for _, f := range m.config.RecentFiles {
		if f != path {
			newList = append(newList, f)
		}
	}

	// Limit to 10 items
	if len(newList) > 10 {
		newList = newList[:10]
	}
	m.config.RecentFiles = newList
	m.mu.Unlock()

	return m.Save()
}

// IsPathAllowed checks if a local file path is within a whitelisted directory.
func (m *ConfigManager) IsPathAllowed(targetPath string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	for _, p := range m.config.WhitelistedPaths {
		absWhitelist, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		// Check if the target is within the whitelisted path
		if strings.HasPrefix(absTarget, absWhitelist) {
			return true
		}
	}
	return false
}

// IsURLAllowed checks if a URL's host is whitelisted.
func (m *ConfigManager) IsURLAllowed(targetURL string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.config.WhitelistedURLs {
		if strings.Contains(targetURL, u) {
			return true
		}
	}
	return false
}

// AddPath adds a directory to the whitelist.
func (m *ConfigManager) AddPath(path string) error {
	m.mu.Lock()
	// Check if already exists
	exists := false
	for _, p := range m.config.WhitelistedPaths {
		if p == path {
			exists = true
			break
		}
	}
	if !exists {
		m.config.WhitelistedPaths = append(m.config.WhitelistedPaths, path)
	}
	m.mu.Unlock()
	return m.Save()
}

// AddURL adds a domain/URL pattern to the whitelist.
func (m *ConfigManager) AddURL(url string) error {
	m.mu.Lock()
	exists := false
	for _, u := range m.config.WhitelistedURLs {
		if u == url {
			exists = true
			break
		}
	}
	if !exists {
		m.config.WhitelistedURLs = append(m.config.WhitelistedURLs, url)
	}
	m.mu.Unlock()
	return m.Save()
}
