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
	Group   string `json:"group,omitempty"`
	Pattern string `json:"pattern,omitempty"`
	Tooltip string `json:"tooltip,omitempty"`
	Chars   int    `json:"chars,omitempty"`
	Rows    int    `json:"rows,omitempty"`
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
		}
		out = append(out, ff)
	}
	return out
}

var (
	reComment  = regexp.MustCompile(`(?s)<!--.*?-->`)
	reListMark = regexp.MustCompile(`^\s*(?:[*+-]|\d+[.)])\s+`)
	reEmphasis = regexp.MustCompile(`\*\*|__|\*|` + "`")
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
	for i := line - 2; i >= 0 && i < len(ls); i-- {
		if t := labelText(ls[i]); t != "" && !strings.HasPrefix(t, "[@") {
			if h, ok := atxHeading(ls[i]); ok {
				return plainText(h)
			}
			return t
		}
		if strings.TrimSpace(ls[i]) != "" {
			return ""
		}
	}
	return ""
}

// groupAbove: for an indented list item the parent item's text.
func groupAbove(ls []string, line int) string {
	if line-1 >= len(ls) || !reListMark.MatchString(ls[line-1]) {
		return ""
	}
	indent := indentOf(ls[line-1])
	if indent == 0 {
		return ""
	}
	for i := line - 2; i >= 0; i-- {
		if strings.TrimSpace(ls[i]) == "" {
			continue
		}
		if indentOf(ls[i]) < indent && reListMark.MatchString(ls[i]) {
			return labelText(ls[i])
		}
	}
	return ""
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
