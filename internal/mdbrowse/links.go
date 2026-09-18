package mdbrowse

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Link validation + rewriting for rendered markdown HTML.
//
// Every internal link/image target is resolved against the current document
// and checked BEFORE the user can click it:
//   - missing files and paths escaping the root become non-clickable
//     placeholders (class "broken-link" / "broken-img") — no more 404s
//   - existing .md files get a data-md attribute with their root-relative
//     path so the Svelte client can open them in-app via /render
//   - existing non-markdown files (images, PDFs, …) are rewritten to /raw
//   - external URLs, anchors and data: URIs pass through untouched

var (
	aTagRe   = regexp.MustCompile(`<a\b([^>]*)>`)
	imgTagRe = regexp.MustCompile(`<img\b([^>]*)>`)
	hrefRe   = regexp.MustCompile(`href="([^"]*)"`)
	srcRe    = regexp.MustCompile(`src="([^"]*)"`)
	altRe    = regexp.MustCompile(`alt="([^"]*)"`)
	schemeRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)
)

// linkKind classifies an href/src value.
type linkKind int

const (
	linkKeep     linkKind = iota // external URL, anchor, empty, data: URI — pass through
	linkInternal                 // root-relative resolvable target
)

// classifyTarget splits off fragment/query and decides whether the target
// is internal (to be validated) or passes through untouched.
func classifyTarget(href string) (linkKind, string) {
	if href == "" {
		return linkKeep, ""
	}
	if strings.HasPrefix(href, "#") {
		return linkKeep, ""
	}
	if strings.HasPrefix(href, "//") || schemeRe.MatchString(href) {
		return linkKeep, ""
	}
	target := href
	if i := strings.IndexByte(target, '#'); i != -1 {
		target = target[:i]
	}
	if i := strings.IndexByte(target, '?'); i != -1 {
		target = target[:i]
	}
	if target == "" {
		return linkKeep, ""
	}
	return linkInternal, target
}

// resolveDocTarget resolves a link target against the document's directory.
// Returns the root-relative slash path, or "" if it escapes the root.
func resolveDocTarget(docRelPath, target string) string {
	if path.IsAbs(target) {
		target = strings.TrimPrefix(target, "/")
	}
	resolved := path.Join(path.Dir(docRelPath), target)
	if resolved == ".." || strings.HasPrefix(resolved, "../") {
		return ""
	}
	return resolved
}

// validateTarget checks existence within the root.
// Directories resolve to their index.md. Returns root-relative path + ok.
func (r *Root) validateTarget(resolved string) (string, bool) {
	abs, err := r.Resolve(resolved)
	if err != nil {
		return "", false
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", false
	}
	if info.IsDir() {
		resolved = path.Join(resolved, "index.md")
		abs, err = r.Resolve(resolved)
		if err != nil {
			return "", false
		}
		if _, err := os.Stat(abs); err != nil {
			return "", false
		}
	}
	return resolved, true
}

// stripAttr removes an attribute (e.g. href) from a raw attribute string.
func stripAttr(attrs, name string) string {
	re := regexp.MustCompile(`\s*` + name + `="[^"]*"`)
	return re.ReplaceAllString(attrs, "")
}

// addClass appends a CSS class to a raw attribute string.
func addClass(attrs, class string) string {
	m := regexp.MustCompile(`class="([^"]*)"`).FindStringSubmatch(attrs)
	if m == nil {
		return attrs + ` class="` + class + `"`
	}
	return strings.Replace(attrs, m[0], `class="`+m[1]+` `+class+`"`, 1)
}

// rewriteMarkdownLinks validates + rewrites <a href> and <img src> in
// rendered HTML. docRelPath is the root-relative path of the document.
func (r *Root) rewriteMarkdownLinks(html, docRelPath string) string {
	docRelPath = filepath.ToSlash(docRelPath)

	html = aTagRe.ReplaceAllStringFunc(html, func(tag string) string {
		attrs := aTagRe.FindStringSubmatch(tag)[1]
		m := hrefRe.FindStringSubmatch(attrs)
		if m == nil {
			return tag // no href — leave alone
		}
		href := m[1]
		kind, target := classifyTarget(href)
		if kind == linkKeep {
			return tag
		}
		resolved := resolveDocTarget(docRelPath, target)
		if resolved == "" {
			return brokenLinkTag(attrs, href, "Ziel außerhalb des Verzeichnisses")
		}
		resolved, ok := r.validateTarget(resolved)
		if !ok {
			return brokenLinkTag(attrs, href, "Datei nicht gefunden: "+target)
		}
		if strings.EqualFold(path.Ext(resolved), ".md") {
			// In-app navigation: client intercepts a[data-md] via /render.
			return `<a` + attrs + ` data-md="` + escapeHTML(resolved) + `">`
		}
		// Existing non-markdown file: serve via /raw (works with/without JS).
		newAttrs := hrefRe.ReplaceAllString(attrs, `href="/raw?path=`+url.QueryEscape(resolved)+`"`)
		return `<a` + newAttrs + `>`
	})

	html = imgTagRe.ReplaceAllStringFunc(html, func(tag string) string {
		attrs := imgTagRe.FindStringSubmatch(tag)[1]
		m := srcRe.FindStringSubmatch(attrs)
		if m == nil {
			return tag
		}
		src := m[1]
		kind, target := classifyTarget(src)
		if kind == linkKeep {
			return tag
		}
		resolved := resolveDocTarget(docRelPath, target)
		if resolved == "" {
			return brokenImgTag(attrs, src, "Bild außerhalb des Verzeichnisses")
		}
		resolved, ok := r.validateTarget(resolved)
		if !ok {
			return brokenImgTag(attrs, src, "Bild nicht gefunden: "+target)
		}
		newAttrs := srcRe.ReplaceAllString(attrs, `src="/raw?path=`+url.QueryEscape(resolved)+`"`)
		return `<img` + newAttrs + `>`
	})

	return html
}

// brokenLinkTag turns a link into a non-clickable placeholder.
func brokenLinkTag(attrs, href, reason string) string {
	attrs = stripAttr(attrs, "href")
	attrs = addClass(attrs, "broken-link")
	return `<a` + attrs + ` aria-disabled="true" title="` + escapeHTML(reason) + `" data-broken="` + escapeHTML(href) + `">`
}

// brokenImgTag replaces an unresolvable image with a placeholder.
func brokenImgTag(attrs, src, reason string) string {
	alt := src
	if m := altRe.FindStringSubmatch(attrs); m != nil && m[1] != "" {
		alt = m[1]
	}
	return `<span class="broken-img" title="` + escapeHTML(reason) + `">Bild nicht gefunden: ` + escapeHTML(alt) + `</span>`
}
