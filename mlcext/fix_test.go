package mlcext

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestMoveHeaderToFrontMatter(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{
			"---\ntitle: test info\nevent_kinds: []\n---\n# Information: test info\n\n**Dokumenten-ID:** INFO-QM-001 | **Version:** 1.0\n\nText\n",
			"---\ndoc_id: INFO-QM-001\nversion: \"1.0\"\ntitle: test info\nevent_kinds: []\n---\n# Information: test info\n\n\nText\n",
		},
		{
			"# SOP\n\n**Titel:** Kühlkette\\\n**Dokumenten-ID:** SOP-QM-002\\\n**Version:** 2.1\n\nText\n",
			"---\ndoc_id: SOP-QM-002\nversion: \"2.1\"\ntitle: Kühlkette\n---\n# SOP\n\n\nText\n",
		},
		{ // front matter wins, the line keeps its other text
			"---\ndoc_id: SOP-QM-001\n---\nStand: **Dokumenten-ID:** SOP-QM-009 | **Version:** 3 | Seite 1\n",
			"---\nversion: \"3\"\ndoc_id: SOP-QM-001\n---\nStand: Seite 1\n",
		},
	} {
		got, ok := MoveHeaderToFrontMatter(c.in)
		if !ok || got != c.want {
			t.Errorf("MoveHeaderToFrontMatter(%q) =\n%q\nwant\n%q", c.in, got, c.want)
			continue
		}
		if hasCode(Check(got, Options{RequireHeader: true}), Warning, "header-in-body") != nil {
			t.Errorf("still header-in-body after the fix:\n%s", got)
		}
	}
	if _, ok := MoveHeaderToFrontMatter("---\ndoc_id: X\n---\nText\n"); ok {
		t.Error("nothing to move must report ok=false")
	}
}

func TestCheckSignature(t *testing.T) {
	content := "---\ndoc_id: SOP-QM-001\n---\n# Inhalt\n"
	sum := sha256.Sum256([]byte(content))
	signed := content + strings.Replace(sigEnd, "sha256: abc", "sha256: "+hex.EncodeToString(sum[:]), 1)
	if s := CheckSignature(signed); !s.Signed || !s.Intact || s.SignedBy != "Dr. Anna Berger" {
		t.Errorf("intact: %+v", s)
	}
	edited := strings.Replace(signed, "# Inhalt", "# Inhalt geändert", 1)
	if s := CheckSignature(edited); !s.Signed || s.Intact {
		t.Errorf("edited: %+v", s)
	}
	if RemoveSignature(edited) != "---\ndoc_id: SOP-QM-001\n---\n# Inhalt geändert\n" {
		t.Errorf("RemoveSignature: %q", RemoveSignature(edited))
	}
	if s := CheckSignature(content); s.Signed {
		t.Errorf("unsigned: %+v", s)
	}
}
