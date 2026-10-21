// Package mlcdoc is the Go interface to the MLC PDF toolkit: the external
// mlcdoc binary (mlcmd2pdf + Typst), which MarkSafe ships next to itself.
// It turns Markdown into PDF or page images, without a browser or network.
//
// mlcdoc is a separate program, not a library: this package finds it,
// checks its version, prepares the Markdown (front matter), runs it in a
// temporary directory and returns typed results and errors. Nothing of the
// PDF layout is re-implemented here.
//
//	c, err := mlcdoc.New(ctx, mlcdoc.Options{})
//	doc, err := c.RenderPDF(ctx, mlcdoc.Request{Markdown: md, BaseDir: dir})
//	var re *mlcdoc.RenderError
//	if errors.As(err, &re) { … re.Line, re.Message … }
package mlcdoc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MinVersion is the oldest mlcdoc this package works with.
const MinVersion = "0.16.0"

var (
	// ErrNotFound: no mlcdoc binary.
	ErrNotFound = errors.New("mlcdoc nicht gefunden")
	// ErrTooOld: the binary is older than MinVersion.
	ErrTooOld = errors.New("mlcdoc ist zu alt")
)

// RenderError is a document mlcdoc rejects, e.g. an unclosed form field or
// an invalid front matter. Line is 0 when mlcdoc names none.
type RenderError struct {
	Line    int
	Message string
}

func (e *RenderError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("Zeile %d: %s", e.Line, e.Message)
	}
	return e.Message
}

// Options configure a Client.
type Options struct {
	// Binary overrides the search (see Binary).
	Binary string
	// Timeout per run (default 60 s).
	Timeout time.Duration
	// MaxParallel limits concurrent runs (default 2).
	MaxParallel int
}

// Client runs one mlcdoc binary.
type Client struct {
	bin     string
	version string
	timeout time.Duration
	sem     chan struct{}
}

// New finds mlcdoc and checks its version.
func New(ctx context.Context, opt Options) (*Client, error) {
	bin := Binary(opt.Binary)
	if bin == "" {
		return nil, ErrNotFound
	}
	vctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(vctx, bin, "--version")
	hideConsole(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("mlcdoc --version: %w", err)
	}
	version := strings.TrimSpace(string(out))
	if v := semver(version); v == nil || compare(v, semver("mlcdoc "+MinVersion)) < 0 {
		return nil, fmt.Errorf("%w: %s, mindestens %s nötig", ErrTooOld, version, MinVersion)
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 60 * time.Second
	}
	if opt.MaxParallel <= 0 {
		opt.MaxParallel = 2
	}
	return &Client{bin: bin, version: version, timeout: opt.Timeout, sem: make(chan struct{}, opt.MaxParallel)}, nil
}

// Path is the binary in use.
func (c *Client) Path() string { return c.bin }

// Version is "mlcdoc 0.16.0 (mlcmd2pdf 3.26.0 …)".
func (c *Client) Version() string { return c.version }

// Request is one document.
type Request struct {
	Markdown string
	// BaseDir is the document's directory; relative images inside it are
	// copied next to the temporary copy (mlcdoc reads only inside its
	// root). Empty for unsaved documents.
	BaseDir string
	// Images are remote images the caller fetched (URL → file content):
	// mlcdoc loads nothing from the network, so they are written next to
	// the document and the URLs point there. The caller decides which URLs
	// may be fetched (MarkSafe: its whitelist).
	Images map[string][]byte
	// NoImages replaces every image by mlcdoc's placeholder ("Abbildung
	// nicht verfügbar: …"), for a fast preview of image-heavy documents.
	NoImages bool
	// Layout is a YAML layout file (mlcext.LayoutYAML: header with id and
	// version, watermark …) that mlcdoc lays over the document's front
	// matter; the document stays unchanged. Empty for none.
	Layout string
}

// Document is a rendered PDF.
type Document struct {
	PDF      []byte
	Pages    int
	Title    string
	Renderer string
	// Warnings from mlcdoc (missing image, missing glyph …).
	Warnings []string
	// Ignored front matter keys mlcdoc does not know (see Prepare).
	Ignored []string
}

// Pages are rendered page images.
type Pages struct {
	Images   [][]byte // PNG per page
	Warnings []string
	Ignored  []string
}

