package mlcext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func builtin(t *testing.T, id string) TemplateInfo {
	t.Helper()
	for _, tpl := range LoadTemplates() {
		if tpl.ID == id {
			return tpl
		}
	}
	t.Fatalf("no built-in template %q", id)
	return TemplateInfo{}
}

func TestBuiltinTemplatesHaveNoProblems(t *testing.T) {
	list := LoadTemplates()
	if len(list) != 4 {
		t.Fatalf("want 4 built-in templates, got %d", len(list))
	}
	for _, tpl := range list {
		if len(tpl.Problems) > 0 || tpl.Source != "builtin" {
			t.Errorf("%s: %v (source %s)", tpl.ID, tpl.Problems, tpl.Source)
		}
	}
}

func TestEveryBuiltinTemplateCreatesAValidDocument(t *testing.T) {
	prefix := map[string]string{"sop": "SOP", "va": "VA", "form": "FB", "info": "INFO"}
	for id, p := range prefix {
		tpl := builtin(t, id)
		res := tpl.Render(map[string]any{
			"titel": "Lagerung: Kühlkette & Co.", "bereich": "QM", "kennung": strings.ToLower(p) + "-qm-001",
			"version": "1.0", "anlass": []any{"OUT_OF_RANGE", "NO_DATA"},
		})
		if len(res.FieldErrors) > 0 {
			t.Fatalf("%s: %v", id, res.FieldErrors)
		}
		if strings.Contains(res.Content, "{{") {
			t.Errorf("%s: placeholder left:\n%s", id, res.Content)
		}
		for _, d := range res.Diagnostics {
			if d.Severity != Info {
				t.Errorf("%s: %d:%d %s %s", id, d.Line, d.Col, d.Code, d.Message)
			}
		}
		h := ParseHeader(res.Content)
		if h.RefNo != p+"-QM-001" || h.Version != "1.0" || KindFromRef(h.RefNo) != id {
			t.Errorf("%s: header %+v", id, h)
		}
		fm, _ := ParseFrontMatter(res.Content)
		if fm.Title != "Lagerung: Kühlkette & Co." {
			t.Errorf("%s: title %q", id, fm.Title)
		}
		if id != "info" && strings.Join(fm.EventKinds, ",") != "OUT_OF_RANGE,NO_DATA" {
			t.Errorf("%s: event_kinds %v", id, fm.EventKinds)
		}
		if res.FileName != p+"-QM-001.md" {
			t.Errorf("%s: file name %q", id, res.FileName)
		}
	}
}

func TestEmptyMultiselectMeansApplyAlways(t *testing.T) {
	res := builtin(t, "sop").Render(map[string]any{"titel": "X", "kennung": "SOP-QM-001", "version": "1.0"})
	fm, _ := ParseFrontMatter(res.Content)
	if !fm.EventKindsSet || len(fm.EventKinds) != 0 {
		t.Errorf("want event_kinds: [] , got %+v\n%s", fm, res.Content)
	}
}

func writeTemplate(t *testing.T, dir, id, schema, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, id+".json"), []byte(schema), 0o600); err != nil {
		t.Fatal(err)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(dir, id+".md"), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFieldTypes(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "messung", `{
	  "name": "Messprotokoll", "body": "messung.md", "file": "Messung {{datum}}",
	  "fields": [
	    {"id": "datum", "label": "Datum", "type": "date", "required": true},
	    {"id": "iso", "label": "ISO", "type": "date", "format": "JJJJ-MM-TT"},
	    {"id": "uhrzeit", "label": "Uhrzeit", "type": "hour"},
	    {"id": "temp", "label": "Temperatur", "type": "number", "min": -80, "max": 30, "step": 0.5},
	    {"id": "ort", "label": "Ort", "type": "select", "options": ["Labor 1", "Kühlraum A"]},
	    {"id": "notiz", "label": "Notiz", "type": "multiline"}
	  ]}`,
		"# Messung {{datum}}\n\nISO {{iso}} um {{uhrzeit}}: {{temp}} °C in {{ort}}.\n\n{{notiz}}\n")
	tpl := LoadTemplates(dir)
	var m TemplateInfo
	for _, x := range tpl {
		if x.ID == "messung" {
			m = x
		}
	}
	if len(m.Problems) > 0 || m.Source != dir {
		t.Fatalf("template: %+v", m)
	}
	if m.Fields[2].Type != "time" {
		t.Errorf("hour must become time, got %s", m.Fields[2].Type)
	}
	ok := m.Render(map[string]any{"datum": "5.3.2026", "iso": "05.03.2026", "uhrzeit": "8:05", "temp": "-70.5", "ort": "Labor 1", "notiz": "Zeile 1\nZeile 2"})
	if len(ok.FieldErrors) > 0 {
		t.Fatalf("errors: %v", ok.FieldErrors)
	}
	want := "# Messung 05.03.2026\n\nISO 2026-03-05 um 08:05: -70,5 °C in Labor 1.\n\nZeile 1\nZeile 2\n"
	if ok.Content != want {
		t.Errorf("content:\n%q\nwant\n%q", ok.Content, want)
	}
	if ok.FileName != "Messung 05.03.2026.md" {
		t.Errorf("file name %q", ok.FileName)
	}
	bad := m.Render(map[string]any{"datum": "31.02.2026", "uhrzeit": "25:00", "temp": "-70.3", "ort": "Keller"})
	for _, id := range []string{"datum", "uhrzeit", "temp", "ort"} {
		if bad.FieldErrors[id] == "" {
			t.Errorf("%s: want a field error, got %v", id, bad.FieldErrors)
		}
	}
	if m.Render(map[string]any{}).FieldErrors["datum"] != "Pflichtfeld" {
		t.Errorf("required field not reported")
	}
}

