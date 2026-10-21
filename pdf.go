package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mlechner911/mdviewer/mlcext"
	"github.com/mlechner911/mdviewer/mlcext/mlcdoc"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// PDF output goes through the MLC PDF toolkit mlcdoc (mlcmd2pdf + Typst),
// which MarkSafe ships next to its executable. Preview and export work for
// any Markdown document; front matter mlcdoc does not know is left out.

// pdfTool finds mlcdoc once and keeps the client, and the remote images
// fetched for it (the preview renders again after every pause in typing).
type pdfTool struct {
	once   sync.Once
	client *mlcdoc.Client
	err    error

	mu     sync.Mutex
	images map[string][]byte // URL → content; nil entry: fetching failed
	size   int
}

const (
	remoteImageMax   = 20 << 20  // one image
	remoteImageCache = 100 << 20 // all images kept
)

// remoteImages fetches the http(s) images of md whose host is whitelisted
// (the same rule as the HTML preview): mlcdoc itself loads nothing from the
// network. Others stay out and become mlcdoc's placeholder.
func (a *App) remoteImages(md string) map[string][]byte {
	urls := mlcdoc.RemoteImages(md)
	if len(urls) == 0 {
		return nil
	}
	out := map[string][]byte{}
	for _, u := range urls {
		if !a.config.IsURLAllowed(u) {
			continue
		}
		if b := a.pdf.fetch(u); b != nil {
			out[u] = b
		}
	}
	return out
}

func (p *pdfTool) fetch(u string) []byte {
	p.mu.Lock()
	b, seen := p.images[u]
	p.mu.Unlock()
	if seen {
		return b
	}
	b = download(u)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.images == nil || p.size+len(b) > remoteImageCache {
		p.images, p.size = map[string][]byte{}, 0
	}
	p.images[u] = b
	p.size += len(b)
	return b
}

// download reads an image (at most remoteImageMax bytes, 15 s); nil when
// it is no image or cannot be loaded.
func download(u string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger().Warn("Image not loaded", "url", u, "error", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger().Warn("Image not loaded", "url", u, "status", resp.StatusCode)
		return nil
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, remoteImageMax+1))
	if err != nil || len(b) > remoteImageMax || !strings.HasPrefix(http.DetectContentType(b), "image/") && !bytes.Contains(b[:min(len(b), 512)], []byte("<svg")) {
		return nil
	}
	return b
}

func (p *pdfTool) get() (*mlcdoc.Client, error) {
	p.once.Do(func() {
		p.client, p.err = mlcdoc.New(context.Background(), mlcdoc.Options{})
	})
	return p.client, p.err
}

// PDFResult is the outcome of a preview or an export. Error and ErrorLine
// say why mlcdoc cannot render the document (ErrorLine 0: no line).
type PDFResult struct {
	Pages     []string `json:"pages"` // preview: PNG data URLs
	Path      string   `json:"path"`  // export: written file, "" if cancelled
	Warnings  []string `json:"warnings"`
	Ignored   []string `json:"ignored"` // front matter keys left out
	Error     string   `json:"error"`
	ErrorLine int      `json:"errorLine"`
}

func pdfError(err error) PDFResult {
	r := PDFResult{Pages: []string{}, Warnings: []string{}, Ignored: []string{}, Error: err.Error()}
	var re *mlcdoc.RenderError
	if errors.As(err, &re) {
		r.Error, r.ErrorLine = re.Message, re.Line
	}
	return r
}

// MLCDocVersion returns the mlcdoc in use ("mlcdoc 0.16.0 (…)"), or ""
// when it is missing or too old; PDF functions are hidden then.
func (a *App) MLCDocVersion() string {
	c, err := a.pdf.get()
	if err != nil {
		logger().Warn("mlcdoc unavailable", "error", err)
		return ""
	}
	return c.Version()
}

// baseDirOf is the directory of docPath for relative images, only when it
// is whitelisted.
func (a *App) baseDirOf(docPath string) string {
	if docPath == "" {
		return ""
	}
	abs, err := filepath.Abs(docPath)
	if err != nil || !a.config.IsPathAllowed(filepath.Dir(abs)) {
		return ""
	}
	return filepath.Dir(abs)
}

// pdfRequest is md for mlcdoc: a layout file with the page header (kind,
// id and version of documents of the MLC extension) and the watermark:
// watermark when given (for this PDF only), else the user's setting.
func (a *App) pdfRequest(md, docPath string, watermark ...string) mlcdoc.Request {
	wm := a.config.GetPDFWatermark()
	if len(watermark) > 0 && strings.TrimSpace(watermark[0]) != "" {
		wm = strings.TrimSpace(watermark[0])
	}
	return mlcdoc.Request{
		Markdown: md,
		BaseDir:  a.baseDirOf(docPath),
		Layout:   mlcext.LayoutYAML(md, mlcext.LayoutOptions{Watermark: wm}),
		Images:   a.remoteImages(md),
	}
}

