package mlcext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryTemplateCreatesAValidDocument(t *testing.T) {
	for _, tpl := range Templates() {
		md, err := NewDocument(NewParams{Kind: tpl.Kind, Title: "Lagerung: Kühlkette & Co.", RefNo: tpl.Prefix + "-QM-001"})
		if err != nil {
			t.Fatalf("%s: %v", tpl.Kind, err)
		}
		if strings.Contains(md, "{{") {
			t.Errorf("%s: placeholder left:\n%s", tpl.Kind, md)
		}
		if !Detect(md) {
			t.Errorf("%s: not detected as MLC document", tpl.Kind)
		}
		for _, d := range Check(md, Options{RequireHeader: true}) {
			if d.Severity != Info {
				t.Errorf("%s: %d:%d %s %s", tpl.Kind, d.Line, d.Col, d.Code, d.Message)
			}
		}
		h := ParseHeader(md)
		if h.RefNo != tpl.Prefix+"-QM-001" || h.Version != "1.0" || h.Kind != tpl.Kind {
			t.Errorf("%s: header %+v", tpl.Kind, h)
		}
		if fm, _ := ParseFrontMatter(md); fm.Title != "Lagerung: Kühlkette & Co." {
			t.Errorf("%s: front matter title %q", tpl.Kind, fm.Title)
		}
	}
}

func TestNewDocumentRejectsBadInput(t *testing.T) {
	for _, p := range []NewParams{
		{Kind: "sop", Title: "", RefNo: "SOP-QM-001"},
		{Kind: "sop", Title: "X", RefNo: ""},
		{Kind: "form", Title: "X", RefNo: "SOP-QM-001"},
		{Kind: "sop", Title: "X", RefNo: "SOP-QM-001 | **Version:** 9"},
		{Kind: "memo", Title: "X", RefNo: "MEMO-1"},
	} {
		if _, err := NewDocument(p); err == nil {
			t.Errorf("%+v: want an error", p)
		}
	}
}

func TestSuggestRefNo(t *testing.T) {
	existing := []string{"FB-QM-002", "FB-QM-010", "FB-TE-001", "SOP-QM-006", "fb-qm-0042x"}
	for _, c := range []struct{ kind, area, want string }{
		{"form", "QM", "FB-QM-011"},
		{"form", "te", "FB-TE-002"},
		{"form", "LAB", "FB-LAB-001"},
		{"sop", "QM", "SOP-QM-007"},
		{"sop", "", ""},
	} {
		if got := SuggestRefNo(c.kind, c.area, existing); got != c.want {
			t.Errorf("SuggestRefNo(%s, %s) = %q, want %q", c.kind, c.area, got, c.want)
		}
	}
	if got := SuggestRefNo("sop", "KAL", []string{"SOP-KAL-0007"}); got != "SOP-KAL-0008" {
		t.Errorf("keeps the width: %q", got)
	}
}

func TestRefNosInDir(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		p := filepath.Join(dir, name)
		_ = os.MkdirAll(filepath.Dir(p), 0o700)
		_ = os.WriteFile(p, []byte(content), 0o600)
	}
	write("a.md", "**Dokumenten-ID:** SOP-QM-001\n")
	write("sub/b.md", "**Dokumenten-ID:** FB-QM-002 | **Version:** 1\n")
	write(".hidden/c.md", "**Dokumenten-ID:** SOP-QM-099\n")
	write("d.txt", "**Dokumenten-ID:** SOP-QM-050\n")
	got := strings.Join(RefNosInDir(dir, 100), ",")
	if got != "SOP-QM-001,FB-QM-002" {
		t.Errorf("RefNosInDir = %s", got)
	}
}
