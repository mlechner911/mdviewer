package mlcext

import (
	"regexp"
	"strconv"
	"strings"
)

// FormField describes one field of a form for an input mask (MarkSafe fill
// mode, the Sensor Monitor's form dialog): what to ask, under which heading,
// and what is already entered. The labels are a heuristic over the
// surrounding text — the form itself has no label syntax, so a mask shows
// them next to the field name, never instead of it.
type FormField struct {
	Name  string `json:"name"`
	Named bool   `json:"named"`
	Kind  string `json:"kind"` // "text" | "multiline" | "check" | "signature"
	// Label: text before the field on its line (up to a previous field), for
	// a check box the text after it; in a table the column header plus the
	// row ("Charge (1)"); for a field alone on its line the line above.
	Label string `json:"label"`
	// Section is the ATX heading above the field.
	Section string `json:"section,omitempty"`
	// Group is the list item a nested item belongs to ("2. Zeuge" above
	// "- Name:"), empty otherwise.
	Group string `json:"group,omitempty"`
	// Question of a check box: the text before the first box of its line
	// when every field up to this one is a box ("Wurde X umgesetzt? [@#x_ja]
	// Ja [@#x_nein] Nein" → "Wurde X umgesetzt?"); Label stays the answer.
	// Empty when there is no such text or it is the label already.
	Question string `json:"question,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
	Tooltip  string `json:"tooltip,omitempty"`
	Chars    int    `json:"chars,omitempty"`
	Rows     int    `json:"rows,omitempty"`
	// Role of a signature field: who signs. A mask shows it as "to be
	// signed by …", never as an input.
	Role string `json:"role,omitempty"`
	// Prefill: the Sensor Monitor fills this field from an event.
	Prefill bool `json:"prefill"`
	// Value as FieldValues reads it ("x" or "" for a check box).
	Value string `json:"value"`
	Line  int    `json:"line"`
}

// FormSchema describes every field of md in document order.
func FormSchema(md string) []FormField {
	fields, _ := ExtractFields(md)
	srcLines := strings.Split(md, "\n")
	out := make([]FormField, 0, len(fields))
	for i, f := range fields {
		ff := FormField{
			Name: f.Name, Named: f.Named, Kind: f.Kind.String(), Pattern: f.Pattern,
			Tooltip: f.Tooltip, Chars: f.Chars, Rows: f.Rows, Line: f.Line,
			Prefill: f.Named && IsPrefillField(f.Name), Value: f.Value, Role: f.Role,
		}
		if f.Kind == CheckBox {
			ff.Rows, ff.Value = 0, ""
			if f.Checked {
				ff.Value = "x"
			}
		}
		ff.Section = sectionAbove(srcLines, f.Line)
		lineStart := strings.LastIndexByte(md[:f.Offset], '\n') + 1
		lineEnd := len(md)
		if k := strings.IndexByte(md[f.End:], '\n'); k >= 0 {
			lineEnd = f.End + k
		}
		line := md[lineStart:lineEnd]
		if strings.HasPrefix(strings.TrimSpace(line), "|") {
			ff.Label = tableLabel(srcLines, f.Line, md[lineStart:f.Offset])
		} else {
			from, to := lineStart, lineEnd
			if i > 0 && fields[i-1].End > from {
				from = fields[i-1].End
			}
			if i+1 < len(fields) && fields[i+1].Offset < to {
				to = fields[i+1].Offset
			}
			before, after := labelText(md[from:f.Offset]), labelText(md[f.End:to])
			// "Kontrolle: [@#k] Ort: [@#o ___]": the box has its label, "Ort" belongs
			// to the next field. Without one, "[@#s] Sonstiges: [@#t ___]" labels the box.
			if before != "" && to < lineEnd && strings.HasSuffix(strings.TrimSpace(md[f.End:to]), ":") {
				after = ""
			}
			switch {
			case f.Kind == SignatureField && f.Role != "":
				ff.Label = f.Role
			case f.Kind == CheckBox && after != "":
				ff.Label = shortLabel(after)
			case before != "":
				ff.Label = before
			case after != "":
				ff.Label = shortLabel(after)
			default:
				ff.Label = lineAbove(srcLines, f.Line)
			}
			ff.Group = groupAbove(srcLines, f.Line)
			if f.Kind == CheckBox {
				if q := checkQuestion(md, fields, i, lineStart); q != ff.Label {
					ff.Question = q
				}
			}
		}
		out = append(out, ff)
	}
	return out
}

// FormBlock is one part of a form in document order: text or a field.
type FormBlock struct {
	// Text is Markdown between fields, for reading.
	Text string `json:"text,omitempty"`
	// Field is nil for text.
	Field *FormField `json:"field,omitempty"`
}

// FormLayout is the whole form for an input mask: the fields of FormSchema
// with the document's text in between, so whoever fills it in also reads its
// instructions. Text a field already shows as its label is left out; so are
// tables with fields, whose cells become labels. Fill hints stay.
func FormLayout(md string) []FormBlock {
	md = ExpandVariables(RemoveSignature(md))
	fields, _ := ExtractFields(md)
	schema := FormSchema(md)
	ls := strings.Split(md, "\n")
	used := make([]bool, len(ls))
	mark := func(i int) {
		if i >= 0 && i < len(used) {
			used[i] = true
		}
	}
	leading, trailing := map[int]string{}, map[int]string{}
	for i, f := range fields {
		ff := schema[i]
		endLine, _ := position(md, f.End)
		for l := f.Line; l <= endLine; l++ {
			mark(l - 1)
		}
		if isTableLine(ls[f.Line-1]) {
			for k := f.Line - 2; k >= 0 && isTableLine(ls[k]); k-- {
				mark(k)
			}
			for k := f.Line; k < len(ls) && isTableLine(ls[k]); k++ {
				mark(k)
			}
			continue
		}
		mark(groupAboveAt(ls, f.Line))
		lineStart := strings.LastIndexByte(md[:f.Offset], '\n') + 1
		lineEnd := len(md)
		if k := strings.IndexByte(md[f.End:], '\n'); k >= 0 {
			lineEnd = f.End + k
		}
		first := i == 0 || fields[i-1].End <= lineStart
		last := i+1 == len(fields) || fields[i+1].Offset >= lineEnd
		before, after := labelText(md[lineStart:f.Offset]), labelText(md[f.End:lineEnd])
		if first && before != "" && before != ff.Label && f.Kind != SignatureField {
			leading[i] = strings.TrimSpace(reListMark.ReplaceAllString(reComment.ReplaceAllString(md[lineStart:f.Offset], ""), ""))
		}
		if last && after != "" && after != ff.Label {
			trailing[i] = strings.TrimSpace(reComment.ReplaceAllString(md[f.End:lineEnd], ""))
		}
		if first && last && before == "" && after == "" && ff.Label != "" && ff.Label == lineAbove(ls, f.Line) {
			if k := lineAboveAt(ls, f.Line); k >= 0 {
				if _, heading := atxHeading(ls[k]); !heading {
					mark(k)
				}
			}
		}
	}
	_, _, bodyLine := splitFrontMatter(md)
	out := []FormBlock{}
	var text []string
	addText := func(s string) {
		s = strings.TrimSpace(reHiddenComment.ReplaceAllString(s, ""))
		if s != "" {
			out = append(out, FormBlock{Text: s})
		}
	}
	flush := func() {
		addText(strings.Join(text, "\n"))
		text = nil
	}
	next := 0
	for n := bodyLine; n < len(ls); n++ {
		for next < len(fields) && fields[next].Line-1 == n {
			flush()
			if t := leading[next]; t != "" {
				addText(t)
			}
			out = append(out, FormBlock{Field: &schema[next]})
			if t := trailing[next]; t != "" {
				addText(t)
			}
			next++
		}
		if !used[n] {
			text = append(text, ls[n])
		}
	}
	flush()
	return out
}

// checkQuestion is the text before the first field on the line of check box
// fields[i], "" when a field other than a check box comes before it.
func checkQuestion(md string, fields []Field, i, lineStart int) string {
	j := i
	for j > 0 && fields[j-1].End > lineStart {
		j--
		if fields[j].Kind != CheckBox {
			return ""
		}
	}
	return labelText(md[lineStart:fields[j].Offset])
}

func isTableLine(s string) bool { return strings.HasPrefix(strings.TrimSpace(s), "|") }

var (
	// reHiddenComment: every comment but a fill hint "<!--! … -->".
	reHiddenComment = regexp.MustCompile(`(?s)<!--(?:[^!].*?)?-->`)
	reComment       = regexp.MustCompile(`(?s)<!--.*?-->`)
	reListMark      = regexp.MustCompile(`^\s*(?:[*+-]|\d+[.)])\s+`)
	reEmphasis      = regexp.MustCompile(`\*\*|__|\*|` + "`")
)

