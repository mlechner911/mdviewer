package mlcext

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

const sigEnd = "<!--mlc-signature\ndoc_no: SOP-2026-0013\nmeaning: approved\nsigned_by: \"Dr. Anna Berger\"\nsigned_at: 2026-10-19T14:02:11Z\nsha256: abc\nalg: ed25519\nkey_id: k1\nsig: s1\n-->\n"

const sigFront = "---\nsignature:\n  doc_no: SOP-2026-0013\n  signed_by: \"Dr. Anna Berger\"\n  signed_at: 2026-10-19T14:02:11Z\n  meaning: approved\n---\n"

func allCodes(ds []Diagnostic) []string {
	out := []string{}
	for _, d := range ds {
		out = append(out, string(d.Severity)+":"+d.Code)
	}
	return out
}

func hasCode(ds []Diagnostic, sev Severity, code string) *Diagnostic {
	for i := range ds {
		if ds[i].Severity == sev && ds[i].Code == code {
			return &ds[i]
		}
	}
	return nil
}

func TestDocumentMetaFrontMatterFirst(t *testing.T) {
	md := "---\ndoc_id: INFO-QM-001\nversion: 1.0\ntitle: test info\n---\n# Hallo\n"
	m := DocumentMeta(md)
	if m.DocID != "INFO-QM-001" || m.Version != "1.0" || m.Title != "test info" || m.Kind != "info" || m.DocIDSource != "frontmatter" {
		t.Errorf("meta %+v", m)
	}
	// Legacy header fills in what the front matter leaves open.
	legacy := "---\ntitle: t\n---\n**Dokumenten-ID:** SOP-QM-002 | **Version:** 2.1\n"
	if m := DocumentMeta(legacy); m.DocID != "SOP-QM-002" || m.Version != "2.1" || m.DocIDSource != "header" || m.Kind != "sop" {
		t.Errorf("legacy meta %+v", m)
	}
	if !Detect("---\ndoc_id: FB-QM-001\n---\nText\n") {
		t.Error("doc_id in the front matter must count as an MLC document")
	}
}

func TestCheckFrontMatterAsOnlySource(t *testing.T) {
	clean := "---\ndoc_id: INFO-QM-001\nversion: \"1.0\"\ntitle: test info\nevent_kinds: [DEVIATION]\n---\n> [!NOTE]\n> Das Wichtigste.\n"
	for _, d := range Check(clean, Options{RequireHeader: true}) {
		if d.Severity != Info {
			t.Errorf("clean document: %s %s", d.Code, d.Message)
		}
	}

	legacy := "---\ntitle: t\n---\n**Dokumenten-ID:** SOP-QM-002 | **Version:** 2.1\n"
	if d := hasCode(Check(legacy, Options{RequireHeader: true}), Warning, "header-in-body"); d == nil || d.Line != 4 {
		t.Errorf("legacy header: want header-in-body warning at line 4, got %v", allCodes(Check(legacy, Options{RequireHeader: true})))
	}

	conflict := "---\ndoc_id: SOP-QM-001\nversion: \"1.0\"\n---\n**Dokumenten-ID:** SOP-QM-009 | **Version:** 1.0\n"
	ds := Check(conflict, Options{RequireHeader: true})
	if d := hasCode(ds, Error, "header-conflict"); d == nil || !strings.Contains(d.Message, "SOP-QM-009") {
		t.Errorf("id conflict: %v", allCodes(ds))
	}
	versionConflict := "---\ndoc_id: SOP-QM-001\nversion: \"1.1\"\n---\n**Version:** 1.0\n"
	if hasCode(Check(versionConflict, Options{}), Error, "header-conflict") == nil {
		t.Error("version conflict not reported")
	}

	missing := "---\ntitle: t\nevent_kinds: []\n---\nText\n"
	ds = Check(missing, Options{RequireHeader: true})
	if hasCode(ds, Warning, "header-id-missing") == nil || hasCode(ds, Warning, "header-version-missing") == nil {
		t.Errorf("missing id/version: %v", allCodes(ds))
	}

	wm := "---\ndoc_id: SOP-QM-001\nwatermark: FREIGEGEBEN\n---\nText\n"
	if d := hasCode(Check(wm, Options{}), Error, "frontmatter-server-owned"); d == nil || d.Line != 3 {
		t.Errorf("watermark: %v", allCodes(Check(wm, Options{})))
	}
}

