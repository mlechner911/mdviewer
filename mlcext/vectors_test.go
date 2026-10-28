package mlcext

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"
)

// testdata/fields.json holds field-scanner cases with the expected result
// of ExtractFields. The TypeScript renderer (packages/mlc-markdown) runs the
// same file, so both recognise fields identically. Regenerate after a
// deliberate change: go test -run TestFieldVectors -update
var update = flag.Bool("update", false, "rewrite testdata/fields.json from ExtractFields")

type vectorField struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Named    bool   `json:"named"`
	Value    string `json:"value"`
	Checked  bool   `json:"checked"`
	Role     string `json:"role"`
	Chars    int    `json:"chars"`
	Rows     int    `json:"rows"`
	FillPage bool   `json:"fillPage"`
	Pattern  string `json:"pattern"`
	Tooltip  string `json:"tooltip"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
}

type vectorError struct {
	Code string `json:"code"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
}

type vector struct {
	Markdown string        `json:"markdown"`
	Fields   []vectorField `json:"fields"`
	Errors   []vectorError `json:"errors"`
}

var vectorInputs = []string{
	"# Bericht\n\nGemeldet: [@] Datum: [@TT.MM.JJJJ] Ort: [@____Halle 3____]\n\n## Beschreibung\n\n[@______________\n______________\n______________]\n",
	"[@#ursache _____] <!--* Was war der Grund? -->\n[@#ok] [@#zeit HH:MM]\n\n[@#vorfall\n____\n____]\n",
	"`[@]` and \\[@] and [@team](https://example.org)\n\n```\n[@]\n```\n\n<!-- [@] -->\n",
	"[@#vorfall __Überschreitung · Kritisch, Kühlraum A › Temperatur__\n____\n__Kommentar: \\]Klammer\\], Rohr \\| Pfeife, \\_Unterstrich\\_, \\#7, \\\\pfad, <!-- kein Kommentar -->__\n____\n____]\n",
	"| [@#ev ____EV\\|0820\\]____] | x |\n",
	"[@\\#kein-name ____]",
	"[@#datum TT.MM.JJJJ=16\\.10\\.2026] [@#zeit HH:MM=__08:00__] [@#leer TT.MM.JJJJ=]\n[@____a=b____]\n",
	"[@#rest __Vorgabe__\n______\n…]\n\n[@#kurz\n...]\n\n[@#fest ____\n____]\n",
	"- [ ] offen\n- [x] erledigt\n",
	"ok\n[@____",
	"[@___\n\n___]",
	"[@Ja|Nein]",
	"[@#a.b ___]",
	"[@#x] [@#x]",
	"## Titel [@]",
	"[@HH:MM=08:00 Uhr]",
	"- [@#sig_v Unterschrift: Vernichtende Person]\n- [@signature: Witness\\] 1] [@UNTERSCHRIFT:] [@Unterschrift]\n",
	"# Ä Öl\n\n* **Ort:** [@#ort ___] <!--* z. B. Kühlraum A -->\n* [@] Kritisch\n",
	"* [@#fest x] Feststoffe\n* [@#fl] Flüssig\n* [@X] Pflaster\n[@xx] [@#n x y]\n",
}

func vectorsFor(inputs []string) []vector {
	out := make([]vector, 0, len(inputs))
	for _, md := range inputs {
		fs, errs := ExtractFields(md)
		v := vector{Markdown: md, Fields: []vectorField{}, Errors: []vectorError{}}
		for _, f := range fs {
			v.Fields = append(v.Fields, vectorField{
				Kind: f.Kind.String(), Name: f.Name, Named: f.Named, Value: f.Value, Checked: f.Checked, Role: f.Role, Chars: f.Chars,
				Rows: f.Rows, FillPage: f.FillPage, Pattern: f.Pattern, Tooltip: f.Tooltip, Line: f.Line, Col: f.Col,
			})
		}
		for _, e := range errs {
			v.Errors = append(v.Errors, vectorError{Code: e.Code, Line: e.Line, Col: e.Col})
		}
		out = append(out, v)
	}
	return out
}

func TestFieldVectors(t *testing.T) {
	got := vectorsFor(vectorInputs)
	if *update {
		b, _ := json.MarshalIndent(got, "", "  ")
		if err := os.WriteFile("testdata/fields.json", append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	b, err := os.ReadFile("testdata/fields.json")
	if err != nil {
		t.Fatalf("%v (run: go test -run TestFieldVectors -update)", err)
	}
	var want []vector
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("testdata/fields.json is out of date with ExtractFields; regenerate with -update and check the diff")
	}
}
