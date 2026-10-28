package mlcext

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Form fields "[@…]" of the MLC Markdown Extension, a faithful port of the
// scanner in mlcmd2pdf (src/form.rs): same kinds, names, presets and the same
// errors, so what passes here renders in mlcdoc.
//
//	[@]                          check box
//	[@x], [@#name x]             check box, ticked (also "X"; mlcmd2pdf 3.27.0)
//	[@__________]                one-line text field, as wide as the source
//	[@___Vorgabe___]             text field with a preset value
//	[@___ … several lines … ___] multi-line text field, one row per line
//	[@HH:MM], [@TT.MM.JJJJ]      text field for the pattern
//	[@TT.MM.JJJJ=16\.10\.2026]   pattern with a preset value
//	multi-line, last line "…"    grows to the end of the page
//	[@#name ___], [@#name]       field named "name"
//	[@___] <!--* Hinweis -->     tooltip, shown on screen, never printed
//	\[@]                         the literal text "[@]"
//
// Code blocks, inline code and HTML comments are left alone, and so is a
// link whose text starts with "@" ("[@team](https://…)").

// FieldKind is what a field holds.
type FieldKind int

const (
	// CheckBox is a check box.
	CheckBox FieldKind = iota
	// TextField is one line of text.
	TextField
	// MultilineField is several lines of text.
	MultilineField
)

func (k FieldKind) String() string {
	switch k {
	case CheckBox:
		return "check"
	case MultilineField:
		return "multiline"
	}
	return "text"
}

// Field is one "[@…]" form field.
type Field struct {
	Kind FieldKind
	// Name is unique in the document; without "#name" it is
	// "<section>_<n>" after the heading above the field.
	Name string
	// Named reports an explicit "#name".
	Named bool
	// Value is the preset value (text fields), lines joined with "\n".
	Value string
	// Checked: a check box written "[@x]" opens ticked.
	Checked bool
	// Chars is the width in characters of a one-line field.
	Chars int
	// Rows of a multi-line field (the minimum when it fills the page).
	Rows int
	// FillPage: a multi-line field ending in a line "…" grows to the end
	// of the page.
	FillPage bool
	// Pattern such as "HH:MM"; empty for plain text fields.
	Pattern string
	// Tooltip from "<!--* … -->", or the pattern.
	Tooltip string
	// Line and Col (1-based, Col in characters) of the "[@".
	Line, Col int
	// Offset is the byte offset of the "[@" in the source, End the offset
	// after its closing "]".
	Offset, End int
}

// Hint is a "<!--* … -->" field tooltip or a "<!--! … -->" fill hint.
type Hint struct {
	Kind            string // "tooltip" | "fill-hint"
	Line, Col       int
	EndLine, EndCol int
	Offset, End     int
}

// FieldError is a field mlcdoc would reject.
type FieldError struct {
	Line, Col int
	Code      string
	Message   string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("Zeile %d: %s", e.Line, e.Message)
}

// ExtractFields returns every form field of markdown in document order and
// every malformed one. Unlike mlcdoc it does not stop at the first error, so
// an editor can show them all; the first error is the one mlcdoc reports.
func ExtractFields(markdown string) ([]Field, []FieldError) {
	s := &scanner{src: markdown, section: "feld"}
	s.run()
	return s.fields, s.errs
}

type scanner struct {
	src          string
	pos          int
	fields       []Field
	errs         []FieldError
	hints        []Hint
	section      string
	sectionCount int
	inHeading    bool
}

func (s *scanner) rest() string { return s.src[s.pos:] }

func (s *scanner) advance(n int) {
	s.pos += n
	if s.pos > len(s.src) {
		s.pos = len(s.src)
	}
}

