package mlcext

import (
	"strings"
	"testing"
)

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

func TestHasPDFFieldGlyph(t *testing.T) {
	for _, r := range "äöüß€°µ≤→αЖ\n" {
		if !HasPDFFieldGlyph(r) {
			t.Errorf("%q must print in a field", r)
		}
	}
	// Measured with mlcdoc 0.19.1: these print in the text, not in a field.
	for _, r := range "✔✘✅❌⇒\u0615\u064B🧠" {
		if HasPDFFieldGlyph(r) {
			t.Errorf("%q must count as missing in a field", r)
		}
	}
}

func TestCheckFieldGlyphs(t *testing.T) {
	md := "# Prüfung\n\nStatus ✔: [@#status __✔ ok ⇒ weiter__]\n\nOrt: [@#ort __Kühlraum 4 °C__]\n"
	var field, text []Diagnostic
	for _, d := range Check(md, Options{}) {
		switch d.Code {
		case "pdf-field-glyph-missing":
			field = append(field, d)
		case "pdf-glyph-missing":
			text = append(text, d)
		}
	}
	if len(text) != 0 {
		t.Errorf("text: %v", text)
	}
	if len(field) != 1 || field[0].Line != 3 || !strings.Contains(field[0].Message, "„✔“, „⇒“") || !strings.Contains(field[0].Message, "`status`") {
		t.Errorf("field: %v", field)
	}
}
