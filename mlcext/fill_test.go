package mlcext

import (
	"os"
	"strings"
	"testing"
)

func TestFillTextKeepsWidth(t *testing.T) {
	md := "Ort: [@#ort ______________________________] <!--* Hinweis -->"
	got, set := FillFields(md, map[string]string{"ort": "Screenshots › Kühlschrank"})
	want := "Ort: [@#ort __Screenshots › Kühlschrank___] <!--* Hinweis -->" // width 30 stays
	if got != want || len(set) != 1 || set[0] != "ort" {
		t.Fatalf("got %q (%v), want %q", got, set, want)
	}
}

func TestFillPatternAndExistingPreset(t *testing.T) {
	md := "[@#d TT.MM.JJJJ] [@#u HH:MM] [@#x TT.MM.JJJJ=01\\.01\\.2026]"
	got, _ := FillFields(md, map[string]string{"d": "16.10.2026", "u": "22:47", "x": "17.10.2026"})
	want := "[@#d TT.MM.JJJJ=16\\.10\\.2026] [@#u HH:MM=22\\:47] [@#x TT.MM.JJJJ=17\\.10\\.2026]"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	got, _ = FillFields(got, map[string]string{"x": ""})
	if !strings.HasSuffix(got, "[@#x TT.MM.JJJJ]") {
		t.Fatalf("clear pattern: %q", got)
	}
}

func TestFillMultiline(t *testing.T) {
	md := "[@#vorfall __________\n__________\n__________]"
	got, _ := FillFields(md, map[string]string{"vorfall": "Zeile 1\nZeile 2"})
	want := "[@#vorfall __Zeile 1__\n__Zeile 2__\n__________]" // at least two underscores after
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	// More lines than the field has: the field grows, nothing is lost.
	got, _ = FillFields("[@#v ____\n____]", map[string]string{"v": "a\nb\nc"})
	if strings.Count(got, "\n") != 2 || !strings.Contains(got, "__c__") {
		t.Fatalf("grow: %q", got)
	}
}

func TestFillSingleLineJoinsLines(t *testing.T) {
	got, _ := FillFields("[@#ort ________]", map[string]string{"ort": "A\nB"})
	if got != "[@#ort __A B___]" {
		t.Fatalf("got %q", got)
	}
}

// A value from a record must neither end a field nor start a new one.
func TestFillMasksInjection(t *testing.T) {
	got, set := FillFields("[@#ort ____] [@#risiko_kritisch]", map[string]string{"ort": "x] [@#risiko_kritisch"})
	if strings.Count(got, "[@#") != 2 || !strings.Contains(got, `x\] \[\@\#risiko\_kritisch`) || len(set) != 1 {
		t.Fatalf("got %q (%v)", got, set)
	}
}

func TestFillCheckBoxes(t *testing.T) {
	md := "* [@#fest] Feststoffe\n* [@#fl x] Flüssig\n* [@] Pflaster\n* [@X] Salbe"
	got, set := FillFields(md, map[string]string{"fest": "x", "fl": "", "feld_1": "ja", "feld_2": "nein"})
	want := "* [@#fest x] Feststoffe\n* [@#fl] Flüssig\n* [@x] Pflaster\n* [@] Salbe"
	if got != want || len(set) != 4 {
		t.Fatalf("got %q (%v), want %q", got, set, want)
	}
	v := FieldValues(got)
	if v["fest"] != "x" || v["fl"] != "" || v["feld_1"] != "x" || v["feld_2"] != "" {
		t.Fatalf("values %v", v)
	}
}

func TestFillUnnamedAndUnknownAndCode(t *testing.T) {
	md := "## Angaben\n[@#unbekannt ____] [@ ________]\n```\n[@#ort ____]\n```\n[@#ort ____]"
	got, set := FillFields(md, map[string]string{"angaben_1": "Lager", "ort": "Keller", "nope": "z"})
	if !strings.Contains(got, "[@#unbekannt ____]") || !strings.Contains(got, "[@ __Lager__]") ||
		!strings.Contains(got, "```\n[@#ort ____]\n```") || !strings.HasSuffix(got, "[@#ort __Keller__]") {
		t.Fatalf("got %q", got)
	}
	if strings.Join(set, ",") != "angaben_1,ort" {
		t.Fatalf("set: %v", set)
	}
}

func TestFieldValuesRoundTrip(t *testing.T) {
	md := "[@#a ________] [@#d TT.MM.JJJJ] [@#m ______\n______] [@#c]"
	in := map[string]string{"a": "Ibu 400", "d": "28.10.2026", "m": "eins\nzwei", "c": "x"}
	filled, _ := FillFields(md, in)
	out := FieldValues(filled)
	for k, v := range in {
		if out[k] != v {
			t.Fatalf("%s: got %q, want %q (md %q)", k, out[k], v, filled)
		}
	}
	again, _ := FillFields(filled, out)
	if again != filled {
		t.Fatalf("refill changed md:\n%q\n%q", filled, again)
	}
}

func TestMaskFieldValueGlyphs(t *testing.T) {
	if got := MaskFieldValue("a ⇒ b, 5 % (k=2)"); got != `a \=\> b\, 5 \% \(k\=2\)` {
		t.Fatalf("got %q", got)
	}
}

func TestFormRecordFrontMatter(t *testing.T) {
	src, err := os.ReadFile("testdata/form_btm.md")
	if err != nil {
		t.Fatal(err)
	}
	md := strings.Replace(string(src), "title: Protokoll", "meta:\n  schema: mlc-form-record/v1\n  form: { doc_id: FB-LAG-001, version: \"2.0\" }\n  filled_by: Dr. Anna Berger\n  completed_at: 2026-10-28T09:12:00+01:00\ntitle: Protokoll", 1)
	md, set := FillFields(md, map[string]string{"datum": "28.10.2026", "verfahren_pflaster": "x", "charge_2": "A-123"})
	if len(set) != 3 || !IsFormRecord(md) || IsFormRecord(string(src)) {
		t.Fatalf("set %v record %v", set, IsFormRecord(md))
	}
	for _, d := range Check(md, Options{}) {
		if d.Severity != Info {
			t.Errorf("filled form: %s %s", d.Code, d.Message)
		}
	}
	bad := strings.Replace(md, "mlc-form-record/v1", "x/v9", 1)
	if hasCode(Check(bad, Options{}), Warning, "frontmatter-meta-schema") == nil {
		t.Error("unknown meta.schema not reported")
	}
}
