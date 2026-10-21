package mlcext

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Column widths of tables in the PDF follow mlcmd2pdf (like Pandoc): a
// separator row up to 72 characters sizes the columns by their content; a
// longer one makes the table full width and gives every column a share in
// proportion to its dashes. Too few dashes for a column make its text
// unreadable in the PDF; the check warns and FitTableColumns sets them.

const (
	// tableWideSeparator: from this length on, dashes set the widths.
	tableWideSeparator = 72
	// tableCharsFullWidth: about this many characters of table text fit
	// on the width of an A4 page (8.5 pt Liberation Sans, 17 mm margins).
	tableCharsFullWidth = 110
)

var (
	reTableSep  = regexp.MustCompile(`^\s*\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)*\|?\s*$`)
	reCellMark  = regexp.MustCompile("\\*\\*|__|`|\\\\(.)")
	reCellLink  = regexp.MustCompile(`!?\[([^\]]*)\]\([^)]*\)`)
	reCellBreak = regexp.MustCompile(`(?i)<br\s*/?>`)
	reCellTag   = regexp.MustCompile(`<[^>]+>`)
	reCellField = regexp.MustCompile(`\[@(?:#[^\s\]]*)?\s*([^\]]*)\]`)
)

// fieldWidth: a form field measured as it prints, as a placeholder word
// (a check box about two characters, a line as long as its underscores or
// its pattern).
func fieldWidth(m string) string {
	sm := reCellField.FindStringSubmatch(m)
	body := strings.TrimSpace(sm[1])
	n := 2
	if u := strings.Count(body, "_"); u > 0 {
		n = u
	} else if body != "" {
		n = utf8.RuneCountInString(strings.SplitN(body, "=", 2)[0])
	}
	return strings.Repeat("x", n)
}

type mdTable struct {
	header, sep, end int // 0-based lines: header row, separator row, last row
	cols             int
	align            []string // "", "left", "right", "center"
	rows             [][]string
}

// splitRow splits a table row at unescaped "|" outside inline code.
func splitRow(l string) []string {
	l = strings.TrimSpace(l)
	l = strings.TrimPrefix(l, "|")
	if strings.HasSuffix(l, "|") && !strings.HasSuffix(l, `\|`) {
		l = l[:len(l)-1]
	}
	var cells []string
	var cur strings.Builder
	code := false
	for i := 0; i < len(l); i++ {
		c := l[i]
		switch {
		case c == '\\' && i+1 < len(l):
			cur.WriteByte(c)
			cur.WriteByte(l[i+1])
			i++
			continue
		case c == '`':
			code = !code
		case c == '|' && !code:
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
			continue
		}
		cur.WriteByte(c)
	}
	return append(cells, strings.TrimSpace(cur.String()))
}

// findTables finds the pipe tables of md (outside code blocks).
func findTables(md string) []mdTable {
	ls := strings.Split(md, "\n")
	var out []mdTable
	inFence, fence := false, ""
	for i := 0; i < len(ls); i++ {
		l := strings.TrimRight(ls[i], "\r")
		if f := codeFence(l); f != "" && (!inFence || strings.HasPrefix(strings.TrimSpace(l), fence)) {
			inFence, fence = !inFence, f
			continue
		}
		if inFence || i == 0 || !strings.Contains(l, "-") || !reTableSep.MatchString(l) || !strings.Contains(ls[i-1], "|") {
			continue
		}
		sepCells := splitRow(l)
		head := splitRow(ls[i-1])
		if len(sepCells) != len(head) {
			continue
		}
		t := mdTable{header: i - 1, sep: i, end: i, cols: len(head)}
		for _, c := range sepCells {
			c = strings.TrimSpace(c)
			switch {
			case strings.HasPrefix(c, ":") && strings.HasSuffix(c, ":"):
				t.align = append(t.align, "center")
			case strings.HasSuffix(c, ":"):
				t.align = append(t.align, "right")
			case strings.HasPrefix(c, ":"):
				t.align = append(t.align, "left")
			default:
				t.align = append(t.align, "")
			}
		}
		t.rows = append(t.rows, head)
		for j := i + 1; j < len(ls) && strings.TrimSpace(ls[j]) != "" && strings.Contains(ls[j], "|"); j++ {
			t.rows = append(t.rows, splitRow(strings.TrimRight(ls[j], "\r")))
			t.end = j
		}
		out = append(out, t)
		i = t.end
	}
	return out
}

