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

// PowerShell's lexer marks the ':' in bare URLs as Error tokens; the style CSS
// must not render those as red boxes.
func TestErrorTokensAreNeutralised(t *testing.T) {
	r := NewRenderer()

	out, err := r.Render("```powershell\ncurl.exe http://localhost:3128/stat\n```", "github")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(out, `class="err"`) {
		t.Fatalf("expected the PowerShell sample to produce Error tokens, got:\n%s", out)
	}

	for _, style := range []string{"github", "github-dark"} {
		css, err := r.GetStyleCSS(style)
		if err != nil {
			t.Fatalf("GetStyleCSS(%q) error: %v", style, err)
		}
		override := strings.LastIndex(css, errorTokenOverrideCSS)
		if override < 0 {
			t.Fatalf("GetStyleCSS(%q) lacks the Error token override", style)
		}
		if styleRule := strings.LastIndex(css, "/* Error */"); styleRule > override {
			t.Errorf("GetStyleCSS(%q): style's Error rule comes after the override", style)
		}
	}
}

func TestCodeBlocks(t *testing.T) {
	r := NewRenderer()

	tests := []struct {
		name  string
		input string
	}{
		{
			name: "Fenced code block without lang",
			input: "```\nhello world\n```\n\n**bold after**",
		},
		{
			name: "Fenced code block with lang",
			input: "```go\nfunc main() {}\n```\n\n*italic after*",
		},
		{
			name: "Unclosed code block",
			input: "```\nhello unclosed",
		},
		{
			name: "Inline code then bold",
			input: "`inline code` and **bold**",
		},
		{
			name: "Code block immediately followed by heading without blank line",
			input: "```\ncode\n```\n# Heading immediately after\n**bold immediately after**",
		},
		{
			name: "Code block immediately followed by text with format end",
			input: "```\ncode\n```\nSome formatted text that ends with something",
		},
		{
			name: "Mermaid block",
			input: "```mermaid\ngraph TD\nA --> B\n```",
		},
		{
			name: "Single line fence with bold",
			input: "``` **bold text** ```",
		},
		{
			name: "Code fence with markdown inside",
			input: "```\n**bold text**\n```",
		},
		{
			name: "Code fence closed with word end",
			input: "```\n**bold text**\nend\n\nnormal text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := r.Render(tt.input, "github")
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			t.Logf("[%s] Output:\n%s\n", tt.name, out)
		})
	}
}
