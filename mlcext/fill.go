package mlcext

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Filling a form: values go INTO the "[@…]" fields as their preset (text) or
// as "[@x]" (a ticked check box), so the result is still the same form — it
// renders in mlcdoc as a PDF/A whose fields show what was entered, and
// FieldValues reads the values back. The Sensor Monitor prefills a form from
// an event and stores a filled-in form as a record; MarkSafe fills one in
// (plan P-SENS-20261028-01). Ported from mlcsensormonitor internal/formfill.
//
// Rules (mlcmd2pdf):
//   - Every field can be filled, named ("[@#name …]") or not (by its
//     generated name "<section>_<n>", see Field.Name).
//   - Check box: "x" (also "X", "true", "1", "ja", "yes") ticks it, anything
//     else unticks it.
//   - Pattern field "[@#d TT.MM.JJJJ]" → "[@#d TT.MM.JJJJ=16\.10\.2026]".
//   - Text field: one value line per marker line, underscores before and
//     after are the line; the field keeps its width. More value lines than
//     marker lines make the field longer; a one-line field gets the lines
//     joined with spaces.
//   - Every ASCII punctuation character is escaped with "\" — no value can
//     end a field or start a new one; characters without a glyph in the
//     field font are replaced (MaskFieldValue).

// FillFields sets values (field name → value, lines separated by "\n") into
// md and returns the result and the names of the fields it changed, in
// document order. Only names present in values are touched: an empty text
// value clears the field, an empty check box value unticks it. Fields of an
// unknown name stay as they are.
func FillFields(md string, values map[string]string) (string, []string) {
	fields, _ := ExtractFields(md)
	var out strings.Builder
	var set []string
	last := 0
	for _, f := range fields {
		v, ok := values[f.Name]
		if !ok || f.Offset < last || f.End > len(md) {
			continue
		}
		inner := md[f.Offset+2 : f.End-1]
		rest := strings.TrimLeft(inner, " \t")
		prefix := inner[:len(inner)-len(rest)]
		if f.Named {
			name, r, ok := splitName(inner)
			if !ok || name != f.Name {
				continue
			}
			prefix, rest = "#"+name+" ", r
		}
		var repl string
		if f.Kind == CheckBox {
			if truthy(v) {
				repl = "x"
			}
			if repl == "" {
				prefix = strings.TrimRight(prefix, " \t")
			}
		} else {
			repl = fillText(rest, strings.TrimSpace(v))
		}
		out.WriteString(md[last:f.Offset])
		out.WriteString("[@" + prefix + repl + "]")
		last = f.End
		set = append(set, f.Name)
	}
	out.WriteString(md[last:])
	return out.String(), set
}

// FieldValues reads the values of every field (name → value): the preset of
// a text field (lines joined with "\n"), "x" for a ticked check box, "" for
// an empty one. Filling md with FieldValues(md) gives the same values back.
func FieldValues(md string) map[string]string {
	fields, _ := ExtractFields(md)
	out := make(map[string]string, len(fields))
	for _, f := range fields {
		if f.Kind == CheckBox {
			if f.Checked {
				out[f.Name] = "x"
			} else {
				out[f.Name] = ""
			}
			continue
		}
		out[f.Name] = f.Value
	}
	return out
}

// IsFormRecord reports a filled-in form: "meta.schema" is FormRecordSchema.
func IsFormRecord(md string) bool {
	_, _, node, err := parseFrontMatter(md)
	if err != nil || node == nil || len(node.Content) == 0 {
		return false
	}
	var fm struct {
		Meta struct {
			Schema string `yaml:"schema"`
		} `yaml:"meta"`
	}
	return node.Content[0].Decode(&fm) == nil && fm.Meta.Schema == FormRecordSchema
}

func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "x", "true", "1", "ja", "yes", "on":
		return true
	}
	return false
}

// splitName splits the inside of a named marker: "#ort ____" → "ort", "____".
func splitName(inner string) (string, string, bool) {
	if !strings.HasPrefix(inner, "#") {
		return "", "", false
	}
	j := 1
	for j < len(inner) {
		c := inner[j]
		if c == '_' || c == '-' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			j++
			continue
		}
		break
	}
	if j == 1 {
		return "", "", false
	}
	name, rest := inner[1:j], inner[j:]
	if rest != "" && rest[0] != ' ' && rest[0] != '\t' && rest[0] != '\n' {
		return "", "", false
	}
	return name, strings.TrimLeft(rest, " \t"), true
}

// fillText is the new inside (without name) of a text field for value v.
func fillText(rest, v string) string {
	rows := strings.Split(rest, "\n")
	// Pattern (TT.MM.JJJJ, HH:MM), also with a preset already set.
	if len(rows) == 1 && !strings.Contains(rest, "_") && !strings.Contains(rest, "\\") || patternWithValue(rest) {
		pat := strings.TrimSpace(rest)
		if k := strings.IndexByte(pat, '='); k > 0 {
			pat = pat[:k]
		}
		if v == "" {
			return pat
		}
		return pat + "=" + MaskFieldValue(strings.Join(strings.Fields(v), " "))
	}
	width := 0
	for _, r := range rows {
		if n := utf8.RuneCountInString(r); n > width {
			width = n
		}
	}
	lines := strings.Split(strings.ReplaceAll(v, "\r\n", "\n"), "\n")
	if len(rows) == 1 {
		lines = []string{strings.Join(strings.Fields(strings.Join(lines, " ")), " ")}
	}
	n := max(len(rows), len(lines))
	out := make([]string, n)
	for k := 0; k < n; k++ {
		if k >= len(lines) || strings.TrimSpace(lines[k]) == "" {
			out[k] = strings.Repeat("_", max(width, 4))
			continue
		}
		val := "__" + MaskFieldValue(strings.TrimSpace(lines[k]))
		pad := max(width-utf8.RuneCountInString(val), 2)
		out[k] = val + strings.Repeat("_", pad)
	}
	return strings.Join(out, "\n")
}

func patternWithValue(rest string) bool {
	if strings.Contains(rest, "\n") {
		return false
	}
	k := strings.IndexByte(rest, '=')
	return k > 0 && !strings.ContainsAny(rest[:k], "_\\ \t")
}

// glyphReplace: characters without a glyph in the field font (Liberation
// Sans) — mlcmd2pdf warns otherwise.
var glyphReplace = strings.NewReplacer("⇒", "=>", "⇐", "<=", "∅", "Ø", "⌀", "Ø", "∈", "in", "∪", "u", "…", "...")

// MaskFieldValue escapes every ASCII punctuation character with "\", turns
// control characters into blanks and replaces characters the field font
// lacks — a value can neither end a field nor start a new one.
func MaskFieldValue(s string) string {
	s = glyphReplace.Replace(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r < 0x80 && (unicode.IsPunct(r) || unicode.IsSymbol(r)):
			b.WriteByte('\\')
			b.WriteRune(r)
		case unicode.IsControl(r):
			b.WriteByte(' ')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
