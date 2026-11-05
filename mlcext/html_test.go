package mlcext

import (
	"slices"
	"strings"
	"testing"
)

func TestPDFHTML(t *testing.T) {
	for _, c := range []struct {
		name, in, want string
		notes          []string
	}{
		{"bold and italic", "A <b>fett</b> <strong>stark</strong> <i>k</i> <em>e</em>.", "A **fett** **stark** *k* *e*.", nil},
		{"strike and code", "<s>a</s> <del>b</del> <code>c*d</code>", "~~a~~ ~~b~~ `c*d`", nil},
		{"blanks stay outside the markers", "x<b> fett </b>y", "x **fett** y", nil},
		{"attributes, upper case", `<B class="x">a</B>`, "**a**", nil},
		{"nested", "<b><i>a</i></b>", "***a***", nil},
		{"other inline tags lose their tags", `H<sub>2</sub>O <u>u</u> <span style="color:red">rot</span>`, "H2O u rot", []string{"HTML im PDF ohne Formatierung: <span>, <sub>, <u>"}},
		{"lone tag is stripped", "a <b>offen", "a offen", []string{"HTML im PDF ohne Formatierung: <b>"}},
		{"br, links and comments stay", `a<br>b <a href="https://x.eu">l</a> <!-- c -->`, `a<br>b <a href="https://x.eu">l</a> <!-- c -->`, nil},
		{"block is left out", "vor\n\n<div align=\"center\">\nText\n</div>\n\nnach", "vor\n\n\n\nnach", []string{"HTML erscheint nicht im PDF: <div>"}},
		{"details twice", "<details>a</details>\n\n<details>b</details>\n", "\n\n\n", []string{"HTML erscheint nicht im PDF: <details> (2×)"}},
		{"img alone", `<img src="a.png">`, "", []string{"HTML erscheint nicht im PDF: <img>"}},
		{"page break stays", `<div style="page-break-after: always"></div>`, `<div style="page-break-after: always"></div>`, nil},
		{"code stays", "`<b>x</b>`\n\n```\n<b>y</b>\n<div>\n```", "`<b>x</b>`\n\n```\n<b>y</b>\n<div>\n```", nil},
		{"escaped and autolinks stay", `\<b>x <https://x.eu> <a@b.de>`, `\<b>x <https://x.eu> <a@b.de>`, nil},
		{"front matter stays", "---\ntitle: <b>T</b>\n---\n<b>x</b>", "---\ntitle: <b>T</b>\n---\n**x**", nil},
		{"multi-line comment stays", "<!--\n<b>x</b>\n-->\n<b>y</b>", "<!--\n<b>x</b>\n-->\n**y**", nil},
		{"comment with a blank line stays", "<!--\n\n<b>x</b>\n-->\n<b>y</b>", "<!--\n\n<b>x</b>\n-->\n**y**", nil},
	} {
		got, notes := PDFHTML(c.in)
		if got != c.want || !slices.Equal(notes, c.notes) {
			t.Errorf("%s:\n got %q %q\nwant %q %q", c.name, got, notes, c.want, c.notes)
		}
	}
}

func TestHTMLDiagnostics(t *testing.T) {
	md := "# T\n\n<details>\nx\n</details>\n\nA <u>u</u> <b>b</b> <img src=\"a.png\"> ok\n"
	var got []string
	for _, d := range Check(md, Options{PDF: true}) {
		if strings.HasPrefix(d.Code, "html-") {
			got = append(got, d.Code+"@"+string(rune('0'+d.Line))+":"+string(rune('0'+d.Col)))
		}
	}
	want := []string{"html-pdf-missing@3:1", "html-pdf-plain@7:3", "html-pdf-missing@7:" + string(rune('0'+21))}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	for _, d := range Check(md, Options{}) {
		if strings.HasPrefix(d.Code, "html-") {
			t.Errorf("without PDF: %+v", d)
		}
	}
	a := AnalyzeWith("Ein README mit <u>u</u>.", Options{PDF: true})
	if a.Detected || len(a.Diagnostics) != 1 {
		t.Errorf("outside the extension: %+v", a)
	}
	if a := AnalyzeWith("Ein README mit <u>u</u>.", Options{}); len(a.Diagnostics) != 0 {
		t.Errorf("without PDF: %+v", a.Diagnostics)
	}
}
