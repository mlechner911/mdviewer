package main

// Templates for new documents (mlcext): the built-in set, copied once
// into the user's folder, plus the RefNo proposal from the document folder.

import (
	"fmt"
	"path/filepath"
	"sync"

	"marksafe/internal/mdbrowse"

	"github.com/mlechner911/mdviewer/mlcext"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// templateSeed copies the built-in templates into the user's template
// folder once per start (mlcext.SeedTemplates).
type templateSeed struct {
	once sync.Once
	ok   bool
}

// MLCTemplates lists the templates for new documents of the MLC Markdown
// Extension: the built-in starting set, replaced or extended by the
// schemas (*.json + Markdown body) in <document folder>/_vorlagen. New
// kinds or fields need a new file there, not a new program.
func (a *App) MLCTemplates() []mlcext.TemplateInfo {
	docs := a.documentRefs()
	list := a.loadTemplates()
	for i := range list {
		list[i] = list[i].WithDocuments(docs)
	}
	return list
}

// documentRefs are the documents of the document folder that docref
// fields can point to.
func (a *App) documentRefs() []mlcext.DocumentRef {
	root := a.documentRoot()
	if root == "" {
		return nil
	}
	r, err := mdbrowse.NewRoot(root)
	if err != nil {
		return nil
	}
	docs, err := r.Documents()
	if err != nil {
		return nil
	}
	refs := make([]mlcext.DocumentRef, 0, len(docs))
	for _, d := range docs {
		refs = append(refs, mlcext.DocumentRef{RefNo: d.RefNo, Title: d.Title})
	}
	return refs
}

// appTemplatesDir is the user's own template folder
// (<user config>/marksafe/vorlagen): the built-in templates are copied
// there once and can be changed without a new release.
func (a *App) appTemplatesDir() string {
	if a.config == nil {
		return ""
	}
	return filepath.Join(a.config.Dir(), "vorlagen")
}

// loadTemplates: the user's template folder (seeded from the built-in
// set; without it the built-in set itself), then _vorlagen in the document
// folder. A later template with the same id replaces an earlier one.
func (a *App) loadTemplates() []mlcext.TemplateInfo {
	dir := a.appTemplatesDir()
	a.templates.once.Do(func() {
		if dir == "" {
			return
		}
		written, err := mlcext.SeedTemplates(dir)
		if err != nil {
			logger().Error("Seeding templates failed", "dir", dir, "error", err)
			return
		}
		if len(written) > 0 {
			logger().Info("Templates copied to the user folder", "dir", dir, "files", written)
		}
		a.templates.ok = true
	})
	if !a.templates.ok {
		return mlcext.LoadTemplates(a.userTemplatesDir())
	}
	return mlcext.LoadTemplatesFrom(false, dir, a.userTemplatesDir())
}

// OpenTemplatesFolder shows the user's template folder in the file
// manager (created and seeded if needed).
func (a *App) OpenTemplatesFolder() error {
	a.loadTemplates()
	dir := a.appTemplatesDir()
	if dir == "" {
		return fmt.Errorf("no template folder")
	}
	return application.Get().Env.OpenFileManager(dir, false)
}

// TemplatesFolder is the path of the user's template folder, for display.
func (a *App) TemplatesFolder() string {
	return a.appTemplatesDir()
}

func (a *App) userTemplatesDir() string {
	if root := a.documentRoot(); root != "" {
		return filepath.Join(root, mdbrowse.TemplatesDir)
	}
	return ""
}

// RenderMLCTemplate fills template id with the form values and checks the
// result (field errors, diagnostics of the new document).
func (a *App) RenderMLCTemplate(id string, values map[string]any) mlcext.RenderResult {
	for _, t := range a.loadTemplates() {
		if t.ID == id {
			return t.WithDocuments(a.documentRefs()).Render(values)
		}
	}
	return mlcext.RenderResult{FieldErrors: map[string]string{"": "Vorlage nicht gefunden: " + id}, Diagnostics: []mlcext.Diagnostic{}}
}

// SuggestMLCRefNo proposes the next RefNo for prefix and area from the
// document folder, or else the documents next to docPath (only whitelisted
// directories are read). It is a proposal: only the Sensor Monitor knows
// which RefNos are taken.
func (a *App) SuggestMLCRefNo(prefix, area, docPath string) string {
	dir := a.documentRoot()
	if dir == "" && docPath != "" {
		if abs, err := filepath.Abs(docPath); err == nil && a.config.IsPathAllowed(filepath.Dir(abs)) {
			dir = filepath.Dir(abs)
		}
	}
	var existing []string
	if dir != "" {
		existing = mlcext.RefNosInDir(dir, 5000)
	}
	return mlcext.SuggestRefNo(prefix, area, existing)
}
