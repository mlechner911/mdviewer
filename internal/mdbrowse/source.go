package mdbrowse

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// DocSource is a document as Markdown source, for clients that render
// themselves (<MlcMarkdown>, mlc-markdown). Path is the file actually
// read: a directory resolves to its index.md.
type DocSource struct {
	Path     string `json:"path"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// Source reads relPath as Markdown source. A directory yields its
// index.md; ok is false for a directory without one (the caller falls
// back to RenderDir's virtual listing). Pasted-markdown defects are
// repaired like in RenderDoc, the front matter is kept.
func (r *Root) Source(relPath string) (doc DocSource, ok bool, err error) {
	rel := filepath.ToSlash(relPath)
	abs, err := r.Resolve(rel)
	if err != nil {
		return DocSource{}, false, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return DocSource{}, false, err
	}
	if info.IsDir() {
		if !fileExists(abs, "index.md") {
			return DocSource{}, false, nil
		}
		rel = path.Join(rel, "index.md")
		if abs, err = r.Resolve(rel); err != nil {
			return DocSource{}, false, err
		}
		if info, err = os.Stat(abs); err != nil {
			return DocSource{}, false, err
		}
	} else if !strings.EqualFold(path.Ext(rel), ".md") {
		return DocSource{}, false, os.ErrInvalid
	}
	content, _, err := r.ReadDoc(rel)
	if err != nil {
		return DocSource{}, false, err
	}
	title, _ := ExtractTitle(abs)
	fm, body := splitFrontmatter(content)
	return DocSource{
		Path:     rel,
		Title:    title,
		Source:   fm + normalizePastedMarkdown(body),
		Size:     info.Size(),
		Modified: info.ModTime().UTC().Format(time.RFC3339),
	}, true, nil
}

// splitFrontmatter splits content into the front matter block (with its
// delimiters and trailing newlines) and the body, like stripFrontmatter.
func splitFrontmatter(content string) (fm, body string) {
	body = stripFrontmatter(content)
	return content[:len(content)-len(body)], body
}

// LinkTarget is how a client should treat one link or image target of a
// document: the same decisions rewriteMarkdownLinks makes for RenderDoc.
type LinkTarget struct {
	// Kind: "md" (in-app navigation, also directories), "file" (download
	// or view via the raw handler), "broken", or "keep" (external, anchor,
	// data: — the client leaves it alone).
	Kind string `json:"kind"`
	// Path: root-relative target for "md" and "file".
	Path string `json:"path,omitempty"`
	// Image: the target is an image browsers display natively.
	Image  bool   `json:"image,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// ResolveLinks classifies the link and image targets of docRelPath
// (hrefs as written in the source; fragment and query are ignored).
func (r *Root) ResolveLinks(docRelPath string, hrefs []string) map[string]LinkTarget {
	docRelPath = filepath.ToSlash(docRelPath)
	out := make(map[string]LinkTarget, len(hrefs))
	for _, href := range hrefs {
		if _, done := out[href]; done {
			continue
		}
		out[href] = r.resolveLink(docRelPath, href)
	}
	return out
}

func (r *Root) resolveLink(docRelPath, href string) LinkTarget {
	if isWebExternal(href) {
		return LinkTarget{Kind: "keep"}
	}
	kind, target := classifyTarget(href)
	if kind == linkKeep {
		return LinkTarget{Kind: "keep"}
	}
	resolved := resolveDocTarget(docRelPath, target)
	if resolved == "" {
		return LinkTarget{Kind: "broken", Reason: "Ziel außerhalb des Verzeichnisses"}
	}
	resolved, isDir, ok := r.validateTarget(resolved)
	if !ok {
		return LinkTarget{Kind: "broken", Reason: "Datei nicht gefunden: " + target}
	}
	if isDir || strings.EqualFold(path.Ext(resolved), ".md") {
		return LinkTarget{Kind: "md", Path: resolved}
	}
	return LinkTarget{Kind: "file", Path: resolved, Image: isViewableImage(resolved)}
}
