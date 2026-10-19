package markdown

import (
	"strconv"

	"github.com/mlechner911/mdviewer/mlcext"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// HeadingSlug turns heading text into an anchor exactly like GitHub and
// the MLC PDF toolkit; the rule lives in mlcext (single source).
func HeadingSlug(text string) string { return mlcext.HeadingSlug(text) }

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
