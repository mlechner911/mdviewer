package mlcext

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// FieldSpec is a form field as an editor's field dialog sees it: what the
// user chooses instead of writing the syntax. FormatField writes it,
// FieldAt reads the field at a position back. The vectors in
// testdata/fieldspec.json pin the written form for a TypeScript port.
type FieldSpec struct {
	// Kind: "text" (one line), "multiline", "check", "pattern" (TT.MM.JJJJ,
	// HH:MM or an own pattern) or "signature".
	Kind string `json:"kind"`
	// Name without "#"; empty for an unnamed field (named after the
	// section when filled in).
	Name string `json:"name"`
	// Chars: width of a text field in characters (multiline: of each line).
	Chars int `json:"chars,omitempty"`
	// Rows of a multiline field (the minimum when FillPage).
	Rows int `json:"rows,omitempty"`
	// FillPage: the multiline field grows to the end of the page.
	FillPage bool `json:"fillPage,omitempty"`
	// Pattern of a pattern field, e.g. "TT.MM.JJJJ".
	Pattern string `json:"pattern,omitempty"`
	// Value is the preset (text, multiline: lines joined with "\n",
	// pattern); the code escapes it.
	Value string `json:"value,omitempty"`
	// Checked: the check box opens ticked.
	Checked bool `json:"checked,omitempty"`
	// Role of a signature field: who signs.
	Role string `json:"role,omitempty"`
	// Tooltip written as "<!--* … -->" right after the field.
	Tooltip string `json:"tooltip,omitempty"`
}

// Defaults of the field dialog.
const (
	DefaultFieldChars = 20
	DefaultFieldRows  = 3
)

// FormatField writes s in the syntax of the MLC Markdown Extension, with
// its tooltip comment; the error says what the dialog has to fix.
func FormatField(s FieldSpec) (string, error) {
	if err := checkFieldName(s.Name); err != nil {
		return "", err
	}
	head := "[@"
	if s.Name != "" {
		head += "#" + s.Name
	}
	var body string
	switch s.Kind {
	case "check":
		if s.Checked {
			body = "x"
		}
	case "signature":
		body = "Unterschrift: " + MaskFieldValue(oneLine(s.Role))
	case "pattern":
		p := oneLine(s.Pattern)
		switch {
		case p == "":
			return "", errors.New("Muster fehlt (z. B. TT.MM.JJJJ oder HH:MM).")
		case strings.Contains(p, "_"):
			return "", errors.New("Ein Muster enthält kein `_` (das wäre eine Linie).")
		case utf8.RuneCountInString(p) > 40:
			return "", errors.New("Das Muster ist länger als 40 Zeichen.")
		}
		body = maskPattern(p)
		if v := oneLine(s.Value); v != "" {
			if utf8.RuneCountInString(v) > utf8.RuneCountInString(p) {
				return "", fmt.Errorf("Die Vorgabe ist länger als das Muster (%d Zeichen).", utf8.RuneCountInString(p))
			}
			body += "=" + MaskFieldValue(v)
		}
	case "multiline":
		rows := max(s.Rows, 1)
		width := s.Chars
		if width <= 0 {
			width = DefaultFieldChars * 2
		}
		values := strings.Split(strings.ReplaceAll(s.Value, "\r\n", "\n"), "\n")
		lines := make([]string, max(rows, len(values)))
		for i := range lines {
			v := ""
			if i < len(values) {
				v = strings.TrimSpace(values[i])
			}
			lines[i] = textLine(v, width)
		}
		if s.FillPage {
			lines = append(lines, "…")
		}
		body = strings.Join(lines, "\n")
	case "text", "":
		width := s.Chars
		if width <= 0 {
			width = DefaultFieldChars
		}
		body = textLine(oneLine(s.Value), width)
	default:
		return "", fmt.Errorf("Unbekannte Feldart %q.", s.Kind)
	}
	// "[@#name ____]" needs the blank after the name, "[@____]" none.
	out := head
	if s.Name != "" && body != "" {
		out += " "
	}
	out += body + "]"
	if t := oneLine(s.Tooltip); t != "" {
		if strings.Contains(t, "-->") {
			return "", errors.New("Der Tooltip darf kein `-->` enthalten.")
		}
		out += " <!--* " + t + " -->"
	}
	return out, nil
}