func TestRefnoAndSharedLists(t *testing.T) {
	tpl := builtin(t, "form")
	res := tpl.Render(map[string]any{"titel": "X", "kennung": "SOP-QM-001", "version": "1"})
	if res.FieldErrors["kennung"] == "" {
		t.Errorf("refno with the wrong prefix must be an error")
	}
	res = tpl.Render(map[string]any{"titel": "X", "kennung": "FB-QM-001", "version": "1", "anlass": "DEVIATION, FOO"})
	if res.FieldErrors["anlass"] == "" {
		t.Errorf("unknown event kind must be an error")
	}
	for _, f := range tpl.Fields {
		if f.ID == "anlass" && strings.Join(f.OptionList, ",") != strings.Join(EventKinds, ",") {
			t.Errorf("event_kinds not resolved: %v", f.OptionList)
		}
	}
}

func TestBrokenTemplatesAreReported(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "kaputt", `{"name": "Kaputt", "body": "fehlt.md", "fields": [
	  {"id": "Titel", "label": "", "type": "farbe"},
	  {"id": "wahl", "label": "Wahl", "type": "select"},
	  {"id": "k", "label": "K", "type": "refno", "area": "nirgends"}
	]}`, "")
	writeTemplate(t, dir, "unbekannt", `{"name": "U", "body": "u.md", "fields": [], "farbe": "rot"}`, "")
	writeTemplate(t, dir, "platzhalter", `{"name": "P", "body": "platzhalter.md", "fields": []}`, "# {{wer}} {{x|upper}}\n")
	problems := map[string]string{}
	for _, tpl := range LoadTemplates(dir) {
		problems[tpl.ID] = strings.Join(tpl.Problems, " | ")
	}
	for id, wants := range map[string][]string{
		"kaputt":      {"nicht gefunden", "Kleinbuchstaben", "label", "unbekannter Typ", "options", "prefix", "area"},
		"unbekannt":   {"kein gültiges Vorlagen-Schema"},
		"platzhalter": {"{{wer}} hat kein Feld", "Filter"},
	} {
		for _, w := range wants {
			if !strings.Contains(problems[id], w) {
				t.Errorf("%s: missing %q in %q", id, w, problems[id])
			}
		}
	}
	if res := (TemplateInfo{Problems: []string{"x"}}).Render(nil); res.FieldErrors[""] == "" {
		t.Errorf("a broken template must not render")
	}
}

func TestUserTemplateReplacesBuiltin(t *testing.T) {
	dir := t.TempDir()
	writeTemplate(t, dir, "sop", `{"name": "Unsere SOP", "body": "sop.md", "fields": [
	  {"id": "kennung", "label": "Kennung", "type": "refno", "prefix": "SOP", "required": true}]}`,
		"# SOP\n\n**Dokumenten-ID:** {{kennung}} | **Version:** 1.0\n")
	n := 0
	for _, tpl := range LoadTemplates(dir) {
		if tpl.ID == "sop" {
			n++
			if tpl.Name != "Unsere SOP" || tpl.Source != dir {
				t.Errorf("not replaced: %+v", tpl)
			}
		}
	}
	if n != 1 {
		t.Errorf("want one sop template, got %d", n)
	}
}

func TestSuggestRefNo(t *testing.T) {
	existing := []string{"FB-QM-002", "FB-QM-010", "FB-TE-001", "SOP-QM-006", "fb-qm-0042x"}
	for _, c := range []struct{ prefix, area, want string }{
		{"FB", "QM", "FB-QM-011"},
		{"fb", "te", "FB-TE-002"},
		{"FB", "LAB", "FB-LAB-001"},
		{"SOP", "QM", "SOP-QM-007"},
		{"SOP", "", ""},
	} {
		if got := SuggestRefNo(c.prefix, c.area, existing); got != c.want {
			t.Errorf("SuggestRefNo(%s, %s) = %q, want %q", c.prefix, c.area, got, c.want)
		}
	}
	if got := SuggestRefNo("SOP", "KAL", []string{"SOP-KAL-0007"}); got != "SOP-KAL-0008" {
		t.Errorf("keeps the width: %q", got)
	}
}

func TestRefNosInDir(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		p := filepath.Join(dir, name)
		_ = os.MkdirAll(filepath.Dir(p), 0o700)
		_ = os.WriteFile(p, []byte(content), 0o600)
	}
	write("a.md", "**Dokumenten-ID:** SOP-QM-001\n")
	write("sub/b.md", "**Dokumenten-ID:** FB-QM-002 | **Version:** 1\n")
	write(".hidden/c.md", "**Dokumenten-ID:** SOP-QM-099\n")
	write("d.txt", "**Dokumenten-ID:** SOP-QM-050\n")
	if got := strings.Join(RefNosInDir(dir, 100), ","); got != "SOP-QM-001,FB-QM-002" {
		t.Errorf("RefNosInDir = %s", got)
	}
}
