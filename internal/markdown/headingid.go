package markdown

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// HeadingSlug turns heading text into an anchor exactly like GitHub and
// the MLC PDF toolkit (mlcmd2pdf heading_slug): lower case, space → "-", keep
// letters (umlauts included), digits, "-" and "_", drop the rest.
// "1. Ziel & Zweck (v2)" → "1-ziel--zweck-v2". Same anchors in the
// preview and in the PDF, so in-document links work in both.
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

// headingIDs implements parser.IDs with HeadingSlug; duplicates get
// "-1", "-2" … like mlcdoc. Explicit {#id} values are registered via Put.
type headingIDs struct {
	used map[string]bool
}

func newHeadingIDs() parser.IDs {
	return &headingIDs{used: map[string]bool{}}
}

func (s *headingIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	base := HeadingSlug(string(value))
	if base == "" {
		base = "heading"
	}
	id := base
	for i := 1; s.used[id]; i++ {
		id = base + "-" + strconv.Itoa(i)
	}
	s.used[id] = true
	return []byte(id)
}

func (s *headingIDs) Put(value []byte) {
	s.used[string(value)] = true
}
