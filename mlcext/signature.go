package mlcext

import (
	"regexp"
	"strconv"
	"strings"
)

// Signature is the seal the Sensor Monitor puts into a released Markdown
// document. It is not checked here (only the server holds the keys); this
// package finds it, so that every tool can leave it out of the content.
//
// Two forms exist:
//
//   - "end" (current): an HTML comment as the last thing in the file,
//
//     <!--mlc-signature
//     doc_no: SOP-2026-0013
//     signed_by: "Dr. Anna Berger"
//     …
//     -->
//
//     The signed content is everything before "<!--mlc-signature", byte for
//     byte; the comment starts on a line of its own.
//
//   - "front" (until 2026-10): a front matter block holding only
//     "signature:" before the document's own front matter. The signed
//     content is everything after the block, as in the Sensor Monitor's
//     document.SplitSignatureBlock. Signed documents are records and are not
//     rewritten, so this form stays readable.
type Signature struct {
	// Form is "end" or "front".
	Form string `json:"form"`
	// Fields of the block: doc_no, meaning, signed_by, signed_at, sha256,
	// alg, key_id, sig. Quoted values are unquoted.
	Fields map[string]string `json:"fields"`
	// Line and EndLine: the 1-based lines the block takes in the file.
	Line    int `json:"line"`
	EndLine int `json:"endLine"`
}

// SignedBy, SignedAt and Meaning are the commonly shown fields.
func (s Signature) SignedBy() string { return s.Fields["signed_by"] }
func (s Signature) SignedAt() string { return s.Fields["signed_at"] }
func (s Signature) Meaning() string  { return s.Fields["meaning"] }

const signatureStart = "<!--mlc-signature"

var (
	reSigEnd   = regexp.MustCompile(`(?s)<!--mlc-signature[ \t]*\r?\n(.*?)-->\s*\z`)
	reSigFront = regexp.MustCompile(`(?s)\A---\r?\n(.*?)\r?\n---\r?\n`)
)

// SplitSignature returns the signature of md (nil when it has none) and the
// signed content: md without the block, exactly the bytes the checksum
// covers. Unsigned documents come back unchanged.
func SplitSignature(md string) (*Signature, string) {
	if m := reSigEnd.FindStringSubmatchIndex(md); m != nil && (m[0] == 0 || md[m[0]-1] == '\n') {
		if fields, ok := signatureFields(md[m[2]:m[3]], false); ok {
			return &Signature{
				Form: "end", Fields: fields,
				Line:    strings.Count(md[:m[0]], "\n") + 1,
				EndLine: strings.Count(strings.TrimRight(md[:m[1]], " \t\r\n"), "\n") + 1,
			}, md[:m[0]]
		}
	}
	if m := reSigFront.FindStringSubmatchIndex(md); m != nil {
		if fields, ok := signatureFields(md[m[2]:m[3]], true); ok {
			return &Signature{
				Form: "front", Fields: fields,
				Line: 1, EndLine: strings.Count(md[:m[1]], "\n"),
			}, md[m[1]:]
		}
	}
	return nil, md
}

// signatureFields reads "key: value" lines. The front form must hold only
// "signature:" with indented lines below it (anything else would not be
// signed and could change the rendering, so it is no signature: fail closed,
// as in the Sensor Monitor).
func signatureFields(block string, front bool) (map[string]string, bool) {
	ls := strings.Split(strings.ReplaceAll(block, "\r\n", "\n"), "\n")
	if front {
		if strings.TrimSpace(ls[0]) != "signature:" {
			return nil, false
		}
		ls = ls[1:]
	}
	fields := map[string]string{}
	for _, l := range ls {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if front && !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
			return nil, false
		}
		k, v, ok := strings.Cut(strings.TrimSpace(l), ":")
		if !ok || strings.TrimSpace(k) == "" {
			return nil, false
		}
		v = strings.TrimSpace(v)
		if uq, err := strconv.Unquote(v); err == nil {
			v = uq
		}
		fields[strings.TrimSpace(k)] = v
	}
	return fields, len(fields) > 0
}

// skipFrontSignature returns the length of a leading signature block of the
// front form, 0 without one.
func skipFrontSignature(md string) int {
	if m := reSigFront.FindStringSubmatchIndex(md); m != nil {
		if _, ok := signatureFields(md[m[2]:m[3]], true); ok {
			return m[1]
		}
	}
	return 0
}

// endSignatureAt is the offset of a trailing signature comment, len(md)
// without one.
func endSignatureAt(md string) int {
	if m := reSigEnd.FindStringSubmatchIndex(md); m != nil && (m[0] == 0 || md[m[0]-1] == '\n') {
		if _, ok := signatureFields(md[m[2]:m[3]], false); ok {
			return m[0]
		}
	}
	return len(md)
}