// labelText cleans a piece of a Markdown line into label text: comments,
// list markers and emphasis removed, a trailing ":" dropped.
func labelText(s string) string {
	return plainText(reListMark.ReplaceAllString(s, ""))
}

// plainText: comments and emphasis removed, escapes resolved, a trailing
// ":" dropped — for headings, whose "1." is part of the title.
func plainText(s string) string {
	s = reComment.ReplaceAllString(s, " ")
	s = reEmphasis.ReplaceAllString(s, "")
	s = strings.Join(strings.Fields(unescape(s)), " ")
	return strings.TrimSpace(strings.TrimRight(s, ": "))
}

// shortLabel is the text after a check box up to its first ":" or sentence
// end, at most 60 characters.
func shortLabel(s string) string {
	if k := strings.IndexAny(s, ":;"); k > 0 {
		s = s[:k]
	} else if k := strings.Index(s, ". "); k > 0 {
		s = s[:k]
	}
	if r := []rune(s); len(r) > 60 {
		s = strings.TrimSpace(string(r[:59])) + "…"
	}
	return strings.TrimSpace(s)
}

func sectionAbove(ls []string, line int) string {
	inFence := ""
	section := ""
	for i := 0; i < line-1 && i < len(ls); i++ {
		if f := codeFence(ls[i]); f != "" {
			switch {
			case inFence == "":
				inFence = f
			case strings.HasPrefix(strings.TrimSpace(ls[i]), inFence):
				inFence = ""
			}
			continue
		}
		if inFence != "" {
			continue
		}
		if t, ok := atxHeading(ls[i]); ok {
			section = plainText(t)
		}
	}
	return section
}