func TestVariables(t *testing.T) {
	md := "---\ndoc_id: FB-QM-001\nversion: \"2.0\"\ntitle: CAPA\n---\n# {{title}}\n\nGehört zu {{doc_id}}, Fassung {{ version }}.\n\n`{{doc_id}}` und\n\n```\n{{title}}\n```\n"
	got := ExpandVariables(md)
	want := "---\ndoc_id: FB-QM-001\nversion: \"2.0\"\ntitle: CAPA\n---\n# CAPA\n\nGehört zu FB-QM-001, Fassung 2.0.\n\n`{{doc_id}}` und\n\n```\n{{title}}\n```\n"
	if got != want {
		t.Errorf("ExpandVariables:\n%s\nwant:\n%s", got, want)
	}
	if st := StoredText(md); !strings.HasPrefix(st, "# CAPA\n") {
		t.Errorf("StoredText must expand and drop the upload front matter:\n%s", st)
	}

	bad := "---\ndoc_id: FB-QM-001\n---\n{{author}} {{kennung}} {{titel|yaml}} {{version}}\n"
	ds := Check(bad, Options{})
	if d := hasCode(ds, Error, "unknown-variable"); d == nil || d.Line != 4 || d.Col != 1 {
		t.Errorf("unknown variable: %v", allCodes(ds))
	}
	if n := strings.Count(strings.Join(allCodes(ds), " "), "error:template-placeholder-left"); n != 2 {
		t.Errorf("template placeholders: want 2, got %v", allCodes(ds))
	}
	if hasCode(ds, Warning, "variable-empty") == nil {
		t.Errorf("{{version}} without value: %v", allCodes(ds))
	}
	if !strings.Contains(ExpandVariables(bad), "{{version}}") {
		t.Error("a variable without a value must stay visible")
	}
}

func TestSplitSignatureBothForms(t *testing.T) {
	content := "---\ndoc_id: SOP-QM-001\nversion: \"1.0\"\n---\n# Inhalt\n\nText [@#ort ___]\n"

	sig, rest := SplitSignature(content + sigEnd)
	if sig == nil || sig.Form != "end" || rest != content {
		t.Fatalf("end form: %+v rest=%q", sig, rest)
	}
	if sig.SignedBy() != "Dr. Anna Berger" || sig.SignedAt() != "2026-10-19T14:02:11Z" || sig.Fields["sha256"] != "abc" || sig.Line != 8 || sig.EndLine != 17 {
		t.Errorf("end form fields/lines: %+v", sig)
	}

	sig, rest = SplitSignature(sigFront + content)
	if sig == nil || sig.Form != "front" || rest != content || sig.Meaning() != "approved" || sig.EndLine != 7 {
		t.Fatalf("front form: %+v rest=%q", sig, rest)
	}
	// The document's own front matter after the old block is still read,
	// with the right lines.
	if m := DocumentMeta(sigFront + content); m.DocID != "SOP-QM-001" {
		t.Errorf("front form meta: %+v", m)
	}
	ds := Check(sigFront+content, Options{RequireHeader: true})
	if hasCode(ds, Error, "frontmatter-unknown-key") != nil {
		t.Errorf("the signature block is no front matter: %v", allCodes(ds))
	}
	if d := hasCode(ds, Info, "signed"); d == nil || !strings.Contains(d.Message, "Dr. Anna Berger") {
		t.Errorf("signed info: %v", allCodes(ds))
	}
	if fs, _ := ExtractFields(sigFront + content); len(fs) != 1 || fs[0].Line != 14 {
		t.Errorf("field line after the old block: %+v", fs)
	}

	// Not a signature: other keys next to "signature:" (fail closed), a
	// comment in the middle, an ordinary front matter.
	for _, md := range []string{
		"---\nsignature:\n  doc_no: X\nwatermark: FREI\n---\nText\n",
		"Text\n" + sigEnd + "mehr Text\n",
		"Text " + strings.TrimSuffix(sigEnd, "\n"),
		content,
	} {
		if sig, rest := SplitSignature(md); sig != nil || rest != md {
			t.Errorf("no signature expected in %q: %+v", md, sig)
		}
	}

	// StoredText drops both forms; the checksum covers exactly the rest.
	if st := StoredText(content + sigEnd); strings.Contains(st, "mlc-signature") {
		t.Errorf("StoredText kept the signature: %q", st)
	}
	if st := StoredText(sigFront + content); strings.Contains(st, "signature") || !strings.HasPrefix(st, "# Inhalt") {
		t.Errorf("StoredText front form: %q", st)
	}
	_, rest = SplitSignature(content + sigEnd)
	sum := sha256.Sum256([]byte(rest))
	if hex.EncodeToString(sum[:]) != hex.EncodeToString(func() []byte { s := sha256.Sum256([]byte(content)); return s[:] }()) {
		t.Error("checksum must cover the content byte for byte")
	}
}

