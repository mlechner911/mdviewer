package mlcext

import (
	"strings"
	"testing"
)

// Ported from mlcmd2pdf src/form.rs: the same documents must give the same
// fields and the same errors.

func fields(t *testing.T, md string) []Field {
	t.Helper()
	fs, errs := ExtractFields(md)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	return fs
}

func TestRecognisesEachKindOfField(t *testing.T) {
	found := fields(t, "# Bericht\n\nGemeldet: [@] Datum: [@TT.MM.JJJJ] Ort: [@____Halle 3____]\n\n"+
		"## Beschreibung\n\n[@______________\n______________\n______________]\n")
	if len(found) != 4 {
		t.Fatalf("got %d fields", len(found))
	}
	if found[0].Kind != CheckBox || found[0].Name != "bericht_1" {
		t.Errorf("field 0: %+v", found[0])
	}
	if found[1].Pattern != "TT.MM.JJJJ" || found[1].Tooltip != "TT.MM.JJJJ" || found[1].Chars != 10 {
		t.Errorf("field 1: %+v", found[1])
	}
	if found[2].Value != "Halle 3" || found[2].Chars != 15 {
		t.Errorf("field 2: %+v", found[2])
	}
	if found[3].Kind != MultilineField || found[3].Rows != 3 || found[3].Name != "beschreibung_1" || found[3].Line != 7 {
		t.Errorf("field 3: %+v", found[3])
	}
}

func TestNamesAndTooltips(t *testing.T) {
	found := fields(t, "[@#ursache _____] <!--* Was war der Grund? -->\n[@#ok] [@#zeit HH:MM]\n\n"+
		"[@#vorfall\n____\n____]\n")
	if found[0].Name != "ursache" || found[0].Kind != TextField || found[0].Chars != 5 || found[0].Tooltip != "Was war der Grund?" {
		t.Errorf("field 0: %+v", found[0])
	}
	if found[1].Name != "ok" || found[1].Kind != CheckBox {
		t.Errorf("field 1: %+v", found[1])
	}
	if found[2].Name != "zeit" || found[2].Pattern != "HH:MM" {
		t.Errorf("field 2: %+v", found[2])
	}
	if found[3].Name != "vorfall" || found[3].Rows != 2 {
		t.Errorf("field 3: %+v", found[3])
	}
}

func TestLeavesCodeCommentsEscapesAndLinksAlone(t *testing.T) {
	src := "`[@]` and \\[@] and [@team](https://example.org)\n\n```\n[@]\n```\n\n<!-- [@] -->\n"
	if fs := fields(t, src); len(fs) != 0 {
		t.Errorf("expected no fields, got %+v", fs)
	}
}

func TestMultiLinePresetsKeepTheirLinesAndEscapes(t *testing.T) {
	found := fields(t, "[@#vorfall __Überschreitung · Kritisch, Kühlraum A › Temperatur__\n"+
		"____\n"+
		"__Kommentar: \\]Klammer\\], Rohr \\| Pfeife, \\_Unterstrich\\_, \\#7, \\\\pfad, <!-- kein Kommentar -->__\n"+
		"____\n"+
		"____]\n")
	if len(found) != 1 || found[0].Rows != 5 {
		t.Fatalf("got %+v", found)
	}
	want := "Überschreitung · Kritisch, Kühlraum A › Temperatur\n\n" +
		"Kommentar: ]Klammer], Rohr | Pfeife, _Unterstrich_, #7, \\pfad, <!-- kein Kommentar -->"
	if found[0].Value != want {
		t.Errorf("value = %q\nwant    %q", found[0].Value, want)
	}
	if single := fields(t, "| [@#ev ____EV\\|0820\\]____] | x |\n"); single[0].Value != "EV|0820]" {
		t.Errorf("single = %q", single[0].Value)
	}
	text := fields(t, "[@\\#kein-name ____]")
	if text[0].Value != "#kein-name" || text[0].Name != "feld_1" {
		t.Errorf("escaped #: %+v", text[0])
	}
}