// GetPDFPreviewNoImages: the PDF preview shows placeholders for images.
func (a *App) GetPDFPreviewNoImages() bool {
	return a.config.GetPDFPreviewNoImages()
}

// SetPDFPreviewNoImages switches placeholders in the PDF preview; export
// and print always embed the images.
func (a *App) SetPDFPreviewNoImages(v bool) error {
	return a.config.SetPDFPreviewNoImages(v)
}

// GetPDFWatermark is the watermark for PDFs from MarkSafe ("" for none).
func (a *App) GetPDFWatermark() string {
	return a.config.GetPDFWatermark()
}

// SetPDFWatermark stores the watermark for PDFs (export, preview, print).
func (a *App) SetPDFWatermark(text string) error {
	return a.config.SetPDFWatermark(strings.TrimSpace(text))
}

// PreviewPDF renders md into page images, exactly as the exported PDF.
func (a *App) PreviewPDF(md string, docPath string) PDFResult {
	c, err := a.pdf.get()
	if err != nil {
		return pdfError(err)
	}
	req := a.pdfRequest(md, docPath)
	if a.config.GetPDFPreviewNoImages() {
		req.NoImages, req.Images = true, nil
	}
	res, err := c.RenderPages(context.Background(), req, 110)
	if err != nil {
		return pdfError(err)
	}
	out := PDFResult{Pages: make([]string, len(res.Images)), Warnings: res.Warnings, Ignored: nonNil(res.Ignored)}
	for i, p := range res.Images {
		out.Pages[i] = "data:image/png;base64," + base64.StdEncoding.EncodeToString(p)
	}
	return out
}

// ExportPDF renders md to a PDF and asks where to save it (next to the
// document, else in the document folder). watermark, when not empty,
// replaces the setting for this PDF (the suggested "ENTWURF").
func (a *App) ExportPDF(md string, docPath string, watermark string) PDFResult {
	c, err := a.pdf.get()
	if err != nil {
		return pdfError(err)
	}
	doc, err := c.RenderPDF(context.Background(), a.pdfRequest(md, docPath, watermark))
	if err != nil {
		return pdfError(err)
	}
	name, dir := "document.pdf", a.documentRoot()
	if docPath != "" {
		base := filepath.Base(docPath)
		name = strings.TrimSuffix(base, filepath.Ext(base)) + ".pdf"
		if d := a.baseDirOf(docPath); d != "" {
			dir = d
		}
	}
	// Passed-on exports carry id and version: INFO-QM-001_v1.0.pdf.
	if s := mlcext.SuggestFileName(md, ".pdf", true); s != "" {
		name = s
	}
	dlg := application.Get().Dialog.SaveFile().
		SetMessage("Als PDF exportieren").
		SetFilename(name).
		AddFilter("PDF (*.pdf)", "*.pdf")
	if dir != "" {
		dlg.SetDirectory(dir)
	}
	path, err := dlg.PromptForSingleSelection()
	if err != nil {
		return pdfError(err)
	}
	res := PDFResult{Pages: []string{}, Warnings: doc.Warnings, Ignored: nonNil(doc.Ignored)}
	if path == "" {
		return res // cancelled
	}
	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		path += ".pdf"
	}
	if err := os.WriteFile(path, doc.PDF, 0o644); err != nil {
		return pdfError(err)
	}
	res.Path = path
	return res
}

// PrintPDF prints through the PDF: md is rendered like the export and
// opened in the system's PDF viewer, which prints it. The page then looks
// exactly like the exported PDF (and like the Sensor Monitor's).
// watermark as with ExportPDF.
func (a *App) PrintPDF(md string, docPath string, watermark string) PDFResult {
	c, err := a.pdf.get()
	if err != nil {
		return pdfError(err)
	}
	doc, err := c.RenderPDF(context.Background(), a.pdfRequest(md, docPath, watermark))
	if err != nil {
		return pdfError(err)
	}
	dir := filepath.Join(os.TempDir(), "marksafe-print")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return pdfError(err)
	}
	name := mlcext.SuggestFileName(md, ".pdf", true)
	if name == "" {
		name = "Dokument.pdf"
		if docPath != "" {
			base := filepath.Base(docPath)
			name = strings.TrimSuffix(base, filepath.Ext(base)) + ".pdf"
		}
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, doc.PDF, 0o600); err != nil {
		return pdfError(err)
	}
	if err := application.Get().Browser.OpenFile(path); err != nil {
		return pdfError(err)
	}
	return PDFResult{Pages: []string{}, Path: path, Warnings: doc.Warnings, Ignored: nonNil(doc.Ignored)}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
