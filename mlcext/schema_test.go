package mlcext

import (
	"os"
	"slices"
	"testing"
)

func TestFormSchemaBtM(t *testing.T) {
	md, err := os.ReadFile("testdata/form_btm.md")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]FormField{}
	for _, f := range FormSchema(string(md)) {
		got[f.Name] = f
	}
	cases := []struct{ name, kind, label, section, group string }{
		{"institution", "text", "Apotheke / Institution", "1. Allgemeine Angaben", ""},
		{"uhrzeit", "text", "Uhrzeit", "1. Allgemeine Angaben", ""},
		{"charge_1", "text", "Charge (1)", "2. Auflistung der zu vernichtenden Betäubungsmittel", ""},
		{"grund_4", "text", "Grund (4)", "2. Auflistung der zu vernichtenden Betäubungsmittel", ""},
		{"verfahren_feststoffe", "check", "Feststoffe", "3. Beschreibung des Vernichtungsverfahrens", ""},
		{"verfahren_sonstiges", "check", "Sonstiges Verfahren", "3. Beschreibung des Vernichtungsverfahrens", ""},
		{"verfahren_sonstiges_text", "text", "Sonstiges Verfahren", "3. Beschreibung des Vernichtungsverfahrens", ""},
		{"name_zeuge1", "text", "Name", "4. Bestätigung und Unterschriften (Vier-Augen-Prinzip)", "1. Zeuge/Zeugin (Pharmazeutisches Personal)"},
	}
	for _, c := range cases {
		f, ok := got[c.name]
		if !ok {
			t.Fatalf("%s missing", c.name)
		}
		if f.Kind != c.kind || f.Label != c.label || f.Section != c.section || f.Group != c.group {
			t.Errorf("%s: got kind %q label %q section %q group %q", c.name, f.Kind, f.Label, f.Section, f.Group)
		}
	}
	if got["datum"].Pattern != "TT.MM.JJJJ" || got["institution"].Tooltip != "Name und Anschrift" {
		t.Errorf("datum %+v / institution %+v", got["datum"], got["institution"])
	}
}

func TestFormSchemaValuesAndLineAbove(t *testing.T) {
	md := "# Befund\nWas ist passiert?\n[@#was ______\n______]\n* [@#ok x] geprüft\n"
	s := FormSchema(md)
	if len(s) != 2 || s[0].Label != "Was ist passiert?" || s[0].Kind != "multiline" || s[0].Rows != 2 {
		t.Fatalf("%+v", s)
	}
	if s[1].Value != "x" || s[1].Label != "geprüft" || s[1].Section != "Befund" {
		t.Fatalf("%+v", s[1])
	}
}

func layoutString(md string) []string {
	out := []string{}
	for _, b := range FormLayout(md) {
		if b.Field != nil {
			out = append(out, "#"+b.Field.Name+"="+b.Field.Label)
		} else {
			out = append(out, b.Text)
		}
	}
	return out
}

func TestFormLayout(t *testing.T) {
	for _, c := range []struct {
		name, md string
		want     []string
	}{
		{"text before, between and after",
			"# T\n\nVorab.\n\nName: [@#a ____]\n\nDazwischen.\n\nOrt: [@#b ____]\n\nAm Ende.",
			[]string{"# T\n\nVorab.", "#a=Name", "Dazwischen.", "#b=Ort", "Am Ende."}},
		{"front matter, variables and signature stay out",
			"---\ndoc_id: X-1\n---\nKennung {{doc_id}}\n\nA: [@#a ____]\n\n<!--mlc-signature\nsha256: x\n-->\n",
			[]string{"Kennung X-1", "#a=A"}},
		{"the line above a lone field is its label, once",
			"Frage?\n\n[@#a ____]\n\nDanach.",
			[]string{"#a=Frage?", "Danach."}},
		{"a heading above a lone field stays",
			"## Befund\n\n[@#a ____]",
			[]string{"## Befund", "#a=Befund"}},
		{"text behind a field",
			"Ort: [@#a ____] und Text dahinter",
			[]string{"#a=Ort", "und Text dahinter"}},
		{"text before a check box with its own label",
			"Kontrolle: [@#k] erledigt",
			[]string{"Kontrolle:", "#k=erledigt"}},
		{"the label of the next field stays with it",
			"Kontrolle: [@#k] Ort: [@#o ____]",
			[]string{"#k=Kontrolle", "#o=Ort"}},
		{"table with fields: its cells become labels",
			"Vor.\n\n| Charge | Menge |\n|---|---|\n| [@#c ____] | [@#m ____] |\n\nNach.",
			[]string{"Vor.", "#c=Charge (1)", "#m=Menge (1)", "Nach."}},
		{"table without fields stays text",
			"| A |\n|---|\n| 1 |\n\nN: [@#a ____]",
			[]string{"| A |\n|---|\n| 1 |", "#a=N"}},
		{"comments: tooltip out, fill hint stays",
			"<!--! So geht es. -->\n\n<!-- intern -->\n\nA: [@#a ____] <!--* Tipp -->",
			[]string{"<!--! So geht es. -->", "#a=A"}},
		{"parent list item is the group",
			"- 2. Zeuge\n  - Name: [@#a ____]\n- Weiter",
			[]string{"#a=Name", "- Weiter"}},
		{"multi-line field",
			"Text: [@#a ________\n________]\n\nEnde.",
			[]string{"#a=Text", "Ende."}},
		{"no field, only text", "Nur Text.", []string{"Nur Text."}},
		{"code stays text", "```\n[@#x ____]\n```\n\nA: [@#a ____]", []string{"```\n[@#x ____]\n```", "#a=A"}},
	} {
		if got := layoutString(c.md); !slices.Equal(got, c.want) {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

// Every field of the real form is in the layout, in document order.
func TestFormLayoutKeepsAllFields(t *testing.T) {
	raw, err := os.ReadFile("testdata/form_btm.md")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, b := range FormLayout(string(raw)) {
		if b.Field != nil {
			names = append(names, b.Field.Name)
		}
	}
	var want []string
	for _, f := range FormSchema(string(raw)) {
		want = append(want, f.Name)
	}
	if !slices.Equal(names, want) {
		t.Errorf("got %v\nwant %v", names, want)
	}
}
