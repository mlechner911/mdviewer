package mlcext

import (
	"strconv"
	"strings"
)

// A table row in the source can span several lines: a multi-line field in a
// cell ("| 1 | [@#kg ____\n____] | … |") keeps its line breaks, and mlcmd2pdf
// renders it inside the cell. The helpers here work on such logical rows —
// a source line extended over every line break inside a field — never on
// single source lines.

// span is a byte range [start, end) of the source.
type span struct{ start, end int }

// logicalLine is the source line around pos, extended over line breaks that
// lie inside a field.
func logicalLine(md string, fields []Field, pos int) span {
	s := span{strings.LastIndexByte(md[:pos], '\n') + 1, len(md)}
	if k := strings.IndexByte(md[pos:], '\n'); k >= 0 {
		s.end = pos + k
	}
	for changed := true; changed; {
		changed = false
		for _, f := range fields {
			if f.Offset < s.start && f.End > s.start {
				s.start = strings.LastIndexByte(md[:f.Offset], '\n') + 1
				changed = true
			}
			if f.Offset < s.end && f.End > s.end {
				s.end = len(md)
				if k := strings.IndexByte(md[f.End:], '\n'); k >= 0 {
					s.end = f.End + k
				}
				changed = true
			}
		}
	}
	return s
}

// tableRows are the logical rows of the table that holds pos, from its
// header on, and the index of the row with pos; nil when pos is in no
// table row.
func tableRows(md string, fields []Field, pos int) (rows []span, at int) {
	row := logicalLine(md, fields, pos)
	if !isTableLine(md[row.start:row.end]) {
		return nil, -1
	}
	rows = []span{row}
	for r := row; r.start > 0; {
		if r = logicalLine(md, fields, r.start-1); !isTableLine(md[r.start:r.end]) {
			break
		}
		rows = append([]span{r}, rows...)
	}
	at = len(rows) - 1
	for r := row; r.end < len(md); {
		if r = logicalLine(md, fields, r.end+1); !isTableLine(md[r.start:r.end]) {
			break
		}
		rows = append(rows, r)
	}
	return rows, at
}

// maskFields is md[s] with every field in it written "[@]": a "|" inside a
// field then splits no cell, and a cell with a field still shows one.
func maskFields(md string, fields []Field, s span) string {
	var b strings.Builder
	pos := s.start
	for _, f := range fields {
		if f.Offset < pos || f.End > s.end {
			continue
		}
		b.WriteString(md[pos:f.Offset])
		b.WriteString("[@]")
		pos = f.End
	}
	b.WriteString(md[pos:s.end])
	return b.String()
}

// tableLabel: column header of the cell of fields[i] plus the row — the
// first cell's text when it has no field, else the row number. "" when
// the field is in no table row.
func tableLabel(md string, fields []Field, i int) string {
	f := fields[i]
	rows, at := tableRows(md, fields, f.Offset)
	if rows == nil {
		return ""
	}
	before := maskFields(md, fields, span{rows[at].start, f.Offset})
	col := countUnescaped(strings.TrimLeft(before, " \t"), '|') - 1
	header := tableCells(maskFields(md, fields, rows[0]))
	label := ""
	if col >= 0 && col < len(header) {
		label = labelText(header[col])
	}
	row := tableCells(maskFields(md, fields, rows[at]))
	key := ""
	if col != 0 && len(row) > 0 && !strings.Contains(row[0], "[@") {
		key = labelText(row[0])
	}
	if key == "" {
		key = strconv.Itoa(at - 1) // header and separator row
	}
	if label == "" {
		return key
	}
	return label + " (" + key + ")"
}

// countUnescaped counts the wanted runes not escaped with "\".
func countUnescaped(text string, wanted rune) int {
	n := 0
	for k := unescaped(text, wanted); k >= 0; k = unescaped(text, wanted) {
		n++
		text = text[k+1:]
	}
	return n
}

func tableCells(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimSuffix(strings.TrimPrefix(row, "|"), "|")
	var cells []string
	for {
		k := unescaped(row, '|')
		if k < 0 {
			return append(cells, row)
		}
		cells = append(cells, row[:k])
		row = row[k+1:]
	}
}