func (s *scanner) run() {
	lineStart := true
	for s.pos < len(s.src) {
		rest := s.rest()
		if lineStart {
			if fence := codeFence(rest); fence != "" {
				s.skipFencedBlock(fence)
				continue
			}
			s.inHeading = false
			if title, ok := atxHeading(rest); ok {
				s.inHeading = true
				s.section = HeadingSlug(title)
				if s.section == "" {
					s.section = "feld"
				}
				s.sectionCount = 0
			}
		}
		r, size := utf8.DecodeRuneInString(rest)
		switch {
		case r == '\\':
			// An escaped character (also "\[@") stays as written.
			n := 1
			if len(rest) > 1 {
				_, next := utf8.DecodeRuneInString(rest[1:])
				n += next
			}
			s.advance(n)
			lineStart = false
		case r == '`':
			s.skipCodeSpan()
			lineStart = false
		case r == '<' && strings.HasPrefix(rest, "<!--"):
			n := len(rest)
			if end := strings.Index(rest, "-->"); end >= 0 {
				n = end + 3
			}
			s.hint(s.pos, s.pos+n)
			s.advance(n)
			lineStart = false
		case r == '[' && strings.HasPrefix(rest, "[@"):
			s.advance(s.field())
			lineStart = false
		case r == '\n':
			s.advance(1)
			lineStart = true
		default:
			s.advance(size)
			lineStart = lineStart && unicode.IsSpace(r)
		}
	}
}

func (s *scanner) skipFencedBlock(fence string) {
	first := true
	for s.pos < len(s.src) {
		rest := s.rest()
		lineLen := len(rest)
		if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
			lineLen = nl + 1
		}
		closes := !first && strings.HasPrefix(strings.TrimLeftFunc(rest[:lineLen], unicode.IsSpace), fence)
		s.advance(lineLen)
		first = false
		if closes {
			return
		}
	}
}

func (s *scanner) skipCodeSpan() {
	rest := s.rest()
	ticks := 0
	for ticks < len(rest) && rest[ticks] == '`' {
		ticks++
	}
	if end := strings.Index(rest[ticks:], rest[:ticks]); end >= 0 {
		s.advance(ticks + end + ticks)
		return
	}
	s.advance(ticks)
}

// hint records a "<!--* … -->" tooltip or "<!--! … -->" fill hint.
func (s *scanner) hint(from, to int) {
	kind := ""
	switch {
	case strings.HasPrefix(s.src[from:], "<!--*"):
		kind = "tooltip"
	case strings.HasPrefix(s.src[from:], "<!--!"):
		kind = "fill-hint"
	default:
		return // a plain HTML comment
	}
	h := Hint{Kind: kind, Offset: from, End: to}
	h.Line, h.Col = position(s.src, from)
	h.EndLine, h.EndCol = position(s.src, to)
	s.hints = append(s.hints, h)
}

func (s *scanner) fail(code, msg string) {
	line, col := position(s.src, s.pos)
	s.errs = append(s.errs, FieldError{Line: line, Col: col, Code: code, Message: msg})
}

// field parses a field at "[@" and returns how many bytes it consumed
// (2 when the text is not a field, e.g. a link, or the field is unclosed).
func (s *scanner) field() int {
	rest := s.rest()
	start := s.pos
	if s.inHeading {
		s.fail("field-in-heading", "Formularfelder sind in Überschriften nicht erlaubt (das Inhaltsverzeichnis würde sie wiederholen).")
		return 2
	}
	closeAt := unescaped(rest[2:], ']')
	if closeAt < 0 {
		s.fail("field-unclosed", "`[@` ohne schließendes `]`.")
		return 2
	}
	body := rest[2 : 2+closeAt]
	n := 2 + closeAt + 1
	if strings.HasPrefix(rest[n:], "(") {
		return 2 // "[@team](https://…)" is a link
	}
	if strings.Contains(body, "\n\n") || anyBlank(lines(body)[min(1, len(lines(body))):]) {
		s.fail("field-blank-line", "Das Feld läuft über eine Leerzeile; erst mit `]` schließen.")
		return n
	}
	if unescaped(body, '|') >= 0 {
		s.fail("field-choice", "Auswahllisten (`[@Ja|Nein]`) gibt es noch nicht; `\\|` schreibt ein `|` im Wert.")
		return n
	}

	name, named := "", false
	if after, ok := strings.CutPrefix(body, "#"); ok {
		end := strings.IndexFunc(after, unicode.IsSpace)
		if end < 0 {
			end = len(after)
		}
		id := after[:end]
		if id == "" || strings.IndexFunc(id, func(r rune) bool {
			return !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_')
		}) >= 0 {
			s.fail("field-name", fmt.Sprintf("Feldname %q: nur Buchstaben, Ziffern, `-` und `_` (kein `.`, er trennt Ebenen in PDF-Feldnamen).", id))
			return n
		}
		name, named = id, true
		rest := strings.TrimLeft(after[end:], " \t")
		body = strings.TrimPrefix(rest, "\n")
	}

	markerEnd := start + n
	tooltip := ""
	after := rest[n:]
	gap := len(after) - len(strings.TrimLeft(after, " \t"))
	if comment, ok := strings.CutPrefix(after[gap:], "<!--*"); ok {
		if end := strings.Index(comment, "-->"); end >= 0 {
			tooltip = strings.Join(strings.Fields(comment[:end]), " ")
			s.hint(start+n+gap, start+n+gap+len("<!--*")+end+3)
			n += gap + len("<!--*") + end + 3
		}
	}

	f, code, msg := classify(body)
	if code != "" {
		s.fail(code, msg)
		return n
	}
	if named {
		f.Name, f.Named = name, true
	} else {
		s.sectionCount++
		f.Name = fmt.Sprintf("%s_%d", s.section, s.sectionCount)
	}
	if tooltip != "" {
		f.Tooltip = tooltip
	}
	for _, other := range s.fields {
		if other.Name == f.Name {
			s.fail("field-duplicate", fmt.Sprintf("Feldname `%s` kommt zweimal vor.", f.Name))
			return n
		}
	}
	f.Line, f.Col = position(s.src, start)
	f.Offset, f.End = start, markerEnd
	s.fields = append(s.fields, f)
	return n
}

