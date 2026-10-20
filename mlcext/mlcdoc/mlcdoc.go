// Package mlcdoc runs the MLC PDF toolkit (the external mlcdoc binary) for
// documents written with the MLC Markdown Extension (non-standard, see
// docs/MLC-EXTENSIONS.md): Markdown to PDF and page images, so a preview
// shows exactly what the PDF will look like. Nothing is re-implemented here;
// without the binary the functions report ErrNotFound.
package mlcdoc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ErrNotFound: no mlcdoc binary.
var ErrNotFound = errors.New("mlcdoc nicht gefunden")

// Binary returns the path of the mlcdoc binary: explicit, $MLCDOC_PATH,
// $PATH, then the installer's default location (Windows:
// %ProgramFiles%\MLC\mlcdoc). Empty when there is none.
func Binary(explicit string) string {
	for _, p := range []string{explicit, os.Getenv("MLCDOC_PATH")} {
		if p != "" {
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
	}
	if p, err := exec.LookPath("mlcdoc"); err == nil {
		return p
	}
	for _, p := range defaultLocations() {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// Version returns "mlcdoc 0.16.0 (mlcmd2pdf 3.26.0 …)" or "" without one.
func Version(ctx context.Context, binary string) string {
	bin := Binary(binary)
	if bin == "" {
		return ""
	}
	cmd := exec.CommandContext(ctx, bin, "--version")
	hideConsole(cmd)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Request is one preview.
type Request struct {
	Markdown string
	// BaseDir is the directory of the document; relative images inside it
	// are copied next to the temporary copy (mlcdoc only reads inside its
	// root). Empty for unsaved documents.
	BaseDir string
	// DPI of the page images (default 110).
	DPI int
	// Binary overrides the mlcdoc location.
	Binary string
	// Timeout (default 60 s).
	Timeout time.Duration
}

// Result of a preview.
type Result struct {
	Pages    [][]byte // PNG per page
	Warnings []string
	Title    string
	Renderer string
}

// Preview renders md to PDF with mlcdoc and returns the pages as PNG.
// A document mlcdoc rejects (e.g. an unclosed form field) returns its
// message as the error.
func Preview(ctx context.Context, req Request) (*Result, error) {
	bin := Binary(req.Binary)
	if bin == "" {
		return nil, ErrNotFound
	}
	if req.DPI <= 0 {
		req.DPI = 110
	}
	if req.Timeout <= 0 {
		req.Timeout = 60 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	dir, err := os.MkdirTemp("", "mlcdoc-preview-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte(req.Markdown), 0o600); err != nil {
		return nil, err
	}
	copyImages(req.Markdown, req.BaseDir, dir)

	script := []map[string]any{
		{"tool": "md_to_pdf", "args": map[string]any{"input": "doc.md", "output": "doc.pdf"}},
		{"tool": "doc_open", "args": map[string]any{"path": "doc.pdf"}},
		{"tool": "doc_export_images", "args": map[string]any{"doc": 1, "folder": "pages", "dpi": req.DPI}},
	}
	sj, _ := json.Marshal(script)
	if err := os.WriteFile(filepath.Join(dir, "script.json"), sj, 0o600); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, bin, "run", "--script", "script.json", "--root", dir)
	cmd.Dir = dir
	hideConsole(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("mlcdoc: Zeitüberschreitung nach %s", req.Timeout)
		}
		msg := strings.TrimSpace(stderr.String())
		msg = strings.TrimPrefix(msg, "mlcdoc: ")
		msg = strings.TrimPrefix(msg, "step 1 (md_to_pdf): ")
		if msg == "" {
			msg = err.Error()
		}
		return nil, errors.New(msg)
	}

	res := &Result{}
	// The first output line is md_to_pdf's {pages, warnings, title, renderer}.
	if line, _, _ := strings.Cut(stdout.String(), "\n"); line != "" {
		var info struct {
			Warnings []string `json:"warnings"`
			Title    string   `json:"title"`
			Renderer string   `json:"renderer"`
		}
		if json.Unmarshal([]byte(line), &info) == nil {
			res.Warnings, res.Title, res.Renderer = info.Warnings, info.Title, info.Renderer
		}
	}
	files, _ := filepath.Glob(filepath.Join(dir, "pages", "*.png"))
	sort.Slice(files, func(i, j int) bool { return pageNo(files[i]) < pageNo(files[j]) })
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		res.Pages = append(res.Pages, b)
	}
	return res, nil
}

var pageRe = regexp.MustCompile(`_page_(\d+)\.png$`)

func pageNo(path string) int {
	n := 0
	if m := pageRe.FindStringSubmatch(path); m != nil {
		fmt.Sscan(m[1], &n)
	}
	return n
}

var imageRe = regexp.MustCompile(`!\[[^\]]*\]\(\s*<?([^)\s>]+)>?`)

// copyImages copies relative images of md from baseDir into dir, keeping
// their relative path. Anything outside baseDir is skipped.
func copyImages(md, baseDir, dir string) {
	if baseDir == "" {
		return
	}
	base, err := filepath.Abs(baseDir)
	if err != nil {
		return
	}
	for _, m := range imageRe.FindAllStringSubmatch(md, -1) {
		ref := m[1]
		if strings.Contains(ref, ":") || strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, "#") {
			continue // URL, data:, absolute path
		}
		rel := filepath.Clean(filepath.FromSlash(ref))
		if rel == "." || strings.HasPrefix(rel, "..") {
			continue
		}
		src := filepath.Join(base, rel)
		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		dst := filepath.Join(dir, rel)
		if os.MkdirAll(filepath.Dir(dst), 0o700) == nil {
			_ = os.WriteFile(dst, data, 0o600)
		}
	}
}
