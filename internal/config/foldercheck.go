package config

// IsFolderCheckTrusted reports whether the user allowed the check with this
// hash to run in dir (the folder of its .marksafe.yaml).
func (m *ConfigManager) IsFolderCheckTrusted(dir, hash string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return hash != "" && m.config.TrustedFolderChecks[dir] == hash
}

// TrustFolderCheck stores the user's permission to run the check with this
// hash in dir; it replaces an earlier permission for the folder.
func (m *ConfigManager) TrustFolderCheck(dir, hash string) error {
	m.mu.Lock()
	if m.config.TrustedFolderChecks == nil {
		m.config.TrustedFolderChecks = map[string]string{}
	}
	m.config.TrustedFolderChecks[dir] = hash
	m.mu.Unlock()
	return m.Save()
}

// WhitelistedRootOf returns the whitelisted folder that contains path (the
// deepest one when several do), "" when none does.
func (m *ConfigManager) WhitelistedRootOf(path string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	best := ""
	for _, p := range m.config.WhitelistedPaths {
		if isWithin(path, p) && len(p) > len(best) {
			best = p
		}
	}
	return best
}