// lineAbove is the nearest non-empty line above (a field alone on its line,
// e.g. a multi-line text area under its question).
func lineAbove(ls []string, line int) string {
	i := lineAboveAt(ls, line)
	if i < 0 {
		return ""
	}
	if h, ok := atxHeading(ls[i]); ok {
		return plainText(h)
	}
	return labelText(ls[i])
}

// lineAboveAt is the index of that line, -1 when there is none.
func lineAboveAt(ls []string, line int) int {
	for i := line - 2; i >= 0 && i < len(ls); i-- {
		if t := labelText(ls[i]); t != "" && !strings.HasPrefix(t, "[@") {
			return i
		}
		if strings.TrimSpace(ls[i]) != "" {
			return -1
		}
	}
	return -1
}

// groupAbove: for an indented list item the parent item's text.
func groupAbove(ls []string, line int) string {
	if i := groupAboveAt(ls, line); i >= 0 {
		return labelText(ls[i])
	}
	return ""
}

// groupAboveAt is the index of that parent item, -1 when there is none.
func groupAboveAt(ls []string, line int) int {
	if line-1 >= len(ls) || !reListMark.MatchString(ls[line-1]) {
		return -1
	}
	indent := indentOf(ls[line-1])
	if indent == 0 {
		return -1
	}
	for i := line - 2; i >= 0; i-- {
		if strings.TrimSpace(ls[i]) == "" {
			continue
		}
		if indentOf(ls[i]) < indent && reListMark.MatchString(ls[i]) {
			return i
		}
	}
	return -1
}

func indentOf(s string) int { return len(s) - len(strings.TrimLeft(s, " \t")) }

// tableLabel: column header of the field's cell plus the row — the first
// cell's text when it has no field, else the row number.
func tableLabel(ls []string, line int, before string) string {
	col := strings.Count(strings.TrimLeft(before, " \t"), "|") - strings.Count(before, `\|`) - 1
	top := line - 1
	for top > 0 && strings.HasPrefix(strings.TrimSpace(ls[top-1]), "|") {
		top--
	}
	header := tableCells(ls[top])
	label := ""
	if col >= 0 && col < len(header) {
		label = labelText(header[col])
	}
	row := tableCells(ls[line-1])
	key := ""
	if col != 0 && len(row) > 0 && !strings.Contains(row[0], "[@") {
		key = labelText(row[0])
	}
	if key == "" {
		key = strconv.Itoa(line - 1 - top - 1) // header and separator line
	}
	if label == "" {
		return key
	}
	return label + " (" + key + ")"
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