// RenderPDF renders req to a PDF.
func (c *Client) RenderPDF(ctx context.Context, req Request) (*Document, error) {
	md, ignored := Prepare(req.Markdown)
	md, files := imageFiles(md, req)
	var doc *Document
	err := c.run(ctx, md, req.BaseDir, req.Layout, files, []map[string]any{
		{"tool": "md_to_pdf", "args": mdToPDFArgs(req.Layout)},
	}, func(dir string, info stepInfo) error {
		b, err := os.ReadFile(filepath.Join(dir, "doc.pdf"))
		if err != nil {
			return err
		}
		doc = &Document{PDF: b, Pages: info.Pages, Title: info.Title, Renderer: info.Renderer, Warnings: info.Warnings, Ignored: ignored}
		return nil
	})
	return doc, err
}

// RenderPages renders req to one PNG per page at dpi (default 110).
func (c *Client) RenderPages(ctx context.Context, req Request, dpi int) (*Pages, error) {
	if dpi <= 0 {
		dpi = 110
	}
	md, ignored := Prepare(req.Markdown)
	md, files := imageFiles(md, req)
	var pages *Pages
	err := c.run(ctx, md, req.BaseDir, req.Layout, files, []map[string]any{
		{"tool": "md_to_pdf", "args": mdToPDFArgs(req.Layout)},
		{"tool": "doc_open", "args": map[string]any{"path": "doc.pdf"}},
		{"tool": "doc_export_images", "args": map[string]any{"doc": 1, "folder": "pages", "dpi": dpi}},
	}, func(dir string, info stepInfo) error {
		files, _ := filepath.Glob(filepath.Join(dir, "pages", "*.png"))
		sort.Slice(files, func(i, j int) bool { return pageNo(files[i]) < pageNo(files[j]) })
		pages = &Pages{Warnings: withoutPlaceholderWarnings(info.Warnings), Ignored: ignored}
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			pages.Images = append(pages.Images, b)
		}
		return nil
	})
	return pages, err
}

// stepInfo is md_to_pdf's JSON line on stdout.
type stepInfo struct {
	Pages    int      `json:"pages"`
	Title    string   `json:"title"`
	Renderer string   `json:"renderer"`
	Warnings []string `json:"warnings"`
}

// mdToPDFArgs: the md_to_pdf step, with the layout file when there is one.
func mdToPDFArgs(layout string) map[string]any {
	args := map[string]any{"input": "doc.md", "output": "doc.pdf"}
	if layout != "" {
		args["layout"] = "layout.yaml"
	}
	return args
}

// run writes md (and the layout file) into a fresh directory, runs the
// script there and hands the directory to collect before removing it.
func (c *Client) run(ctx context.Context, md, baseDir, layout string, files map[string][]byte, script []map[string]any, collect func(string, stepInfo) error) error {
	select {
	case c.sem <- struct{}{}:
		defer func() { <-c.sem }()
	case <-ctx.Done():
		return ctx.Err()
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	dir, err := os.MkdirTemp("", "mlcdoc-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte(md), 0o600); err != nil {
		return err
	}
	if layout != "" {
		if err := os.WriteFile(filepath.Join(dir, "layout.yaml"), []byte(layout), 0o600); err != nil {
			return err
		}
	}
	for name, b := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(p, b, 0o600); err != nil {
			return err
		}
	}
	copyImages(md, baseDir, dir)
	sj, _ := json.Marshal(script)
	if err := os.WriteFile(filepath.Join(dir, "script.json"), sj, 0o600); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, c.bin, "run", "--script", "script.json", "--root", dir)
	cmd.Dir = dir
	hideConsole(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("mlcdoc: Zeitüberschreitung nach %s", c.timeout)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return parseError(stderr.String(), err)
	}
	var info stepInfo
	if line, _, _ := strings.Cut(stdout.String(), "\n"); line != "" {
		_ = json.Unmarshal([]byte(line), &info)
	}
	if info.Warnings == nil {
		info.Warnings = []string{}
	}
	return collect(dir, info)
}

var (
	stepPrefixRe = regexp.MustCompile(`^step \d+ \([a-z_]+\): `)
	lineRe       = regexp.MustCompile(`\bin line (\d+): (.*)$`)
)

// parseError turns mlcdoc's stderr into a RenderError.
func parseError(stderr string, runErr error) error {
	msg := strings.TrimSpace(stderr)
	if msg == "" {
		return fmt.Errorf("mlcdoc: %w", runErr)
	}
	if i := strings.LastIndex(msg, "\n"); i >= 0 {
		msg = strings.TrimSpace(msg[i+1:]) // the last line carries the reason
	}
	msg = strings.TrimPrefix(msg, "mlcdoc: ")
	msg = stepPrefixRe.ReplaceAllString(msg, "")
	re := &RenderError{Message: msg}
	if m := lineRe.FindStringSubmatch(msg); m != nil {
		re.Line, _ = strconv.Atoi(m[1])
		re.Message = m[2]
	}
	return re
}

