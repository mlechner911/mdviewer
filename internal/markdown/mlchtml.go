package markdown

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// The MLC document dialect (mlcdoc, MARKDOWN-PDF-LAYOUT.md) uses exactly
// two pieces of raw HTML: a page break block and empty anchor targets.
// Goldmark drops all raw HTML in safe mode; this renderer lets these two
// forms through verbatim and keeps dropping everything else. bluemonday
// still sanitizes the result afterwards.
var (
	pageBreakRe  = regexp.MustCompile(`^\s*<div class="page-break">\s*</div>\s*$`)
	anchorOpenRe = regexp.MustCompile(`^<a id="[\p{L}\p{N}_-]+">$`)
	anchorFullRe = regexp.MustCompile(`^\s*<a id="[\p{L}\p{N}_-]+">\s*</a>\s*$`)
)

// mlcHTMLRenderer renders allowlisted raw HTML (block and inline).
type mlcHTMLRenderer struct{}

func (r *mlcHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
}

func (r *mlcHTMLRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.HTMLBlock)
	var buf bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		buf.Write(line.Value(source))
	}
	if n.HasClosure() {
		buf.Write(n.ClosureLine.Value(source))
	}
	if raw := buf.Bytes(); pageBreakRe.Match(raw) || anchorFullRe.Match(raw) {
		_, _ = w.Write(bytes.TrimSpace(raw))
		_ = w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}

func (r *mlcHTMLRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	n := node.(*ast.RawHTML)
	var buf bytes.Buffer
	for i := 0; i < n.Segments.Len(); i++ {
		seg := n.Segments.At(i)
		buf.Write(seg.Value(source))
	}
	if raw := buf.Bytes(); anchorOpenRe.Match(raw) || string(raw) == "</a>" {
		_, _ = w.Write(raw)
	}
	return ast.WalkSkipChildren, nil
}
