package markdown

import (
	"strings"
	"testing"
)

func TestFrontMatterRendering(t *testing.T) {
	r := NewRenderer()

	input := `---
title: Test Document
author: Michael
tags:
  - markdown
  - viewer
draft: false
---

# Main Heading

This is the document content.
`
	html, err := r.Render(input, "github-dark")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	t.Logf("Rendered HTML:\n%s", html)

	if !strings.Contains(html, "Main Heading") {
		t.Errorf("Expected HTML to contain 'Main Heading', got:\n%s", html)
	}
	// Check that raw yaml markers like <hr> are not shown as broken text
	if strings.Contains(html, "<hr>") && !strings.Contains(html, "frontmatter") {
		t.Errorf("Raw front matter was rendered as horizontal rule: %s", html)
	}
}
