package browse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// Server holds the HTTP server configuration and state.
type Server struct {
	cfg     *Config
	scanner *Scanner
}

// NewServer creates a new Server instance.
func NewServer(cfg *Config) (*Server, error) {
	scanner, err := NewScanner(cfg.Root)
	if err != nil {
		return nil, err
	}
	return &Server{
		cfg:     cfg,
		scanner: scanner,
	}, nil
}

// Start begins the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Bind, s.cfg.Port)
	mux := http.NewServeMux()
	secureMux := s.securityMiddleware(mux)

	mux.HandleFunc("/", s.handlePage)
	mux.HandleFunc("/tree", s.handleTree)
	mux.HandleFunc("/render", s.handleRender)
	mux.HandleFunc("/md/", s.handleMarkdown)
	mux.HandleFunc("/assets/", s.handleAssets)

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

	if indexPath, found := s.scanner.FindIndex(); found {
		content, err := readFile(indexPath)
		if err != nil {
			http.Error(w, "Error reading index.md", http.StatusInternalServerError)
			return
		}
		title, _ := ExtractTitle(indexPath)
		s.renderMarkdownPage(w, title, content)
		return
	}

	entries, err := s.scanner.Scan()
	if err != nil {
		http.Error(w, "Error scanning directory", http.StatusInternalServerError)
		return
	}

	s.renderTOCPage(w, entries)
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
	html := renderMarkdownToHTML(content)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"title": title,
		"html":  html,
		"path":  path,
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

// handleAssets serves static files.
func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/assets/")
	if strings.Contains(path, "..") || strings.HasSuffix(path, "/") {
		http.NotFound(w, r)
		return
	}
	absPath, err := s.cfg.ResolvePath("frontend/" + path)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, absPath)
}

// renderTOCPage generates the HTML table of contents.
func (s *Server) renderTOCPage(w http.ResponseWriter, entries []TOCEntry) {
	var html strings.Builder
	html.WriteString(`<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>MarkSafe Browse</title>
<style>` + s.getCSS() + `</style>
</head>
<body class="dark">
<nav class="sidebar">
<h2>📚 Inhaltsverzeichnis</h2>
<button class="theme-toggle" onclick="toggleTheme()">🌓</button>
<ul class="toc">`)

	for _, entry := range entries {
		html.WriteString(s.renderTOCEntry(entry, 0))
	}

	html.WriteString(`</ul>
</nav>
<main class="content">
<h1>MarkSafe Browse</h1>
<p>Navigiere durch die Markdown-Dokumentation.</p>
<script>
function toggleTheme() {
  const body = document.body;
  body.className = body.className === 'dark' ? 'light' : 'dark';
}
</script>
</main>
</body>
</html>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html.String()))
}

// renderTOCEntry recursively renders a TOC entry as HTML.
func (s *Server) renderTOCEntry(entry TOCEntry, indent int) string {
	var html strings.Builder
	pad := strings.Repeat("&nbsp;&nbsp;&nbsp;", indent)

	if entry.IsDir {
		html.WriteString(fmt.Sprintf(`<li class="dir"><strong>%s%s</strong><ul>`, pad, escapeHTML(entry.Title)))
		for _, child := range entry.Children {
			html.WriteString(s.renderTOCEntry(child, indent+1))
		}
		html.WriteString("</ul></li>")
	} else {
		html.WriteString(fmt.Sprintf(`<li class="file"><a href="/render?path=%s">%s%s</a></li>`,
			escapeHTML(entry.Path), pad, escapeHTML(entry.Title)))
	}
	return html.String()
}

// renderMarkdownPage renders a markdown file as a full HTML page.
func (s *Server) renderMarkdownPage(w http.ResponseWriter, title, markdownContent string) {
	html := renderMarkdownToHTML(markdownContent)

	var page strings.Builder
	page.WriteString(`<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + escapeHTML(title) + ` — MarkSafe Browse</title>
<style>` + s.getCSS() + `</style>
</head>
<body class="dark">
<nav class="sidebar">
<h2>📚 MarkSafe Browse</h2>
<a href="/">← Zurück zur Übersicht</a>
<button class="theme-toggle" onclick="toggleTheme()">🌓</button>
</nav>
<main class="content">
` + html + `
</main>
</body>
</html>`)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(page.String()))
}

// getCSS returns the base CSS for the browse interface.
func (s *Server) getCSS() string {
	return `
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; transition: background 0.3s, color 0.3s; }
body.dark { background: #0f172a; color: #f1f5f9; }
body.light { background: #ffffff; color: #0f172a; }
.sidebar { position: fixed; left: 0; top: 0; width: 280px; height: 100vh; background: inherit; border-right: 1px solid #334155; padding: 20px; overflow-y: auto; }
body.light .sidebar { border-right-color: #e2e8f0; }
.content { margin-left: 300px; padding: 40px; max-width: 900px; }
.sidebar h2 { margin-bottom: 20px; font-size: 1.2rem; }
.sidebar a { color: inherit; text-decoration: none; }
.sidebar a:hover { text-decoration: underline; }
.theme-toggle { position: absolute; top: 20px; right: 20px; cursor: pointer; padding: 8px 16px; border-radius: 8px; border: 1px solid #334155; background: transparent; color: inherit; font-size: 1.2rem; }
.toc { list-style: none; padding-left: 0; }
.toc .dir { margin-left: 10px; }
.toc .file { margin: 4px 0; }
.toc .file a { color: #60a5fa; text-decoration: none; }
.toc .file a:hover { text-decoration: underline; }
.markdown h1 { font-size: 2rem; margin-bottom: 1rem; border-bottom: 2px solid #334155; padding-bottom: 0.5rem; }
.markdown h2 { font-size: 1.5rem; margin: 1.5rem 0 0.5rem; color: #60a5fa; }
.markdown h3 { font-size: 1.25rem; margin: 1rem 0 0.5rem; }
.markdown p { margin: 1rem 0; line-height: 1.7; }
.markdown code { background: #1e293b; padding: 2px 6px; border-radius: 4px; font-size: 0.9em; }
.markdown pre { background: #1e293b; padding: 16px; border-radius: 8px; overflow-x: auto; }
.markdown pre code { background: none; padding: 0; }
.markdown table { border-collapse: collapse; width: 100%; margin: 1rem 0; }
.markdown th, .markdown td { border: 1px solid #334155; padding: 8px 12px; text-align: left; }
.markdown th { background: #1e293b; }
.markdown blockquote { border-left: 4px solid #60a5fa; margin: 1rem 0; padding: 0.5rem 1rem; opacity: 0.8; }
.markdown a { color: #60a5fa; }
.markdown ul, .markdown ol { margin: 1rem 0; padding-left: 2rem; }
.markdown img { max-width: 100%; border-radius: 8px; }
`
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
// It strips YAML frontmatter before rendering.
func renderMarkdownToHTML(content string) string {
	content = stripFrontmatter(content)

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
	var html bytes.Buffer
	if err := md.Convert([]byte(content), &html); err != nil {
		return "<div class=\"markdown\"><p>Error rendering markdown: " + escapeHTML(err.Error()) + "</p></div>"
	}
	return "<div class=\"markdown\">" + html.String() + "</div>"
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
