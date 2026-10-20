package mdbrowse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentsListsOnlyMLCDocuments(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		p := filepath.Join(dir, filepath.FromSlash(name))
		_ = os.MkdirAll(filepath.Dir(p), 0o700)
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("README.md", "# Readme\n\nnormal\n")
	write("sop/kal.md", "---\ntitle: Kalibrierung\n---\n# SOP\n\n**Dokumenten-ID:** SOP-KAL-001 | **Version:** 1.1\n\n---\n")
	write("fb/capa.md", "# Formblatt: CAPA\n\n**Dokumenten-ID:** FB-QM-002 | **Version:** 2.1\n\nOrt: [@#ort ___] [@#x\n")
	write("ohne.md", "# Ohne Kennung\n\n[@#feld ___]\n")
	write("_vorlagen/sop.md", "# {{TITEL}}\n\n**Dokumenten-ID:** {{KENNUNG}}\n")

	root, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	docs, err := root.Documents()
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 3 {
		t.Fatalf("want 3 documents, got %+v", docs)
	}
	if d := docs[0]; d.RefNo != "FB-QM-002" || d.Kind != "form" || d.Fields != 1 || d.Errors != 1 || d.Title != "Formblatt: CAPA" {
		t.Errorf("doc 0: %+v", d)
	}
	if d := docs[1]; d.RefNo != "SOP-KAL-001" || d.Title != "Kalibrierung" || d.Version != "1.1" || d.Path != "sop/kal.md" {
		t.Errorf("doc 1: %+v", d)
	}
	if d := docs[2]; d.RefNo != "" || d.Path != "ohne.md" || d.Warnings != 2 || d.NameMismatch {
		t.Errorf("doc 2 (no RefNo, last): %+v", d)
	}
	if !docs[0].NameMismatch {
		t.Errorf("capa.md with FB-QM-002: name mismatch expected")
	}
	write("neu/INFO-QM-001.md", "---\ndoc_id: INFO-QM-001\nversion: \"1.0\"\n---\nText\n")
	docs, _ = root.Documents()
	for _, d := range docs {
		if d.RefNo == "INFO-QM-001" && (d.NameMismatch || d.Version != "1.0") {
			t.Errorf("doc_id from the front matter, matching name: %+v", d)
		}
	}
}

func TestExtractTitleIgnoresFrontMatterComments(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a.md")
	_ = os.WriteFile(p, []byte("---\n# Bei welchen Ereignissen …\nevent_kinds: []\n---\n# Echte Überschrift\n"), 0o600)
	if got, _ := ExtractTitle(p); got != "Echte Überschrift" {
		t.Errorf("ExtractTitle = %q", got)
	}
}
