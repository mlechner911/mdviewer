package mlcext

import (
	"os"
	"strings"
	"testing"
)

func TestSetFormRecord(t *testing.T) {
	src, err := os.ReadFile("testdata/form_btm.md")
	if err != nil {
		t.Fatal(err)
	}
	form := string(src)
	ref := FormRefOf(form)
	if ref.DocID != "FB-LAG-001" || ref.Version != "2.0" || len(ref.SHA256) != 64 {
		t.Fatalf("ref %+v", ref)
	}
	filled, _ := FillFields(form, map[string]string{"datum": "28.10.2026"})
	rec, err := SetFormRecord(filled, FormRecord{Form: ref, FilledBy: "Dr. Anna Berger", CompletedAt: "2026-10-28T09:12:00+01:00",
		Context: map[string]string{"event_no": "EV-2026-2069"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rec, "---\ndoc_id: FB-LAG-001\nversion: \"2.0\"\n") || !strings.Contains(rec, "\nmeta:\n  schema: mlc-form-record/v1\n  form:\n    doc_id: FB-LAG-001\n") {
		t.Fatalf("front matter:\n%s", rec[:400])
	}
	got, ok := FormRecordOf(rec)
	if !ok || got.FilledBy != "Dr. Anna Berger" || got.Context["event_no"] != "EV-2026-2069" || got.Form != ref {
		t.Fatalf("%+v %v", got, ok)
	}
	// The record still pins the blank form, and the values are still there.
	if FormRefOf(rec) != ref || FieldValues(rec)["datum"] != "28.10.2026" {
		t.Fatal("ref or values lost")
	}
	for _, d := range Check(rec, Options{}) {
		if d.Severity != Info {
			t.Errorf("record: %s %s", d.Code, d.Message)
		}
	}
	// Writing it again replaces the block, it does not add a second one.
	again, err := SetFormRecord(rec, FormRecord{Form: ref, FilledBy: "B"})
	if err != nil || strings.Count(again, "meta:") != 1 || !strings.Contains(again, "filled_by: B") {
		t.Fatalf("again:\n%s", again)
	}
}

func TestSetFormRecordWithoutFrontMatterAndSigned(t *testing.T) {
	md := "# Prüfung\n[@#ok]\n\n<!--mlc-signature\ndoc_no: SOP-2026-0001\nsha256: abc\n-->\n"
	rec, err := SetFormRecord(md, FormRecord{FilledBy: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rec, "---\nmeta:\n  schema: mlc-form-record/v1\n") || strings.Contains(rec, "mlc-signature") || !strings.Contains(rec, "---\n\n# Prüfung") {
		t.Fatalf("got:\n%s", rec)
	}
	if !IsFormRecord(rec) {
		t.Fatal("not a record")
	}
}
