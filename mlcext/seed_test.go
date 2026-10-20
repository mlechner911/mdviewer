package mlcext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSeedTemplates(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "vorlagen")
	written, err := SeedTemplates(dir)
	if err != nil || len(written) != 17 { // 8 templates × 2 files + LIESMICH.md
		t.Fatalf("first seed: %v %v", written, err)
	}
	if got := LoadTemplatesFrom(false, dir); len(got) != 8 || got[0].Source != dir {
		t.Fatalf("seeded folder: %d templates", len(got))
	}
	// The user edits one, deletes another: a second seed (an update)
	// changes neither.
	_ = os.WriteFile(filepath.Join(dir, "sop.md"), []byte("# eigene SOP {{title}}\n"), 0o644)
	_ = os.Remove(filepath.Join(dir, "leer.json"))
	_ = os.Remove(filepath.Join(dir, "leer.md"))
	if written, err := SeedTemplates(dir); err != nil || len(written) != 0 {
		t.Fatalf("second seed: %v %v", written, err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "sop.md")); string(b) != "# eigene SOP {{title}}\n" {
		t.Errorf("user change overwritten: %q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "leer.json")); err == nil {
		t.Error("deleted template came back")
	}
	if got := LoadTemplatesFrom(false, dir); len(got) != 7 {
		t.Errorf("want 7 templates after deleting one, got %d", len(got))
	}
}
