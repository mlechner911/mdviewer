package main

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mlechner911/mdviewer/mlcext"
	"github.com/mlechner911/mdviewer/mlcext/mlcdoc"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// PDF output goes through the MLC PDF toolkit mlcdoc (mlcmd2pdf + Typst),
// which MarkSafe ships next to its executable. Preview and export work for
// any Markdown document; front matter mlcdoc does not know is left out.

// pdfTool finds mlcdoc once and keeps the client.
type pdfTool struct {
	once   sync.Once
	client *mlcdoc.Client
	err    error
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
// id and version of documents of the MLC extension) and the watermark of
// the user's setting.
func (a *App) pdfRequest(md, docPath string) mlcdoc.Request {
	return mlcdoc.Request{
		Markdown: md,
		BaseDir:  a.baseDirOf(docPath),
		Layout:   mlcext.LayoutYAML(md, mlcext.LayoutOptions{Watermark: a.config.GetPDFWatermark()}),
	}
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
	res, err := c.RenderPages(context.Background(), a.pdfRequest(md, docPath), 110)
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
// document, else in the document folder).
func (a *App) ExportPDF(md string, docPath string) PDFResult {
	c, err := a.pdf.get()
	if err != nil {
		return pdfError(err)
	}
	doc, err := c.RenderPDF(context.Background(), a.pdfRequest(md, docPath))
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
func (a *App) PrintPDF(md string, docPath string) PDFResult {
	c, err := a.pdf.get()
	if err != nil {
		return pdfError(err)
	}
	doc, err := c.RenderPDF(context.Background(), a.pdfRequest(md, docPath))
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
