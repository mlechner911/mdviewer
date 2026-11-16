package mlcext

import "testing"

func TestIsMarkdownFile(t *testing.T) {
	for name, want := range map[string]bool{
		"a.md": true, "A.MD": true, "a.markdown": true, "a.mdown": true,
		"SOP-QM-001.mlcmd": true, "x.MLCMD": true,
		"a.txt": false, "mlcmd": false, "a.md.bak": false, "": false,
	} {
		if got := IsMarkdownFile(name); got != want {
			t.Errorf("IsMarkdownFile(%q) = %v", name, got)
		}
	}
	if got := TrimMarkdownExt("Messung 05.03.2026.mlcmd"); got != "Messung 05.03.2026" {
		t.Errorf("TrimMarkdownExt: %q", got)
	}
	if got := TrimMarkdownExt("notes.txt"); got != "notes.txt" {
		t.Errorf("TrimMarkdownExt: %q", got)
	}
}

func TestSaveExt(t *testing.T) {
	if got := SaveExt("# Notizen\n\nNur Text.\n"); got != ".md" {
		t.Errorf("plain: %q", got)
	}
	if got := SaveExt("---\ndoc_id: SOP-QM-001\n---\n# SOP\n"); got != ExtMLC {
		t.Errorf("doc_id: %q", got)
	}
	if got := SaveExt("Ort: [@#ort ____]\n"); got != ExtMLC {
		t.Errorf("field: %q", got)
	}
}
