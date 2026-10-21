package mlcext

import (
	"strings"
	"testing"
)

const reportTable = "| Metrik | Wert | Interpretation / Anomalie |\n" +
	"| :---------------- | ----------: | --------------------------------------------------------------------------------: |\n" +
	"| **Gesamtzahl Alarme** | 42 | Stabil im Vergleich zum Vormonat (45 Alarme). |\n" +
	"| **Z-Score Peak-Slots** | Di 08:00h (Z=4.2)<br>Do 08:00h (Z=3.8) | Statistisch signifikante Häufungen exakt zur Warenannahme / Bestückung. |\n"

func TestTableColumnTooNarrow(t *testing.T) {
	md := "# Bericht\n\n" + reportTable
	var msgs []string
	for _, d := range Check(md, Options{}) {
		if d.Code == "table-column-narrow" {
			msgs = append(msgs, d.Message)
			if d.Line != 4 {
				t.Errorf("line %d", d.Line)
			}
		}
	}
	if len(msgs) != 1 || !strings.Contains(msgs[0], "Spalte 2 („Wert“)") {
		t.Errorf("want a warning for column 2: %q", msgs)
	}
}

func TestFitTableColumns(t *testing.T) {
	md := "# Bericht\n\n" + reportTable + "\nText\n"
	auto, ok := FitTableColumns(md, 5, "auto")
	if !ok || !strings.Contains(auto, "\n| :--- | ---: | ---: |\n") {
		t.Fatalf("auto:\n%s", auto)
	}
	content, ok := FitTableColumns(md, 3, "content")
	if !ok {
		t.Fatal("content: no table")
	}
	sep := strings.Split(content, "\n")[3]
	cells := splitRow(sep)
	if !strings.HasPrefix(cells[0], ":") || !strings.HasSuffix(cells[1], ":") || len(sep) <= 72 {
		t.Errorf("alignment or full width lost: %s", sep)
	}
	for _, d := range Check(content, Options{}) {
		if d.Code == "table-column-narrow" {
			t.Errorf("still too narrow after the fix: %s\n%s", d.Message, sep)
		}
	}
	if _, ok := FitTableColumns(md, 1, "auto"); ok {
		t.Error("line 1 is no table")
	}
	// Tables in code blocks are left alone.
	if len(findTables("```\n| a | b |\n| - | - |\n```\n")) != 0 {
		t.Error("table inside a code block")
	}
}
