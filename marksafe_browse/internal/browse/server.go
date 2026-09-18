package browse

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

//go:embed assets/*
var assetsFS embed.FS

// Server holds the HTTP server configuration and state.
type Server struct {
	cfg      *Config
	scanner  *Scanner
	jsAsset  string // discovered hashed bundle, e.g. "index-a1b2c3.js"
	cssAsset string // discovered hashed stylesheet, e.g. "style-d4e5f6.css"
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
	scanner, err := NewScanner(cfg.Root)
	if err != nil {
		return nil, err
	}
	js, css := discoverAssets()
	return &Server{cfg: cfg, scanner: scanner, jsAsset: js, cssAsset: css}, nil
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
	s.renderTOCPage(w, nil)
}

// handleTree returns the directory structure as JSON.
func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	entries, err := s.scanner.Scan()
	if err != nil {
		http.Error(w, "Error scanning directory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// handleRender converts markdown to HTML via goldmark.
func (s *Server) handleRender(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}
	absPath, err := s.cfg.ResolvePath(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Security violation: %v", err), http.StatusForbidden)
		return
	}
	if !strings.HasSuffix(absPath, ".md") {
		http.Error(w, "Not a markdown file", http.StatusBadRequest)
		return
	}
	content, err := readFile(absPath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	title, _ := ExtractTitle(absPath)
	result := renderMarkdownToHTML(content)
	// Validate internal links/images against the root so broken targets
	// never reach the user as a clickable 404.
	result = s.rewriteMarkdownLinks(result, path)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"title": title, "html": result, "path": path,
	})
}

// handleMarkdown serves a raw markdown file.
func (s *Server) handleMarkdown(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/md/")
	absPath, err := s.cfg.ResolvePath(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Security violation: %v", err), http.StatusForbidden)
		return
	}
	content, err := readFile(absPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	title, _ := ExtractTitle(absPath)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "# %s\n\n%s", title, content)
}

// renderTOCPage generates the HTML table of contents.
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
</head>
<body class="dark">
<div id="app">%s</div>
<script defer src="/assets/%s"></script>
</body>
</html>`, title, s.cssAsset, appHTML, s.jsAsset)
}

// renderTOCPage generates the HTML table of contents.
func (s *Server) renderTOCPage(w http.ResponseWriter, entries []TOCEntry) {
	page := s.pageShell("MarkSafe Browse", "")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(page))
}

// renderMarkdownPage renders a markdown file as a full HTML page.
func (s *Server) renderMarkdownPage(w http.ResponseWriter, title, markdownContent string) {
	result := renderMarkdownToHTML(markdownContent)
	page := s.pageShell(escapeHTML(title)+" — MarkSafe Browse", result)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(page))
}

// readFile reads a file's content.
func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

// stripFrontmatter removes YAML frontmatter delimiters from markdown content.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	endIdx := strings.Index(content[3:], "---")
	if endIdx == -1 {
		return content
	}
	rest := content[3+endIdx+3:]
	return strings.TrimLeft(rest, "\n")
}

// renderMarkdownToHTML converts markdown to HTML using goldmark.
func renderMarkdownToHTML(content string) string {
	content = stripFrontmatter(content)
	// Repair glued block markers from pasted/AI-generated sources
	// (normalize.go); well-formed documents pass through unchanged.
	content = normalizePastedMarkdown(content)
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Table, extension.Strikethrough, extension.TaskList),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return "<div class=\"markdown\"><p>Error: " + escapeHTML(err.Error()) + "</p></div>"
	}
	return "<div class=\"markdown\">" + buf.String() + "</div>"
}

// escapeHTML provides basic HTML escaping.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