// classify reads kind, size and preset value from the text between "[@"
// and "]"; a non-empty code reports a malformed field.
func classify(body string) (f Field, code, msg string) {
	f = Field{Kind: TextField, Rows: 1}
	if strings.TrimSpace(body) == "" && !strings.Contains(body, "\n") {
		f.Kind = CheckBox
		return f, "", ""
	}
	// "[@x]": ticked, as on paper and as "- [x]" in a task list (mlcmd2pdf
	// 3.27.0). A one-letter pattern "x" would be a meaningless field.
	if t := strings.TrimSpace(body); (t == "x" || t == "X") && !strings.Contains(body, "\n") {
		f.Kind, f.Checked = CheckBox, true
		return f, "", ""
	}
	if t := strings.TrimSpace(body); t == "…" || t == "..." {
		f.Kind, f.FillPage = MultilineField, true
		return f, "", ""
	}
	ls := lines(body)
	if len(ls) > 1 {
		f.Kind = MultilineField
		if t := strings.TrimSpace(ls[len(ls)-1]); t == "…" || t == "..." {
			ls = ls[:len(ls)-1]
			f.FillPage = true
		}
		f.Rows = max(len(ls), 1)
		values := make([]string, len(ls))
		for i, l := range ls {
			f.Chars = max(f.Chars, utf8.RuneCountInString(unescape(strings.TrimSpace(l))))
			values[i] = preset(l)
		}
		for len(values) > 0 && values[len(values)-1] == "" {
			values = values[:len(values)-1]
		}
		f.Value = strings.Join(values, "\n")
		return f, "", ""
	}
	f.Chars = utf8.RuneCountInString(unescape(strings.TrimSpace(body)))
	// "TT.MM.JJJJ=16\.10\.2026": pattern, then the preset after the first
	// unescaped "=" (a pattern has no "_").
	if eq := unescaped(body, '='); eq >= 0 && !strings.Contains(body[:eq], "_") && strings.TrimSpace(body[:eq]) != "" {
		pattern := unescape(strings.TrimSpace(body[:eq]))
		value := preset(body[eq+1:])
		limit := utf8.RuneCountInString(pattern)
		if utf8.RuneCountInString(value) > limit {
			return f, "field-preset-too-long", fmt.Sprintf("Vorgabewert %q ist länger als sein Muster `%s` (%d Zeichen).", value, pattern, limit)
		}
		f.Chars, f.Value, f.Pattern, f.Tooltip = limit, value, pattern, pattern
		return f, "", ""
	}
	if strings.Contains(body, "_") {
		f.Value = preset(body)
		return f, "", ""
	}
	pattern := unescape(strings.TrimSpace(body))
	if utf8.RuneCountInString(pattern) > 40 {
		return f, "field-not-a-field", fmt.Sprintf("`[@%s]` ist weder eine Linie aus `_` noch ein kurzes Muster wie `[@HH:MM]`.", pattern)
	}
	f.Pattern, f.Tooltip = pattern, pattern
	return f, "", ""
}

