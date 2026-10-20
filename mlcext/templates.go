package mlcext

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Templates for new documents of the MLC Markdown Extension. The files
// carry {{TITEL}}, {{TITEL_YAML}}, {{KENNUNG}} and {{VERSION}}; NewDocument
// fills them, so a created document contains no placeholder syntax. Every
// created document passes Check without errors (tests).
//
//go:embed templates/*.md
var templateFS embed.FS

// Template describes one document kind.
type Template struct {
	Kind   string `json:"kind"`   // sop | va | form | info
	Prefix string `json:"prefix"` // SOP | VA | FB | INFO
	Label  string `json:"label"`
}

// Templates lists the built-in templates.
func Templates() []Template {
	return []Template{
		{"sop", "SOP", KindLabel("sop")},
		{"va", "VA", KindLabel("va")},
		{"form", "FB", KindLabel("form")},
		{"info", "INFO", KindLabel("info")},
	}
}

func templateFor(kind string) (Template, bool) {
	for _, t := range Templates() {
		if t.Kind == kind {
			return t, true
		}
	}
	return Template{}, false
}

// NewParams are the inputs for a new document.
type NewParams struct {
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	RefNo   string `json:"refNo"`
	Version string `json:"version"` // default "1.0"
}

// NewDocument creates the Markdown of a new document from the built-in
// template of p.Kind. The RefNo must carry the kind's prefix (FB-… for a
// form), because the Sensor Monitor derives the kind from it.
func NewDocument(p NewParams) (string, error) {
	t, ok := templateFor(p.Kind)
	if !ok {
		return "", fmt.Errorf("unbekannte Dokumentart %q", p.Kind)
	}
	title := strings.Join(strings.Fields(p.Title), " ")
	if title == "" {
		return "", errors.New("Titel fehlt")
	}
	ref := strings.TrimSpace(p.RefNo)
	if ref == "" {
		return "", errors.New("Kennung fehlt")
	}
	if KindFromRef(ref) != t.Kind {
		return "", fmt.Errorf("Kennung %q passt nicht zur Art: erwartet %s-…", ref, t.Prefix)
	}
	if strings.ContainsAny(ref, "|*\n") {
		return "", fmt.Errorf("Kennung %q enthält unzulässige Zeichen", ref)
	}
	version := strings.TrimSpace(p.Version)
	if version == "" {
		version = "1.0"
	}
	raw, err := templateFS.ReadFile("templates/" + t.Kind + ".md")
	if err != nil {
		return "", err
	}
	yamlTitle, err := yaml.Marshal(title)
	if err != nil {
		return "", err
	}
	return strings.NewReplacer(
		"{{TITEL_YAML}}", strings.TrimSpace(string(yamlTitle)),
		"{{TITEL}}", title,
		"{{KENNUNG}}", ref,
		"{{VERSION}}", version,
	).Replace(string(raw)), nil
}

var refNoRe = regexp.MustCompile(`^([A-Za-z]+)-([A-Za-z0-9]+)-(\d+)$`)

// SuggestRefNo proposes the next free RefNo for kind and area ("QM",
// "KAL" …) after the ones in existing: FB-QM-002 → FB-QM-003, with the
// same number of digits (at least three). Only a proposal: whether a RefNo
// is free among the approved documents only the Sensor Monitor knows.
func SuggestRefNo(kind, area string, existing []string) string {
	t, ok := templateFor(kind)
	area = strings.ToUpper(strings.TrimSpace(area))
	if !ok || area == "" {
		return ""
	}
	maxNo, width := 0, 3
	for _, e := range existing {
		m := refNoRe.FindStringSubmatch(strings.TrimSpace(e))
		if m == nil || !strings.EqualFold(m[1], t.Prefix) || !strings.EqualFold(m[2], area) {
			continue
		}
		if n, err := strconv.Atoi(m[3]); err == nil && n > maxNo {
			maxNo, width = n, max(width, len(m[3]))
		}
	}
	return fmt.Sprintf("%s-%s-%0*d", t.Prefix, area, width, maxNo+1)
}

// RefNosInDir collects the RefNos of the Markdown files below dir (at most
// maxFiles files, hidden directories skipped), for SuggestRefNo.
func RefNosInDir(dir string, maxFiles int) []string {
	var refs []string
	n := 0
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != dir && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}
		if n++; n > maxFiles {
			return filepath.SkipAll
		}
		b, err := os.ReadFile(path)
		if err == nil {
			if ref := ParseHeader(string(b)).RefNo; ref != "" {
				refs = append(refs, ref)
			}
		}
		return nil
	})
	return refs
}