// textLine is one line of a text field: underscores of width, or the
// value between them ("__Max Mustermann__"), as FillFields writes it. The
// width counts the value unescaped, as the scanner does.
func textLine(v string, width int) string {
	if v == "" {
		return strings.Repeat("_", max(width, 4))
	}
	return "__" + MaskFieldValue(v) + strings.Repeat("_", max(width-2-utf8.RuneCountInString(v), 2))
}

// maskPattern escapes what would end the field or start a preset.
func maskPattern(p string) string {
	var b strings.Builder
	for _, r := range p {
		if strings.ContainsRune(`\]=|[`, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// checkFieldName: letters, digits, "-" and "_" (ASCII, as FillFields and
// PDF field names expect); empty for an unnamed field.
func checkFieldName(name string) error {
	for _, r := range name {
		if !(r == '-' || r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return fmt.Errorf("Feldname %q: nur Buchstaben a–z, Ziffern, `-` und `_` (keine Umlaute, kein Punkt).", name)
		}
	}
	return nil
}

// FieldEdit is the field at a position, for the field dialog: its spec and
// the range to replace (1-based lines, columns in characters, End
// exclusive), its tooltip comment included.
type FieldEdit struct {
	OK      bool      `json:"ok"`
	Spec    FieldSpec `json:"spec"`
	Line    int       `json:"line"`
	Col     int       `json:"col"`
	EndLine int       `json:"endLine"`
	EndCol  int       `json:"endCol"`
}

// FieldAt reads the field at line and col (1-based, characters; the
// field's tooltip comment counts as part of it). OK is false when there
// is no well-formed field there.
func FieldAt(md string, line, col int) FieldEdit {
	s := &scanner{src: md, section: "feld"}
	s.run()
	at := offsetOf(md, line, col)
	for _, f := range s.fields {
		end := f.End
		hasTooltip := false
		for _, h := range s.hints {
			if h.Kind == "tooltip" && h.Offset >= f.End && strings.TrimSpace(md[f.End:h.Offset]) == "" {
				end, hasTooltip = h.End, true
				break
			}
		}
		if at < f.Offset || at > end {
			continue
		}
		e := FieldEdit{OK: true, Spec: specOf(f, hasTooltip)}
		e.Line, e.Col = position(md, f.Offset)
		e.EndLine, e.EndCol = position(md, end)
		return e
	}
	return FieldEdit{}
}

// specOf turns a scanned field into its spec.
func specOf(f Field, hasTooltip bool) FieldSpec {
	s := FieldSpec{}
	if f.Named {
		s.Name = f.Name
	}
	switch f.Kind {
	case CheckBox:
		s.Kind, s.Checked = "check", f.Checked
	case SignatureField:
		s.Kind, s.Role = "signature", f.Role
	case MultilineField:
		s.Kind, s.Rows, s.FillPage, s.Value = "multiline", f.Rows, f.FillPage, f.Value
		s.Chars = f.Chars
	default:
		if f.Pattern != "" {
			s.Kind, s.Pattern, s.Value = "pattern", f.Pattern, f.Value
		} else {
			s.Kind, s.Chars, s.Value = "text", f.Chars, f.Value
		}
	}
	if hasTooltip {
		s.Tooltip = f.Tooltip
	}
	return s
}

// offsetOf is the byte offset of line and col (1-based, col in characters).
func offsetOf(src string, line, col int) int {
	off := 0
	for l := 1; l < line; l++ {
		i := strings.IndexByte(src[off:], '\n')
		if i < 0 {
			return len(src)
		}
		off += i + 1
	}
	for c := 1; c < col && off < len(src) && src[off] != '\n'; c++ {
		_, size := utf8.DecodeRuneInString(src[off:])
		off += size
	}
	return off
}
