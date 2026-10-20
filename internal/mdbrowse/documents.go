package mdbrowse

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mlechner911/mdviewer/mlcext"
)

// TemplatesDir is the folder of a document root that holds the user's own
// templates. It is not listed among the documents (templates carry
// placeholders and would only show errors).
const TemplatesDir = "_vorlagen"

// DocumentInfo describes one document of the MLC Markdown Extension
// (non-standard, docs/MLC-EXTENSIONS.md) in the root, read by mlcext — the
// same code the Sensor Monitor uses. Status, version history and approval
// live in the Sensor Monitor only and are deliberately not here.
type DocumentInfo struct {
	Path      string `json:"path"` // root-relative, slash-separated
	RefNo     string `json:"refNo"`
	Kind      string `json:"kind"`
	KindLabel string `json:"kindLabel"`
	Version   string `json:"version"`
	Title     string `json:"title"`
	Fields    int    `json:"fields"`
	Errors    int    `json:"errors"`
	Warnings  int    `json:"warnings"`
}

// Documents lists the MLC documents below the root, ordered by RefNo
// (documents without one last, by path).
func (r *Root) Documents() ([]DocumentInfo, error) {
	files, err := r.GetMarkdownFiles()
	if err != nil {
		return nil, err
	}
	docs := []DocumentInfo{}
	for _, f := range files {
		rel, err := filepath.Rel(r.Dir, f)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		if rel == TemplatesDir || strings.HasPrefix(rel, TemplatesDir+"/") {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		md := string(b)
		if !mlcext.Detect(md) {
			continue
		}
		h := mlcext.ParseHeader(md)
		fm, _ := mlcext.ParseFrontMatter(md)
		d := DocumentInfo{
			Path: rel, RefNo: h.RefNo, Kind: h.Kind, KindLabel: mlcext.KindLabel(h.Kind),
			Version: h.Version, Title: fm.Title,
		}
		if d.Version == "" {
			d.Version = fm.Version
		}
		if d.Title == "" {
			d.Title = h.Title // "**Titel:** …" of SOPs
		}
		if d.Title == "" {
			d.Title, _ = ExtractTitle(f)
		}
		fields, _ := mlcext.ExtractFields(md)
		d.Fields = len(fields)
		for _, diag := range mlcext.Check(md, mlcext.Options{RequireHeader: true}) {
			switch diag.Severity {
			case mlcext.Error:
				d.Errors++
			case mlcext.Warning:
				d.Warnings++
			}
		}
		docs = append(docs, d)
	}
	sort.SliceStable(docs, func(i, j int) bool {
		a, b := docs[i], docs[j]
		if (a.RefNo == "") != (b.RefNo == "") {
			return a.RefNo != ""
		}
		if a.RefNo != b.RefNo {
			return a.RefNo < b.RefNo
		}
		return a.Path < b.Path
	})
	return docs, nil
}