func TestLayoutYAML(t *testing.T) {
	md := "---\ndoc_id: INFO-QM-001\nversion: \"1.0\"\ntitle: test info\n---\nText\n"
	got := LayoutYAML(md, LayoutOptions{Watermark: "  DRAFT  ", Organization: "Demo GmbH"})
	want := "lang: de\ntitle: test info\nwatermark: DRAFT\nheader:\n    left: Information\n    center: Demo GmbH\n    right: INFO-QM-001 · Version 1.0\n"
	if got != want {
		t.Errorf("LayoutYAML:\n%s\nwant:\n%s", got, want)
	}
	if got := LayoutYAML("# Nur Text\n", LayoutOptions{}); got != "lang: de\n" {
		t.Errorf("plain Markdown: %q", got)
	}
	if got := LayoutYAML(md, LayoutOptions{Watermark: strings.Repeat("x", 80)}); !strings.Contains(got, "watermark: "+strings.Repeat("x", 60)+"\n") {
		t.Errorf("watermark not cut to 60: %s", got)
	}
}

func TestSuggestFileName(t *testing.T) {
	md := "---\ndoc_id: INFO-QM-001\nversion: \"1.0\"\n---\nText\n"
	for _, c := range []struct {
		ext  string
		ver  bool
		want string
	}{{".md", false, "INFO-QM-001.md"}, {".pdf", true, "INFO-QM-001_v1.0.pdf"}, {"html", true, "INFO-QM-001_v1.0.html"}} {
		if got := SuggestFileName(md, c.ext, c.ver); got != c.want {
			t.Errorf("%s/%v: %q, want %q", c.ext, c.ver, got, c.want)
		}
	}
	if got := SuggestFileName("# ohne Kennung\n", ".md", false); got != "" {
		t.Errorf("without doc_id: %q", got)
	}
	if got := SuggestFileName("---\ndoc_id: \"SOP/QM:1\"\nversion: v2\n---\n", ".pdf", true); got != "SOP_QM_1_v2.pdf" {
		t.Errorf("unsafe characters: %q", got)
	}
}

// testdata/meta.json: variables and signature cases with the expected
// result, shared with the TypeScript renderer (packages/mlc-markdown).
// Regenerate after a deliberate change: go test -run TestMetaVectors -update
type metaVector struct {
	Markdown  string         `json:"markdown"`
	Meta      metaVectorMeta `json:"meta"`
	Expanded  string         `json:"expanded"`
	Signature *metaVectorSig `json:"signature"`
	Content   string         `json:"content"`
	Stored    string         `json:"stored"`
}

type metaVectorMeta struct {
	DocID   string `json:"docId"`
	Version string `json:"version"`
	Title   string `json:"title"`
}

type metaVectorSig struct {
	Form   string            `json:"form"`
	Fields map[string]string `json:"fields"`
}

var metaInputs = []string{
	"---\ndoc_id: FB-QM-001\nversion: \"2.0\"\ntitle: CAPA\n---\n# {{title}}\n\nGehört zu {{doc_id}}, Fassung {{ version }}.\n\n`{{doc_id}}` und\n\n```\n{{title}}\n```\n",
	"---\ntitle: t\n---\n**Dokumenten-ID:** SOP-QM-002 | **Version:** 2.1\n\n{{doc_id}} / {{version}} / {{unbekannt}} / {{kennung|yaml}}\n",
	"---\ndoc_id: SOP-QM-001\n---\n{{version}} bleibt stehen\n",
	"---\ndoc_id: SOP-QM-001\nversion: \"1.0\"\n---\n# Inhalt {{doc_id}}\n" + sigEnd,
	sigFront + "---\ndoc_id: SOP-QM-001\nversion: \"1.0\"\ntitle: Alt\n---\n# {{title}}\n",
	"---\nsignature:\n  doc_no: X\nwatermark: FREI\n---\nText {{doc_id}}\n",
	"Text\n" + sigEnd + "mehr Text\n",
	"# Ohne Front Matter {{title}}\n",
	"---\ndoc_id: SOP-QM-001\nversion: \"1.0\"\nheader:\n  left: X\n---\n{{doc_id}}\n",
}

func metaVectors() []metaVector {
	out := []metaVector{}
	for _, md := range metaInputs {
		m := DocumentMeta(md)
		sig, rest := SplitSignature(md)
		v := metaVector{
			Markdown: md, Meta: metaVectorMeta{DocID: m.DocID, Version: m.Version, Title: m.Title},
			Expanded: ExpandVariables(md), Content: rest, Stored: StoredText(md),
		}
		if sig != nil {
			v.Signature = &metaVectorSig{Form: sig.Form, Fields: sig.Fields}
		}
		out = append(out, v)
	}
	return out
}

func TestMetaVectors(t *testing.T) {
	got := metaVectors()
	if *update {
		b, _ := json.MarshalIndent(got, "", "  ")
		if err := os.WriteFile("testdata/meta.json", append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	b, err := os.ReadFile("testdata/meta.json")
	if err != nil {
		t.Fatalf("%v (run: go test -run TestMetaVectors -update)", err)
	}
	var want []metaVector
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("testdata/meta.json is out of date; regenerate with -update and check the diff")
	}
}
