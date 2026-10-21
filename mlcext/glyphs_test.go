package mlcext

import "testing"

func TestPDFSymbolsPrint(t *testing.T) {
	for _, g := range PDFSymbols {
		for _, s := range g.Symbols {
			for _, r := range s {
				if !HasPDFGlyph(r) {
					t.Errorf("%s: %q (U+%04X) does not print in the PDF", g.Name, s, r)
				}
			}
		}
	}
	for _, r := range "🧠🔥📌🔒" {
		if HasPDFGlyph(r) {
			t.Errorf("%q must count as missing", r)
		}
	}
	for _, r := range "äöüß€✔☐⚠️✅" {
		if !HasPDFGlyph(r) {
			t.Errorf("%q must print", r)
		}
	}
}

func TestCheckGlyphs(t *testing.T) {
	md := "---\ndoc_id: SOP-QM-001\n---\n# 🧠 Analyse\n\n✅ ok ⚠️ Achtung 🧠 nochmal 📌\n"
	var got []string
	for _, d := range Check(md, Options{}) {
		if d.Code == "pdf-glyph-missing" {
			got = append(got, d.Message[:len("„🧠“")])
			if d.Line != 4 && d.Line != 6 {
				t.Errorf("line %d", d.Line)
			}
		}
	}
	if len(got) != 2 || got[0] != "„🧠“" || got[1] != "„📌“" {
		t.Errorf("missing glyphs: %q", got)
	}
}