// unescaped is the byte index of the first wanted rune not escaped with
// "\", or -1.
func unescaped(text string, wanted rune) int {
	escaped := false
	for i, r := range text {
		switch {
		case escaped:
			escaped = false
		case r == '\\':
			escaped = true
		case r == wanted:
			return i
		}
	}
	return -1
}

// unescape reads "\x" as "x" for ASCII punctuation, as Markdown does; any
// other backslash stays.
func unescape(text string) string {
	var b strings.Builder
	rs := []rune(text)
	for i := 0; i < len(rs); i++ {
		if rs[i] == '\\' && i+1 < len(rs) && isASCIIPunct(rs[i+1]) {
			b.WriteRune(rs[i+1])
			i++
			continue
		}
		b.WriteRune(rs[i])
	}
	return b.String()
}

func isASCIIPunct(r rune) bool {
	return r < 0x80 && (r >= '!' && r <= '/' || r >= ':' && r <= '@' || r >= '[' && r <= '`' || r >= '{' && r <= '~')
}

// preset is the preset value of one marker line: the underscores around it
// (escaped ones count as text) and surrounding spaces removed, escapes
// resolved.
func preset(line string) string {
	text := strings.TrimLeft(strings.TrimSpace(line), "_")
	trimmed := strings.TrimRight(text, "_")
	out := trimmed
	backslashes := len(trimmed) - len(strings.TrimRight(trimmed, `\`))
	if len(trimmed) < len(text) && backslashes%2 == 1 {
		out += "_" // "\_" at the end is an underscore of the value
	}
	return unescape(strings.TrimSpace(out))
}

// codeFence returns the fence ("```" or "~~~" run) when line opens a fenced
// code block.
func codeFence(line string) string {
	indent := len(line) - len(strings.TrimLeft(line, " "))
	if indent > 3 {
		return ""
	}
	t := line[indent:]
	if t == "" || (t[0] != '`' && t[0] != '~') {
		return ""
	}
	run := 0
	for run < len(t) && t[run] == t[0] {
		run++
	}
	if run < 3 {
		return ""
	}
	return t[:run]
}

// atxHeading returns the heading text when line is an ATX heading.
func atxHeading(line string) (string, bool) {
	if nl := strings.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}
	t := strings.TrimLeft(line, " ")
	marks := 0
	for marks < len(t) && t[marks] == '#' {
		marks++
	}
	if marks < 1 || marks > 6 {
		return "", false
	}
	text := t[marks:]
	if text != "" && text[0] != ' ' && text[0] != '\t' {
		return "", false
	}
	text = strings.TrimSpace(text)
	if i := strings.LastIndex(text, "{#"); i >= 0 && strings.HasSuffix(text, "}") {
		text = strings.TrimRight(text[:i], " \t")
	}
	return strings.TrimSpace(strings.TrimRight(text, "#")), true
}

// lines splits like Rust's str::lines: on "\n", without a trailing empty
// line, "\r" removed.
func lines(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\n")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	for i, p := range parts {
		parts[i] = strings.TrimSuffix(p, "\r")
	}
	return parts
}

func anyBlank(ls []string) bool {
	for _, l := range ls {
		if strings.TrimSpace(l) == "" {
			return true
		}
	}
	return false
}

// position converts a byte offset to a 1-based line and column (in
// characters).
func position(src string, offset int) (line, col int) {
	before := src[:offset]
	line = strings.Count(before, "\n") + 1
	lineStart := strings.LastIndexByte(before, '\n') + 1
	return line, utf8.RuneCountInString(before[lineStart:]) + 1
}

// HeadingSlug turns heading text into an anchor exactly like GitHub and
// mlcmd2pdf (heading_slug): lower case, space becomes "-", letters
// (umlauts included), digits, "-" and "_" stay, everything else is dropped.
// "1. Ziel & Zweck (v2)" gives "1-ziel--zweck-v2".
func HeadingSlug(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}
