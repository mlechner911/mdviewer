package mdbrowse

import "testing"

func TestSourceAndResolveLinks(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "sop/index.md", "---\ntitle: SOP\n---\n# Übersicht\n")
	writeFixture(t, dir, "sop/a.md", "# A\n")
	writeFixture(t, dir, "sop/bild.png", "png")
	writeFixture(t, dir, "sop/plan.pdf", "pdf")
	writeFixture(t, dir, "leer/x.txt", "x")
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}

	doc, ok, err := root.Source("sop")
	if err != nil || !ok {
		t.Fatalf("Source(dir) = %v, %v", ok, err)
	}
	if doc.Path != "sop/index.md" || doc.Source != "---\ntitle: SOP\n---\n# Übersicht\n" || doc.Size == 0 {
		t.Errorf("unexpected doc %+v", doc)
	}
	if _, ok, err := root.Source("leer"); ok || err != nil {
		t.Errorf("directory without index.md: ok=%v err=%v", ok, err)
	}
	if _, _, err := root.Source("sop/plan.pdf"); err == nil {
		t.Error("non-markdown file must be rejected")
	}
	if _, _, err := root.Source("../x.md"); err == nil {
		t.Error("escape must be rejected")
	}

	got := root.ResolveLinks("sop/index.md", []string{
		"a.md#teil", "bild.png", "plan.pdf", "../leer", "fehlt.md", "../../etc/passwd",
		"https://example.com", "#anker", "mailto:x@y.de",
	})
	want := map[string]LinkTarget{
		"a.md#teil":           {Kind: "md", Path: "sop/a.md"},
		"bild.png":            {Kind: "file", Path: "sop/bild.png", Image: true},
		"plan.pdf":            {Kind: "file", Path: "sop/plan.pdf"},
		"../leer":             {Kind: "md", Path: "leer"},
		"fehlt.md":            {Kind: "broken", Reason: "Datei nicht gefunden: fehlt.md"},
		"../../etc/passwd":    {Kind: "broken", Reason: "Ziel außerhalb des Verzeichnisses"},
		"https://example.com": {Kind: "keep"},
		"#anker":              {Kind: "keep"},
		"mailto:x@y.de":       {Kind: "keep"},
	}
	for href, w := range want {
		if got[href] != w {
			t.Errorf("%s: got %+v, want %+v", href, got[href], w)
		}
	}
}
