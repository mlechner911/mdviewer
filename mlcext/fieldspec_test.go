package mlcext

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

// fieldSpecCases: spec → written form, and the spec FieldAt reads back
// (defaults filled in). testdata/fieldspec.json holds the same for a
// TypeScript port; regenerate with: go test -run TestFieldSpecVectors -update
var fieldSpecCases = []struct {
	spec FieldSpec
	want string
	back FieldSpec // zero: same as spec
}{
	{FieldSpec{Kind: "text", Name: "ort", Chars: 12}, "[@#ort ____________]", FieldSpec{}},
	{FieldSpec{Kind: "text"}, "[@____________________]", FieldSpec{Kind: "text", Chars: 20}},
	{FieldSpec{Kind: "text", Name: "name", Chars: 20, Value: "Dr. Anna Berger"}, `[@#name __Dr\. Anna Berger___]`, FieldSpec{}},
	{FieldSpec{Kind: "text", Name: "ort", Chars: 10, Tooltip: "Raum oder  Station"}, "[@#ort __________] <!--* Raum oder Station -->", FieldSpec{Kind: "text", Name: "ort", Chars: 10, Tooltip: "Raum oder Station"}},
	{FieldSpec{Kind: "check", Name: "ok"}, "[@#ok]", FieldSpec{}},
	{FieldSpec{Kind: "check", Name: "ok", Checked: true}, "[@#ok x]", FieldSpec{}},
	{FieldSpec{Kind: "check", Checked: true}, "[@x]", FieldSpec{}},
	{FieldSpec{Kind: "pattern", Name: "datum", Pattern: "TT.MM.JJJJ"}, "[@#datum TT.MM.JJJJ]", FieldSpec{}},
	{FieldSpec{Kind: "pattern", Name: "datum", Pattern: "TT.MM.JJJJ", Value: "16.10.2026"}, `[@#datum TT.MM.JJJJ=16\.10\.2026]`, FieldSpec{}},
	{FieldSpec{Kind: "pattern", Name: "zeit", Pattern: "HH:MM"}, "[@#zeit HH:MM]", FieldSpec{}},
	{FieldSpec{Kind: "signature", Name: "zeuge", Role: "Zeuge 1"}, "[@#zeuge Unterschrift: Zeuge 1]", FieldSpec{}},
	{FieldSpec{Kind: "signature", Role: "Vernichtende Person"}, "[@Unterschrift: Vernichtende Person]", FieldSpec{}},
	{FieldSpec{Kind: "multiline", Name: "befund", Chars: 8, Rows: 2}, "[@#befund ________\n________]", FieldSpec{}},
	{FieldSpec{Kind: "multiline", Name: "befund", Chars: 10, Rows: 2, Value: "kein Befund"}, "[@#befund __kein Befund__\n__________]", FieldSpec{Kind: "multiline", Name: "befund", Chars: 15, Rows: 2, Value: "kein Befund"}},
	{FieldSpec{Kind: "multiline", Name: "notiz", Chars: 6, Rows: 1, FillPage: true}, "[@#notiz ______\n…]", FieldSpec{}},
}

func TestFormatFieldAndFieldAt(t *testing.T) {
	for _, c := range fieldSpecCases {
		got, err := FormatField(c.spec)
		if err != nil || got != c.want {
			t.Errorf("FormatField(%+v) = %q, %v; want %q", c.spec, got, err, c.want)
			continue
		}
		// Read it back inside a document, cursor on the field's last line.
		md := "# Prüfung\n\nFeld: " + got + "\n"
		fieldLines := strings.Count(got, "\n")
		e := FieldAt(md, 3+fieldLines, 2)
		if fieldLines == 0 {
			e = FieldAt(md, 3, 8)
		}
		want := c.back
		if reflect.DeepEqual(want, FieldSpec{}) {
			want = c.spec
		}
		if !e.OK || !reflect.DeepEqual(e.Spec, want) {
			t.Errorf("FieldAt(%q) = %+v; want %+v", got, e, want)
			continue
		}
		if e.Line != 3 || e.Col != 7 {
			t.Errorf("FieldAt(%q) starts at %d:%d, want 3:7", got, e.Line, e.Col)
		}
		// The range covers the field and its tooltip exactly.
		if end := offsetOf(md, e.EndLine, e.EndCol); md[offsetOf(md, e.Line, e.Col):end] != got {
			t.Errorf("FieldAt(%q) range = %q", got, md[offsetOf(md, e.Line, e.Col):end])
		}
		if d := Check(md, Options{}); len(d) > 0 {
			for _, x := range d {
				if x.Severity == Error {
					t.Errorf("Check(%q): %s %s", got, x.Code, x.Message)
				}
			}
		}
	}
}

func TestFormatFieldErrors(t *testing.T) {
	for _, s := range []FieldSpec{
		{Kind: "text", Name: "größe"},
		{Kind: "text", Name: "a.b"},
		{Kind: "pattern", Name: "d"},
		{Kind: "pattern", Name: "d", Pattern: "__.__"},
		{Kind: "pattern", Name: "d", Pattern: "HH:MM", Value: "12:30:00"},
		{Kind: "text", Name: "t", Tooltip: "a --> b"},
		{Kind: "auswahl", Name: "x"},
	} {
		if got, err := FormatField(s); err == nil {
			t.Errorf("FormatField(%+v) = %q, want an error", s, got)
		}
	}
}

func TestFieldAtOutside(t *testing.T) {
	md := "Text [@#a ____] Text\n"
	if e := FieldAt(md, 1, 2); e.OK {
		t.Errorf("FieldAt before the field = %+v", e)
	}
	if e := FieldAt(md, 1, 6); !e.OK || e.Spec.Name != "a" {
		t.Errorf("FieldAt on the field = %+v", e)
	}
}

type fieldSpecVector struct {
	Spec FieldSpec `json:"spec"`
	Text string    `json:"text"`
}

func TestFieldSpecVectors(t *testing.T) {
	var got []fieldSpecVector
	for _, c := range fieldSpecCases {
		text, _ := FormatField(c.spec)
		got = append(got, fieldSpecVector{Spec: c.spec, Text: text})
	}
	b, _ := json.MarshalIndent(got, "", "  ")
	b = append(b, '\n')
	if *update {
		if err := os.WriteFile("testdata/fieldspec.json", b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old, err := os.ReadFile("testdata/fieldspec.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(old) != string(b) {
		t.Errorf("testdata/fieldspec.json is out of date; regenerate with -update and check the diff")
	}
}
