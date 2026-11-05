package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mlechner911/mdviewer/mlcext"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// maxImageBytes is the largest image MarkSafe embeds.
const maxImageBytes = 50 << 20

// pastedImage: the image types the PDF reads as a data: URL.
var pastedImage = regexp.MustCompile(`^data:image/(png|jpeg|gif);base64,`)

// ImageChoice is an image file for the image dialog.
type ImageChoice struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	// Reference: inside the document's folder, so a reference also works in
	// the PDF, which reads nothing outside it. Otherwise the image is embedded.
	Reference string `json:"reference"`
	// Embeddable: false for WebP, which the PDF reads only from a file.
	Embeddable bool `json:"embeddable"`
}

// ImageSpec is what the image dialog inserts: a file, or data pasted from the
// clipboard (DataURI).
type ImageSpec struct {
	Path    string `json:"path"`
	DataURI string `json:"dataUri"`
	Alt     string `json:"alt"`
	Embed   bool   `json:"embed"`
}

// EmbedPlan is what embedding the images of a document would do.
type EmbedPlan struct {
	Images  []string `json:"images"`
	Bytes   int64    `json:"bytes"`
	Missing []string `json:"missing"`
}

// EmbedResult is the document with its images embedded.
type EmbedResult struct {
	Markdown string   `json:"markdown"`
	Embedded []string `json:"embedded"`
	Missing  []string `json:"missing"`
}

// ChooseImage lets the user pick an image file; "" when cancelled.
func (a *App) ChooseImage(docPath string) (string, error) {
	dlg := application.Get().Dialog.OpenFile().
		SetTitle("Bild wählen").
		AddFilter("Bilder", "*.png;*.jpg;*.jpeg;*.gif;*.webp;*.svg")
	if docPath != "" {
		dlg.SetDirectory(filepath.Dir(docPath))
	} else if dir := a.documentRoot(); dir != "" {
		dlg.SetDirectory(dir)
	}
	return promptPath(dlg)
}

// InspectImage checks a file the user chose or dropped.
func (a *App) InspectImage(path, docPath string) (*ImageChoice, error) {
	data, err := readImage(path)
	if err != nil {
		return nil, err
	}
	embeddable := mlcext.ImageDataURI(data, path) != ""
	reference := referenceTo(path, docPath)
	if !embeddable && !mlcext.IsWebP(data) {
		return nil, fmt.Errorf("%s ist kein unterstütztes Bild (PNG, JPEG, GIF, WebP, SVG).", filepath.Base(path))
	}
	if !embeddable && reference == "" {
		return nil, fmt.Errorf("%s: WebP kann nicht eingebettet werden. Bitte in den Dokumentordner legen oder als PNG oder JPEG speichern.", filepath.Base(path))
	}
	return &ImageChoice{Path: path, Name: filepath.Base(path), Size: int64(len(data)), Reference: reference, Embeddable: embeddable}, nil
}

// ImageMarkdown writes the image of the dialog.
func (a *App) ImageMarkdown(spec ImageSpec, docPath string) (string, error) {
	if spec.DataURI != "" {
		if !pastedImage.MatchString(spec.DataURI) {
			return "", errors.New("Die Zwischenablage enthält kein unterstütztes Bild (PNG, JPEG, GIF).")
		}
		return mlcext.FormatImage(spec.Alt, spec.DataURI, ""), nil
	}
	if !spec.Embed {
		ref := referenceTo(spec.Path, docPath)
		if ref == "" {
			return "", errors.New("Das Bild liegt außerhalb des Dokumentordners und kann nur eingebettet werden.")
		}
		return mlcext.FormatImage(spec.Alt, ref, ""), nil
	}
	data, err := readImage(spec.Path)
	if err != nil {
		return "", err
	}
	uri := mlcext.ImageDataURI(data, spec.Path)
	if uri == "" {
		return "", fmt.Errorf("%s ist kein unterstütztes Bild.", filepath.Base(spec.Path))
	}
	return mlcext.FormatImage(spec.Alt, uri, filepath.Base(spec.Path)), nil
}

// PlanEmbedImages says what EmbedImages would do; docPath is where relative
// images start from.
func (a *App) PlanEmbedImages(md, docPath string) EmbedPlan {
	plan := EmbedPlan{Images: []string{}, Missing: []string{}}
	for _, t := range mlcext.ImageTargets(md) {
		data, ok := a.loadImageTarget(t, docPath)
		if !ok || mlcext.ImageDataURI(data, t) == "" {
			plan.Missing = append(plan.Missing, t)
			continue
		}
		plan.Images = append(plan.Images, t)
		plan.Bytes += int64(len(data))
	}
	return plan
}

// EmbedImages embeds the images of md that exist and are allowed.
func (a *App) EmbedImages(md, docPath string) EmbedResult {
	out, embedded, failed := mlcext.EmbedImages(md, func(t string) ([]byte, bool) {
		return a.loadImageTarget(t, docPath)
	})
	return EmbedResult{Markdown: out, Embedded: nonNil(embedded), Missing: nonNil(failed)}
}

// loadImageTarget reads an image of a document under the rules of the
// preview: allowed domains and allowed folders only.
func (a *App) loadImageTarget(target, docPath string) ([]byte, bool) {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		if !a.config.IsURLAllowed(target) {
			return nil, false
		}
		b := a.pdf.fetch(target)
		return b, b != nil
	}
	path := target
	if p, err := url.PathUnescape(path); err == nil {
		path = p
	}
	if !filepath.IsAbs(path) {
		if docPath == "" {
			return nil, false
		}
		path = filepath.Join(filepath.Dir(docPath), filepath.FromSlash(path))
	}
	if !a.config.IsPathAllowed(path) {
		return nil, false
	}
	data, err := readImage(path)
	return data, err == nil
}

func readImage(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("%s kann nicht gelesen werden.", filepath.Base(path))
	}
	if info.Size() > maxImageBytes {
		return nil, fmt.Errorf("%s ist größer als %d MB.", filepath.Base(path), maxImageBytes>>20)
	}
	return os.ReadFile(path)
}

// referenceTo is path relative to the document's folder with forward slashes,
// "" when it is not inside that folder.
func referenceTo(path, docPath string) string {
	if docPath == "" {
		return ""
	}
	rel, err := filepath.Rel(filepath.Dir(docPath), path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ""
	}
	return filepath.ToSlash(rel)
}
