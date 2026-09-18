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

// isWebExternal reports http(s) links (and protocol-relative ones).
// mailto:, tel:, data: and friends keep native behavior and need no
// confirmation or new window.
func isWebExternal(href string) bool {
	lh := strings.ToLower(href)
	return strings.HasPrefix(lh, "http://") ||
		strings.HasPrefix(lh, "https://") ||
		strings.HasPrefix(lh, "//")
}

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

// validateTarget checks existence within the root and reports whether
// the target is a directory. Directories stay directories: /render
// serves their index.md or a virtual listing, so no silent remapping.
func (r *Root) validateTarget(resolved string) (target string, isDir bool, ok bool) {
	abs, err := r.Resolve(resolved)
	if err != nil {
		return "", false, false
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", false, false
	}
	return resolved, info.IsDir(), true
}

// ensureRelTokens adds tokens to the rel attribute, creating it if absent.
func ensureRelTokens(attrs string, tokens ...string) string {
	m := regexp.MustCompile(`rel="([^"]*)"`).FindStringSubmatch(attrs)
	if m == nil {
		return attrs + ` rel="` + strings.Join(tokens, " ") + `"`
	}
	have := strings.Fields(m[1])
	for _, tok := range tokens {
		found := false
		for _, h := range have {
			if h == tok {
				found = true
				break
			}
		}
		if !found {
			have = append(have, tok)
		}
	}
	return strings.Replace(attrs, m[0], `rel="`+strings.Join(have, " ")+`"`, 1)
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
		if isWebExternal(href) {
			// External links: always recognizable, always a new window.
			// The client additionally asks for confirmation (see
			// ExternalLinkModal); the markup alone already does the
			// safe thing without JS.
			attrs = addClass(attrs, "external-link")
			if !strings.Contains(attrs, "target=") {
				attrs += ` target="_blank"`
			}
			// Target _blank without noopener hands window.opener to the
			// foreign page (reverse tabnabbing). Goldmark already emits
			// rel="nofollow", so merge instead of duplicating the attr.
			attrs = ensureRelTokens(attrs, "noopener", "noreferrer")
			return `<a` + attrs + `>`
		}
		kind, target := classifyTarget(href)
		if kind == linkKeep {
			return tag
		}
		resolved := resolveDocTarget(docRelPath, target)
		if resolved == "" {
			return brokenLinkTag(attrs, href, "Ziel außerhalb des Verzeichnisses")
		}
		resolved, isDir, ok := r.validateTarget(resolved)
		if !ok {
			return brokenLinkTag(attrs, href, "Datei nicht gefunden: "+target)
		}
		if isDir || strings.EqualFold(path.Ext(resolved), ".md") {
			// In-app navigation: client intercepts a[data-md] via /render,
			// which serves index.md or a virtual listing for directories.
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
		resolved, isDir, ok := r.validateTarget(resolved)
		if !ok || isDir {
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
