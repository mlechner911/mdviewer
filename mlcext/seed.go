package mlcext

// The built-in templates copied into the user's template folder, where
// they can be changed without a new release.

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// seedManifest lists the built-in template files already copied into a
// user folder, so that a file the user deleted is not copied again.
const seedManifest = ".mlc-vorlagen.json"

// SeedTemplates copies the built-in templates into dir, the user's own
// template folder, where they can be changed without a new release: files
// copied before (also deleted ones) and files that exist are left alone, so
// an update never overwrites the user's changes; a template that is new in
// the program is added. It returns the files it wrote. Load the folder with
// LoadTemplatesFrom(false, dir, …) afterwards.
func SeedTemplates(dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	seeded := map[string]bool{}
	manifest := filepath.Join(dir, seedManifest)
	if b, err := os.ReadFile(manifest); err == nil {
		var names []string
		if json.Unmarshal(b, &names) == nil {
			for _, n := range names {
				seeded[n] = true
			}
		}
	}
	entries, err := fs.ReadDir(builtinTemplates, "templates")
	if err != nil {
		return nil, err
	}
	var written []string
	for _, e := range entries {
		name := e.Name()
		if seeded[name] {
			continue
		}
		seeded[name] = true
		target := filepath.Join(dir, name)
		if _, err := os.Stat(target); err == nil {
			continue // the user's file wins
		}
		b, err := fs.ReadFile(builtinTemplates, "templates/"+name)
		if err != nil {
			return written, err
		}
		if err := os.WriteFile(target, b, 0o644); err != nil {
			return written, err
		}
		written = append(written, name)
	}
	if _, err := os.Stat(filepath.Join(dir, "LIESMICH.md")); err != nil && !seeded["LIESMICH.md"] {
		if err := os.WriteFile(filepath.Join(dir, "LIESMICH.md"), []byte(templatesReadme), 0o644); err == nil {
			written = append(written, "LIESMICH.md")
		}
	}
	seeded["LIESMICH.md"] = true
	names := make([]string, 0, len(seeded))
	for n := range seeded {
		names = append(names, n)
	}
	sort.Strings(names)
	b, _ := json.MarshalIndent(names, "", "  ")
	if err := os.WriteFile(manifest, append(b, '\n'), 0o644); err != nil {
		return written, err
	}
	return written, nil
}

const templatesReadme = `# Vorlagen

Jede Vorlage besteht aus zwei Dateien: <id>.json (Name, Eingabefelder,
Dateiname) und dem Markdown-Text, auf den "body" verweist. Ändern, löschen
und neue anlegen geht ohne neues Programm; "Neu aus Vorlage" liest den
Ordner bei jedem Öffnen.

- Felder mit "target": "frontmatter" landen im Front Matter (doc_id,
  version, title, description, event_kinds bei Dokumenten der
  MLC-Erweiterung).
- Platzhalter im Text: {{id}}, {{id|yaml}}, {{id|list}}, {{id|ref}}.
- {{doc_id}}, {{version}} und {{title}} bleiben im Dokument stehen und
  werden erst beim Anzeigen und im PDF ersetzt.
- "category": "mlc" oder "general" ordnet die Vorlage im Dialog ein.

Ein Programmupdate überschreibt hier nichts; neue eingebaute Vorlagen kommen
dazu. Eine gelöschte eingebaute Vorlage bleibt gelöscht. Vorlagen mit
gleicher id in _vorlagen/ im Dokumentenordner ersetzen die von hier.

Beschreibung des Formats: mlc-template.schema.json im MarkSafe-Repository
(github.com/mlechner911/mdviewer, Ordner mlcext).
`