// cellLines is a cell as printed: markup removed, split at <br>.
func cellLines(c string) []string {
	c = reCellField.ReplaceAllStringFunc(c, fieldWidth)
	c = reCellLink.ReplaceAllString(c, "$1")
	parts := reCellBreak.Split(c, -1)
	for i, p := range parts {
		p = reCellTag.ReplaceAllString(p, "")
		parts[i] = strings.TrimSpace(reCellMark.ReplaceAllString(p, "$1"))
	}
	return parts
}

// columnStats: the longest printed line and the longest word per column.
func (t mdTable) columnStats() (longest, word []int, wordText []string) {
	longest, word, wordText = make([]int, t.cols), make([]int, t.cols), make([]string, t.cols)
	for _, r := range t.rows {
		for c := 0; c < t.cols && c < len(r); c++ {
			for _, l := range cellLines(r[c]) {
				longest[c] = max(longest[c], utf8.RuneCountInString(l))
				for _, w := range strings.Fields(l) {
					if n := utf8.RuneCountInString(w); n > word[c] {
						word[c], wordText[c] = n, w
					}
				}
			}
		}
	}
	return
}

func (t mdTable) separatorRow(dashes []int) string {
	cells := make([]string, t.cols)
	for i, n := range dashes {
		d := strings.Repeat("-", n)
		switch t.align[i] {
		case "left":
			d = ":" + d
		case "right":
			d += ":"
		case "center":
			d = ":" + d + ":"
		}
		cells[i] = d
	}
	return "| " + strings.Join(cells, " | ") + " |"
}

// contentDashes gives every column dashes in proportion to its longest
// line (bounded, so one long description does not squeeze the rest), with
// a total that makes the table full width.
func (t mdTable) contentDashes() []int {
	longest, word, _ := t.columnStats()
	weights := make([]float64, t.cols)
	sum := 0.0
	for i := range weights {
		w := float64(min(max(longest[i], word[i], 4), 40))
		weights[i] = w
		sum += w
	}
	dashes := make([]int, t.cols)
	for i, w := range weights {
		dashes[i] = max(3, int(math.Round(w/sum*100)))
	}
	return dashes
}

// FitTableColumns rewrites the separator row of the table at line (1-based,
// any line of the table): mode "auto" sets "---" per column (mlcdoc sizes
// the columns by their content), mode "content" makes the table full width
// with shares after the content of each column. Alignment colons stay. ok
// is false when line is in no table.
func FitTableColumns(md string, line int, mode string) (string, bool) {
	ls := strings.Split(md, "\n")
	for _, t := range findTables(md) {
		if line-1 < t.header || line-1 > t.end {
			continue
		}
		dashes := make([]int, t.cols)
		for i := range dashes {
			dashes[i] = 3
		}
		if mode == "content" {
			dashes = t.contentDashes()
		}
		row := t.separatorRow(dashes)
		if strings.HasSuffix(ls[t.sep], "\r") {
			row += "\r"
		}
		ls[t.sep] = row
		return strings.Join(ls, "\n"), true
	}
	return md, false
}

// checkTables warns when a column of a full-width table is too narrow in
// the PDF for its longest word (an estimate from the dashes).
func checkTables(md string, add func(int, int, Severity, string, string)) {
	ls := strings.Split(md, "\n")
	for _, t := range findTables(md) {
		sep := strings.TrimSpace(ls[t.sep])
		if utf8.RuneCountInString(sep) <= tableWideSeparator {
			continue
		}
		cells := splitRow(sep)
		total := 0
		dash := make([]int, len(cells))
		for i, c := range cells {
			dash[i] = strings.Count(c, "-")
			total += dash[i]
		}
		longest, word, wordText := t.columnStats()
		for i := 0; i < t.cols && i < len(dash); i++ {
			room := max(int(float64(dash[i])/float64(total)*tableCharsFullWidth)-2, 1)
			head := ""
			if len(t.rows) > 0 && i < len(t.rows[0]) {
				head = strings.Join(cellLines(t.rows[0][i]), " ")
			}
			const hint = " Die Breiten folgen der Zahl der Striche in dieser Zeile."
			switch {
			case word[i] > room:
				add(t.sep+1, 1, Warning, "table-column-narrow", fmt.Sprintf("Spalte %d („%s“) ist im PDF etwa %d Zeichen breit; „%s“ passt nicht hinein.%s", i+1, head, room, wordText[i], hint))
			case (longest[i]+room-1)/room > 2:
				add(t.sep+1, 1, Warning, "table-column-narrow", fmt.Sprintf("Spalte %d („%s“) ist im PDF etwa %d Zeichen breit; ihr Text bricht auf %d Zeilen um.%s", i+1, head, room, (longest[i]+room-1)/room, hint))
			}
		}
	}
}
