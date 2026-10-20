package config

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// AppConfig holds the application settings, primarily the security whitelists.
type AppConfig struct {
	WhitelistedPaths []string `json:"whitelisted_paths"`
	WhitelistedURLs  []string `json:"whitelisted_urls"`
	RecentFiles      []string `json:"recent_files"`
	// DocumentRoot is the fixed base folder of the document collection
	// (like an Obsidian vault); empty until the user chooses one.
	DocumentRoot string `json:"document_root,omitempty"`
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

// GetDocumentRoot returns the base folder of the document collection.
// Dir is the folder of the configuration file (<user config>/marksafe).
func (m *ConfigManager) Dir() string {
	return filepath.Dir(m.path)
}

func (m *ConfigManager) GetDocumentRoot() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.DocumentRoot
}

// SetDocumentRoot stores the base folder of the document collection.
func (m *ConfigManager) SetDocumentRoot(dir string) error {
	m.mu.Lock()
	m.config.DocumentRoot = dir
	m.mu.Unlock()
	return m.Save()
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
		if isWithin(absTarget, absWhitelist) {
			return true
		}
	}
	return false
}

// isWithin reports whether target equals dir or lies below it. Comparing on a
// path-separator boundary keeps a whitelisted "/docs" from also allowing
// "/docs-private". Windows paths are compared case-insensitively.
func isWithin(target, dir string) bool {
	if runtime.GOOS == "windows" {
		target, dir = strings.ToLower(target), strings.ToLower(dir)
	}
	if target == dir {
		return true
	}
	if !strings.HasSuffix(dir, string(filepath.Separator)) {
		dir += string(filepath.Separator)
	}
	return strings.HasPrefix(target, dir)
}

// hostOf extracts the lower-cased hostname from a full URL
// ("https://Example.com:8080/x") or a bare host ("example.com") — the
// preview passes bare hostnames, older configs may hold full URLs.
func hostOf(s string) string {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "://") {
		s = "//" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

// IsURLAllowed checks if a URL's host is whitelisted. Matching is by
// domain, not substring: "example.com" allows "sub.example.com" but
// not "evil-example.com".
func (m *ConfigManager) IsURLAllowed(targetURL string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	target := hostOf(targetURL)
	if target == "" {
		return false
	}
	for _, u := range m.config.WhitelistedURLs {
		pattern := hostOf(u)
		if pattern == "" {
			continue
		}
		if target == pattern || strings.HasSuffix(target, "."+pattern) {
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