func TestPatternsTakeAPresetAfterTheEqualsSign(t *testing.T) {
	found := fields(t, "[@#datum TT.MM.JJJJ=16\\.10\\.2026] [@#zeit HH:MM=__08:00__] [@#leer TT.MM.JJJJ=]\n"+
		"[@____a=b____]\n")
	if found[0].Pattern != "TT.MM.JJJJ" || found[0].Value != "16.10.2026" || found[0].Chars != 10 {
		t.Errorf("field 0: %+v", found[0])
	}
	if found[1].Value != "08:00" {
		t.Errorf("field 1: %+v", found[1])
	}
	if found[2].Value != "" || found[2].Pattern != "TT.MM.JJJJ" {
		t.Errorf("field 2: %+v", found[2])
	}
	if found[3].Value != "a=b" || found[3].Pattern != "" {
		t.Errorf("field 3: %+v", found[3])
	}
	_, errs := ExtractFields("[@HH:MM=08:00 Uhr]")
	if len(errs) != 1 || errs[0].Code != "field-preset-too-long" {
		t.Errorf("too long: %v", errs)
	}
}

func TestALastLineOfDotsFillsThePage(t *testing.T) {
	found := fields(t, "[@#rest __Vorgabe__\n______\n…]\n\n[@#kurz\n...]\n\n[@#fest ____\n____]\n")
	if !found[0].FillPage || found[0].Rows != 2 || found[0].Value != "Vorgabe" {
		t.Errorf("field 0: %+v", found[0])
	}
	if !found[1].FillPage || found[1].Rows != 1 {
		t.Errorf("field 1: %+v", found[1])
	}
	if found[2].FillPage {
		t.Errorf("field 2: %+v", found[2])
	}
}

func TestTaskListBoxesStayUntouched(t *testing.T) {
	if fs := fields(t, "- [ ] offen\n- [x] erledigt\n"); len(fs) != 0 {
		t.Errorf("task list became fields: %+v", fs)
	}
}

func TestRejectsMalformedFieldsWithTheirLine(t *testing.T) {
	for src, code := range map[string]string{
		"ok\n[@____":                         "field-unclosed",
		"[@___\n\n___]":                      "field-blank-line",
		"[@Ja|Nein]":                         "field-choice",
		"[@#a.b ___]":                        "field-name",
		"[@# ___]":                           "field-name",
		"[@#x] [@#x]":                        "field-duplicate",
		"## Titel [@]":                       "field-in-heading",
		"[@" + strings.Repeat("x", 41) + "]": "field-not-a-field",
	} {
		_, errs := ExtractFields(src)
		if len(errs) == 0 || errs[0].Code != code {
			t.Errorf("%q: got %v, want %s", src, errs, code)
		}
	}
	_, errs := ExtractFields("a\nb\n[@___")
	if len(errs) != 1 || errs[0].Line != 3 || errs[0].Col != 1 {
		t.Errorf("unclosed position: %v", errs)
	}
}

func TestReportsAllErrorsNotOnlyTheFirst(t *testing.T) {
	_, errs := ExtractFields("[@#x] [@#x]\n\n[@Ja|Nein]\n\n[@___")
	if len(errs) != 3 {
		t.Errorf("want 3 errors, got %v", errs)
	}
}

// Cases from mlcmd2pdf src/markdown.rs (slugs_headings_like_github).
func TestHeadingSlug(t *testing.T) {
	for in, want := range map[string]string{
		"EV-2026-0666 · Kritisch": "ev-2026-0666--kritisch",
		"Übersicht der Messwerte": "übersicht-der-messwerte",
		"1. Ziel & Zweck (v2)":    "1-ziel--zweck-v2",
		"snake_case ✅":            "snake_case-",
	} {
		if got := HeadingSlug(in); got != want {
			t.Errorf("HeadingSlug(%q) = %q, want %q", in, got, want)
		}
	}
}
