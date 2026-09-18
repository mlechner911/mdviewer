package mdbrowse

import (
	"regexp"
	"strconv"
	"strings"
)

// normalizePastedMarkdown repairs the most common defects of pasted or
// AI-generated markdown where newlines got eaten ("glued" block markers):
//   - "Backups## 1. Titel"      -> heading on its own line
//   - "text ## Heading"         -> heading on its own line
//   - "---## 2. Titel"          -> thematic break + heading
//   - unclosed fenced block at EOF -> auto-closed
//
// Deliberately NOT touched (too ambiguous / spec-meaningful):
//   - glued list markers ("a* b" could be multiplication)
//   - setext underlines ("text\n---" is a valid H2 per CommonMark)
//   - anything inside fenced code blocks or inline code spans
//   - single "#" without leading space ("C# rocks" must survive)
//
// The function is idempotent for well-formed documents: every rule only
// fires when a marker is glued to other text.

var (
	hrGlueRe     = regexp.MustCompile(`^( {0,3})(-{3,}|\*{3,}|_{3,})(.+)$`)
	gluedHashes  = regexp.MustCompile(`(\S)(#{2,6})( )`)
	spacedHashes = regexp.MustCompile("([^\\S\\n])(#{1,6})( )")
	// RE2 has no backreferences: match symmetric runs greedily instead.
	// Unmatched lone backticks simply don't match and stay as-is.
	inlineCodeSpan = regexp.MustCompile("`+[^`\n]*`+")
)

// normalizePastedMarkdown normalizes glued block markers line by line.
func normalizePastedMarkdown(body string) string {
	lines := strings.Split(body, "\n")
	out := make([]string, 0, len(lines))

	inFence := false
	fenceMark := ""

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)

		// Fenced code blocks pass through verbatim.
		if indent < 4 && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			mark := trimmed[:3]
			if !inFence {
				inFence = true
				fenceMark = mark
			} else if mark == fenceMark {
				inFence = false
				fenceMark = ""
			}
			out = append(out, line)
			continue
		}
		if inFence {
			out = append(out, line)
			continue
		}
		// Indented code blocks pass through verbatim.
		if indent >= 4 {
			out = append(out, line)
			continue
		}
		out = append(out, normalizeProseLine(line))
	}

	s := strings.Join(out, "\n")
	if inFence {
		// Truncated paste: close the fence so the rest of the
		// document does not disappear into a code block.
		if fenceMark == "" {
			fenceMark = "```"
		}
		s += "\n" + fenceMark + "\n"
	}
	return s
}

// placeholder wraps an index in NUL bytes (cannot collide with markdown).
func placeholder(i int) string {
	return "\x00" + strconv.Itoa(i) + "\x00"
}

// normalizeProseLine splits glued hr / heading markers outside code spans.
func normalizeProseLine(line string) string {
	// Protect inline code spans via placeholders.
	var spans []string
	line = inlineCodeSpan.ReplaceAllStringFunc(line, func(m string) string {
		spans = append(spans, m)
		return placeholder(len(spans) - 1)
	})
	restore := func(s string) string {
		for i, span := range spans {
			s = strings.ReplaceAll(s, placeholder(i), span)
		}
		return s
	}

	// 1. Thematic break glued to text at line start ("---## 2. T").
	if m := hrGlueRe.FindStringSubmatch(line); m != nil {
		rest := m[3]
		if len(rest) > 0 && rest[0] != ' ' && rest[0] != '\t' {
			// Guard "***bold*** stays.": if the same run reappears
			// later in the line, this is emphasis, not a rule.
			run := m[2]
			isEmphasis := (run[0] == '*' || run[0] == '_') &&
				strings.Contains(rest, run)
			if !isEmphasis {
				line = m[1] + run + "\n" + m[1] + rest
			}
		}
	}

	// 2. Glued headings ("Backups## 1. T", "text ## T", "text # T").
	parts := strings.Split(line, "\n")
	for i, p := range parts {
		p = gluedHashes.ReplaceAllString(p, "$1\n$2$3")
		p = spacedHashes.ReplaceAllString(p, "\n$2$3")
		parts[i] = p
	}
	line = strings.Join(parts, "\n")

	return restore(line)
}