// Binary returns the path of the mlcdoc binary: explicit, $MLCDOC_PATH,
// next to the running program (where MarkSafe ships it), $PATH, then the
// installer's default location (Windows: %ProgramFiles%\MLC\mlcdoc). Empty
// when there is none.
func Binary(explicit string) string {
	for _, p := range []string{explicit, os.Getenv("MLCDOC_PATH")} {
		if p != "" && isFile(p) {
			return p
		}
	}
	if self, err := os.Executable(); err == nil {
		for _, name := range []string{"mlcdoc", "mlcdoc.exe"} {
			if p := filepath.Join(filepath.Dir(self), name); isFile(p) {
				return p
			}
		}
	}
	if p, err := exec.LookPath("mlcdoc"); err == nil {
		return p
	}
	for _, p := range defaultLocations() {
		if isFile(p) {
			return p
		}
	}
	return ""
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

var versionRe = regexp.MustCompile(`mlcdoc (\d+)\.(\d+)\.(\d+)`)

func semver(s string) []int {
	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	v := make([]int, 3)
	for i := range v {
		v[i], _ = strconv.Atoi(m[i+1])
	}
	return v
}

func compare(a, b []int) int {
	for i := range a {
		if a[i] != b[i] {
			return a[i] - b[i]
		}
	}
	return 0
}

var pageRe = regexp.MustCompile(`_page_(\d+)\.png$`)

func pageNo(path string) int {
	n := 0
	if m := pageRe.FindStringSubmatch(path); m != nil {
		n, _ = strconv.Atoi(m[1])
	}
	return n
}

var imageRe = regexp.MustCompile(`!\[[^\]]*\]\(\s*<?([^)\s>]+)>?`)

// noImageTarget is the image path used for placeholders; it never exists.
const noImageTarget = "mlc-ohne-bild"

// RemoteImages are the http(s) image URLs of md (Markdown images), for a
// caller that fetches them before rendering.
func RemoteImages(md string) []string {
	var urls []string
	seen := map[string]bool{}
	for _, m := range imageRe.FindAllStringSubmatch(md, -1) {
		u := m[1]
		if (strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")) && !seen[u] {
			seen[u] = true
			urls = append(urls, u)
		}
	}
	return urls
}

// imageFiles applies req.Images and req.NoImages to the prepared md: the
// files to write next to it and md pointing at them.
func imageFiles(md string, req Request) (string, map[string][]byte) {
	if req.NoImages {
		return imageRe.ReplaceAllStringFunc(md, func(m string) string {
			sm := imageRe.FindStringSubmatchIndex(m)
			return m[:sm[2]] + noImageTarget + m[sm[3]:]
		}), nil
	}
	if len(req.Images) == 0 {
		return md, nil
	}
	files := map[string][]byte{}
	names := map[string]string{}
	for u, b := range req.Images {
		sum := sha256.Sum256([]byte(u))
		name := "remote/" + hex.EncodeToString(sum[:8]) + imageExt(b)
		files[name] = b
		names[u] = name
	}
	md = imageRe.ReplaceAllStringFunc(md, func(m string) string {
		sm := imageRe.FindStringSubmatchIndex(m)
		if name, ok := names[m[sm[2]:sm[3]]]; ok {
			return m[:sm[2]] + name + m[sm[3]:]
		}
		return m
	})
	return md, files
}

// withoutPlaceholderWarnings drops mlcdoc's "skipped missing image" for the
// placeholders NoImages asked for.
func withoutPlaceholderWarnings(ws []string) []string {
	out := ws[:0:0]
	for _, w := range ws {
		if !strings.Contains(w, noImageTarget) {
			out = append(out, w)
		}
	}
	return out
}

// imageExt is the file extension for image data (mlcdoc looks at it).
func imageExt(b []byte) string {
	switch http.DetectContentType(b) {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	}
	if bytes.Contains(b[:min(len(b), 512)], []byte("<svg")) {
		return ".svg"
	}
	return ".png"
}

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
		data, err := os.ReadFile(filepath.Join(base, rel))
		if err != nil {
			continue
		}
		dst := filepath.Join(dir, rel)
		if os.MkdirAll(filepath.Dir(dst), 0o700) == nil {
			_ = os.WriteFile(dst, data, 0o600)
		}
	}
}
