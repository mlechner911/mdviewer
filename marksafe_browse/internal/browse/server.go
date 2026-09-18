package browse

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"marksafe/internal/markdown"
	"marksafe/internal/mdbrowse"
)

//go:embed assets/*
var assetsFS embed.FS

// Server holds the HTTP server configuration and state.
type Server struct {
	cfg      *Config
	root     *mdbrowse.Root
	jsAsset  string // discovered hashed bundle, e.g. "index-a1b2c3.js"
	cssAsset string // discovered hashed stylesheet, e.g. "index-d4e5f6.css"
	// Chroma code colors per UI theme, generated once at startup from the
	// shared renderer (same implementation as the Wails preview).
	chromaDark  string
	chromaLight string
}

// discoverAssets finds the current hashed bundle names in the embedded
// assets dir (vite emits content hashes so browsers never cache stale
// builds). Falls back to legacy fixed names if discovery fails.
func discoverAssets() (js, css string) {
	js, css = "index.js", "style.css"
	entries, err := fs.ReadDir(assetsFS, "assets")
	if err != nil {
		return js, css
	}
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "index-") && strings.HasSuffix(n, ".js") {
			js = n
		}
		if strings.HasSuffix(n, ".css") {
			css = n
		}
	}
	return js, css
}

// NewServer creates a new Server instance.
func NewServer(cfg *Config) (*Server, error) {
	root, err := mdbrowse.NewRoot(cfg.Root)
	if err != nil {
		return nil, err
	}
	js, css := discoverAssets()
	chroma := markdown.NewRenderer()
	darkCSS, _ := chroma.GetStyleCSS("github-dark")
	lightCSS, _ := chroma.GetStyleCSS("github")
	return &Server{cfg: cfg, root: root, jsAsset: js, cssAsset: css, chromaDark: darkCSS, chromaLight: lightCSS}, nil
}

// Start begins the HTTP server.
// Start begins the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Bind, s.cfg.Port)
	mux := http.NewServeMux()
	secureMux := s.securityMiddleware(mux)

	mux.HandleFunc("/", s.handlePage)
	mux.HandleFunc("/tree", s.handleTree)
	mux.HandleFunc("/render", s.handleRender)
	mux.HandleFunc("/md/", s.handleMarkdown)
	mux.HandleFunc("/raw", s.handleRaw)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/assets/chroma-dark.css", s.serveText("text/css; charset=utf-8", s.chromaDark))
	mux.HandleFunc("/assets/chroma-light.css", s.serveText("text/css; charset=utf-8", s.chromaLight))

	// Serve static assets from embedded directory
	assetsSub, _ := fs.Sub(assetsFS, "assets")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsSub))))
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := assetsFS.ReadFile("assets/favicon.ico")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/x-icon")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	fmt.Printf("  📡 MarkSafe Browse serving %s on %s:%d\n", s.cfg.Root, s.cfg.Bind, s.cfg.Port)
	return http.ListenAndServe(addr, secureMux)
}

// serveText serves a fixed in-memory text asset.
func (s *Server) serveText(contentType, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}
}

// chromaStyle maps the UI theme to code colors (mirrors the Wails preview).
func chromaStyle(theme string) string {
	if theme == "light" {
		return "github"
	}
	return "github-dark"
}

// securityMiddleware validates that every request stays within root.
func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			http.Error(w, "Forbidden: path escape detected", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handlePage serves the main TOC page or index.md content.
func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Always serve the empty app shell — the Svelte client fetches
	// index.md itself via /render. Pre-filling #app with goldmark HTML
	// would mismatch the client render and break mount/hydrate.
	s.renderTOCPage(w)
}

// handleTree returns the directory structure as JSON.
func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	entries, err := s.root.Scan()
	if err != nil {
		http.Error(w, "Error scanning directory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// handleRender converts markdown to HTML via the shared mdbrowse pipeline
// (frontmatter stripped, pasted-markdown repaired, links validated).
func (s *Server) handleRender(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}
	if _, err := s.root.Resolve(path); err != nil {
		http.Error(w, fmt.Sprintf("Security violation: %v", err), http.StatusForbidden)
		return
	}
	if !strings.HasSuffix(path, ".md") {
		http.Error(w, "Not a markdown file", http.StatusBadRequest)
		return
	}
	title, result, err := s.root.RenderDoc(path, chromaStyle(r.URL.Query().Get("theme")))
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"title": title, "html": result, "path": path,
	})
}

// handleMarkdown serves a raw markdown file.
func (s *Server) handleMarkdown(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/md/")
	absPath, err := s.root.Resolve(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Security violation: %v", err), http.StatusForbidden)
		return
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	title, _ := mdbrowse.ExtractTitle(absPath)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "# %s\n\n%s", title, content)
}

// handleRaw serves any file within the root with its proper content type
// (images, PDFs, …). Missing files → 404, escapes → 403 via Resolve.
func (s *Server) handleRaw(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}
	absPath, err := s.root.Resolve(relPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Security violation: %v", err), http.StatusForbidden)
		return
	}
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	f, err := os.Open(absPath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	ctype := mime.TypeByExtension(strings.ToLower(filepath.Ext(absPath)))
	if ctype == "" {
		ctype = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ctype)
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

// handleSearch answers GET /api/search?q=... with ranked JSON hits.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	results, err := s.root.Search(q)
	if err != nil {
		http.Error(w, "Error scanning directory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// pageShell wraps the Svelte app shell around a title and #app content.
// Asset filenames carry content hashes (see discoverAssets); the shell
// always references the current build, so browsers cache safely.
func (s *Server) pageShell(title, appHTML string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<link rel="stylesheet" href="/assets/%s">
<link rel="stylesheet" href="/assets/katex.min.css">
<link id="chroma-css" rel="stylesheet" href="/assets/chroma-dark.css">
</head>
<body class="dark">
<div id="app">%s</div>
<script type="module" src="/assets/%s"></script>
</body>
</html>`, title, s.cssAsset, appHTML, s.jsAsset)
}

// renderTOCPage serves the empty app shell. The Svelte client fetches
// index.md itself via /render; pre-filling #app with server HTML would
// mismatch the client render and break mount/hydrate.
func (s *Server) renderTOCPage(w http.ResponseWriter) {
	page := s.pageShell("MarkSafe Browse", "")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(page))
}
