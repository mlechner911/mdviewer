package mlcext

import (
	"os"
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
