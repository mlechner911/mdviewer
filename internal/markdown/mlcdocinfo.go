package markdown

import (
	"fmt"
	htmlpkg "html"
	"strings"

	"github.com/mlechner911/mdviewer/mlcext"
)

// renderMLCDocInfo shows how the Sensor Monitor reads a document of the MLC
// Markdown Extension (non-standard, docs/MLC-EXTENSIONS.md): id, kind,
// version, title, event kinds and form fields, all parsed by mlcext — the
// same code the server uses. It replaces the generic front matter table for
// such documents; it is an editing aid, hidden in print and export.
func renderMLCDocInfo(md string) string {
	h := mlcext.ParseHeader(md)
	fm, _ := mlcext.ParseFrontMatter(md)
	fields, errs := mlcext.ExtractFields(md)

	var sb strings.Builder
	row := func(label, value string) {
		sb.WriteString(`<dt>` + htmlpkg.EscapeString(label) + `</dt>`)
		if value == "" {
			sb.WriteString(`<dd class="mlc-docinfo-missing">fehlt</dd>`)
			return
		}
		sb.WriteString(`<dd>` + value + `</dd>`)
	}
	esc := htmlpkg.EscapeString

	sb.WriteString(`<div class="mlc-docinfo">`)
	sb.WriteString(`<div class="mlc-docinfo-title">MLC-Erweiterung · so liest der Sensor Monitor dieses Dokument</div><dl>`)
	row("Kennung", esc(h.RefNo))
	kind := mlcext.KindLabel(h.Kind)
	if kind == "" && h.RefNo != "" {
		kind = "unbekannt (Präfix SOP-, VA-, FB- oder INFO-)"
	}
	row("Art", esc(kind))
	version := h.Version
	if version == "" {
		version = fm.Version
	}
	row("Version", esc(version))
	if fm.Title != "" {
		row("Titel", esc(fm.Title))
	}
	if fm.Description != "" {
		row("Beschreibung", esc(fm.Description))
	}
	if fm.EventKindsSet {
		kinds := "gilt allgemein"
		if len(fm.EventKinds) > 0 {
			tags := make([]string, len(fm.EventKinds))
			for i, k := range fm.EventKinds {
				tags[i] = `<span class="frontmatter-tag">` + esc(k) + `</span>`
			}
			kinds = strings.Join(tags, " ")
		}
		row("Gilt bei", kinds)
	}
	if len(fields) > 0 || len(errs) > 0 {
		prefill := 0
		for _, f := range fields {
			if f.Named && mlcext.IsPrefillField(f.Name) {
				prefill++
			}
		}
		v := fmt.Sprintf("%d", len(fields))
		if prefill > 0 {
			v += fmt.Sprintf(", davon %d aus einem Ereignis vorausgefüllt", prefill)
		}
		if len(errs) > 0 {
			v += fmt.Sprintf(`, <span class="mlc-docinfo-error">%d fehlerhaft</span>`, len(errs))
		}
		row("Formularfelder", v)
	}
	sb.WriteString(`</dl></div>`)
	return sb.String()
}
