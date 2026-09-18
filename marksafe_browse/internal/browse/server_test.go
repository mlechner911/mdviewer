package browse

import (
	"strings"
	"testing"
)

// Regression test: discoverAssets must pick the app bundle, never
// katex.min.css (same dir!). A wrong pick silently ships an unstyled
// page — exactly what happened before this test existed.
func TestDiscoverAssetsFindsAppBundle(t *testing.T) {
	js, css := discoverAssets()

	if !(strings.HasPrefix(js, "index-") && strings.HasSuffix(js, ".js")) {
		t.Errorf("js asset = %q, want index-*.js", js)
	}
	if !(strings.HasPrefix(css, "index-") && strings.HasSuffix(css, ".css")) {
		t.Errorf("css asset = %q, want index-*.css", css)
	}
	if css == "katex.min.css" {
		t.Errorf("css asset picked katex.min.css instead of the app bundle")
	}
	if _, err := assetsFS.ReadFile("assets/" + js); err != nil {
		t.Errorf("js asset not embedded: %v", err)
	}
	if _, err := assetsFS.ReadFile("assets/" + css); err != nil {
		t.Errorf("css asset not embedded: %v", err)
	}
}

// The shell must reference the app bundle exactly once (plus katex;
// duplicated/missing links ship broken styling with zero errors).
func TestPageShellLinksAppBundle(t *testing.T) {
	js, css := discoverAssets()
	s := &Server{jsAsset: js, cssAsset: css}
	html := s.pageShell("T", "")
	if got := strings.Count(html, "/assets/"+css); got != 1 {
		t.Errorf("app css referenced %d times, want 1", got)
	}
	if got := strings.Count(html, "/assets/"+js); got != 1 {
		t.Errorf("app js referenced %d times, want 1", got)
	}
	if !strings.Contains(html, `type="module"`) {
		t.Error("bundle must load as ES module (code-split chunks use import/export)")
	}
}
