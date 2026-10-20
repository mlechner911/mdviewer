package markdown

import (
	"regexp"
	"strings"
)

// goldmark-mathjax renders a display formula written on one line
// ("$$x^2$$") as an empty \[\]; only the multi-line form keeps its content.
// splitOneLineDisplayMath rewrites such lines into the multi-line form,
// outside fenced code blocks.
var oneLineDisplayMath = regexp.MustCompile(`^(\s*)\$\$(.+?)\$\$\s*$`)

func splitOneLineDisplayMath(md string) string {
	if !strings.Contains(md, "$$") {
		return md
	}
	lines := strings.Split(md, "\n")
	fence := ""
	for i, l := range lines {
		t := strings.TrimLeft(l, " ")
		switch {
		case fence != "":
			if strings.HasPrefix(t, fence) {
				fence = ""
			}
			continue
		case strings.HasPrefix(t, "```"):
			fence = "```"
			continue
		case strings.HasPrefix(t, "~~~"):
			fence = "~~~"
			continue
		}
		if m := oneLineDisplayMath.FindStringSubmatch(l); m != nil && strings.TrimSpace(m[2]) != "" {
			lines[i] = m[1] + "$$\n" + m[1] + strings.TrimSpace(m[2]) + "\n" + m[1] + "$$"
		}
	}
	return strings.Join(lines, "\n")
}
