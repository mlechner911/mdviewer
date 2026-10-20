package mlcext

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// MoveHeaderToFrontMatter is the quick fix for "header-in-body": id,
// version and title from the legacy header lines ("**Dokumenten-ID:** … |
// **Version:** …", "**Titel:** …") move into the front matter as doc_id,
// version and title; the lines disappear from the text (a line keeps what
// else it holds). Values already in the front matter win and are not
// overwritten. A heading that repeats the title stays: removing it is the
// author's call. ok is false when there is nothing to move.
func MoveHeaderToFrontMatter(md string) (out string, ok bool) {
	h := ParseHeader(md)
	if h.refLine == 0 && h.versionLine == 0 && h.titleLine == 0 {
		return md, false
	}
	fm, _ := ParseFrontMatter(md)
	var add []string
	if fm.DocID == "" && h.RefNo != "" {
		add = append(add, "doc_id: "+yamlScalar(h.RefNo))
	}
	if strings.TrimSpace(fm.Version) == "" && h.Version != "" {
		add = append(add, "version: "+yamlScalar(h.Version))
	}
	if fm.Title == "" && h.Title != "" {
		add = append(add, "title: "+yamlScalar(h.Title))
	}

	// Remove the header fields from their lines (1-based file lines).
	ls := strings.Split(md, "\n")
	drop := map[int]bool{}
	for _, n := range []int{h.refLine, h.versionLine, h.titleLine} {
		if n < 1 || n > len(ls) {
			continue
		}
		l := reHeaderField.ReplaceAllString(ls[n-1], "")
		if strings.Trim(l, " \t|\\\r") == "" {
			drop[n] = true
		} else {
			ls[n-1] = strings.TrimRight(strings.TrimLeft(l, " |"), " |")
		}
	}
	kept := make([]string, 0, len(ls))
	for i, l := range ls {
		if !drop[i+1] {
			kept = append(kept, l)
		}
	}
	text := strings.Join(kept, "\n")

	// Insert into the front matter, or start one (after a signature block
	// of the old form).
	if len(add) > 0 {
		b := frontMatterBlock(text)
		if b.present {
			at := lineOffset(text, b.yamlLine-1) // start of the YAML's first line
			text = text[:at] + strings.Join(add, "\n") + "\n" + text[at:]
		} else {
			at := skipFrontSignature(text)
			rest := strings.TrimLeft(text[at:], "\r\n")
			text = text[:at] + "---\n" + strings.Join(add, "\n") + "\n---\n" + rest
		}
	}
	return text, true
}

// reHeaderField is one "**Feld:** Wert" of a header line, with a following
// separator "|".
var reHeaderField = regexp.MustCompile(`\*\*\s*(?:Titel|Dokumenten-ID|Dokument-ID|Version)\s*:\s*\*\*\s*[^|\n*]*(?:\|\s*)?`)

// lineOffset is the byte offset of the start of 0-based line n.
func lineOffset(s string, n int) int {
	off := 0
	for i := 0; i < n; i++ {
		j := strings.IndexByte(s[off:], '\n')
		if j < 0 {
			return len(s)
		}
		off += j + 1
	}
	return off
}

// SignatureState tells whether md is signed and whether its content still
// matches the signed checksum (sha256 of the content, see SplitSignature).
// The signature itself is not verified (only the server holds the keys); a
// mismatch is enough to know that an edit voided it.
type SignatureState struct {
	Signed   bool   `json:"signed"`
	Intact   bool   `json:"intact"`
	SignedBy string `json:"signedBy"`
	SignedAt string `json:"signedAt"`
}

// CheckSignature returns the SignatureState of md.
func CheckSignature(md string) SignatureState {
	sig, rest := SplitSignature(md)
	if sig == nil {
		return SignatureState{}
	}
	sum := sha256.Sum256([]byte(rest))
	return SignatureState{
		Signed:   true,
		Intact:   strings.EqualFold(hex.EncodeToString(sum[:]), strings.TrimSpace(sig.Fields["sha256"])),
		SignedBy: sig.SignedBy(),
		SignedAt: sig.SignedAt(),
	}
}

// RemoveSignature returns md without its signature block: what an edited
// signed document is saved as (a new version, to be signed again).
func RemoveSignature(md string) string {
	_, rest := SplitSignature(md)
	return rest
}
