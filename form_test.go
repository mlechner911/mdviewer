package main

import "testing"

func TestParseFormValues(t *testing.T) {
	got, err := parseFormValues([]byte(`{"ort": "Kühlraum 2", "ok": true, "nein": false, "menge": 2.5, "leer": null}`))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"ort": "Kühlraum 2", "ok": "x", "nein": "", "menge": "2.5", "leer": ""}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %q, want %q", k, got[k], v)
		}
	}
	if _, err := parseFormValues([]byte(`{"a": [1]}`)); err == nil {
		t.Error("list accepted")
	}
	if _, err := parseFormValues([]byte(`[1]`)); err == nil {
		t.Error("array accepted")
	}
}
