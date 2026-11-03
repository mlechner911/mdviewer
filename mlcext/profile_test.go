package mlcext

import (
	"slices"
	"testing"
)

const profileDoc = `---
doc_id: SOP-KUE-001
version: "1.0"
title: Kühlschrank zu warm
profile:
  purpose: Was tun, wenn der Kühlschrank zu warm ist.
  appears: Bei Grenzwert-Ereignissen an Kühlgeräten.
  adapt:
    - name: verantwortlich
      note: Wer über die Ware entscheidet.
    - name: notdienst
      note: Telefon des Notdienstes.
  keep:
    - name: Ereignis-Nr.
      note: Füllt das System aus dem Ereignis.
  orientation:
    - ApBetrO § 2a
---
# {{title}}

Entscheidet: {{verantwortlich}}. Notdienst: {{notdienst}}. Code: ` + "`{{nicht_gemeint}}`" + `
`

// Steckbrief (Sensor Monitor P-SENS-20261103-02): Ein in profile.adapt erklärter Platzhalter ist
// beim Hochladen nur eine Warnung — der Entwurf darf gespeichert werden; ein unerklärter bleibt
// ein Fehler (Gegenprobe), der Schlüssel profile ist bekannt.
func TestProfileDeclaredPlaceholdersAreWarnings(t *testing.T) {
	ds := Check(profileDoc, Options{})
	if errs := codes(ds, Error); len(errs) != 0 {
		t.Fatalf("Fehler im gültigen Steckbrief: %v", errs)
	}
	if n := len(slices.DeleteFunc(codes(ds, Warning), func(c string) bool { return c != "placeholder-open" })); n != 2 {
		t.Errorf("%d Warnungen placeholder-open, erwartet 2: %v", n, ds)
	}
	undeclared := Check(profileDoc+"\nUnd {{fremd}}.\n", Options{})
	if !slices.Contains(codes(undeclared, Error), "unknown-variable") {
		t.Errorf("unerklärter Platzhalter kein Fehler: %v", undeclared)
	}
	p, ok := ParseProfile(profileDoc)
	if !ok || p.Purpose == "" || !slices.Equal(p.Placeholders(), []string{"verantwortlich", "notdienst"}) || len(p.Keep) != 1 || len(p.Orientation) != 1 {
		t.Errorf("ParseProfile: %+v %v", p, ok)
	}
}

// Offen ist, was noch als {{name}} im Text steht — auch im gespeicherten Text ohne Front Matter,
// nicht in Code.
func TestOpenPlaceholders(t *testing.T) {
	stored := "# Titel\n\nEntscheidet: {{verantwortlich}}. Notdienst: 0941 12345.\n`{{notdienst}}`\n"
	if got := OpenPlaceholders(stored, []string{"verantwortlich", "notdienst"}); !slices.Equal(got, []string{"verantwortlich"}) {
		t.Errorf("offen: %v, erwartet [verantwortlich]", got)
	}
	if got := OpenPlaceholders("# fertig\n", []string{"verantwortlich"}); len(got) != 0 {
		t.Errorf("nichts offen erwartet: %v", got)
	}
}

func TestProfileErrors(t *testing.T) {
	for name, tc := range map[string]struct{ profile, code string }{
		"ohne purpose":          {"profile:\n  appears: x\n", "profile-purpose"},
		"unbekannter Schlüssel": {"profile:\n  purpose: x\n  wozu: y\n", "profile-unknown-key"},
		"Variable als Name":     {"profile:\n  purpose: x\n  adapt:\n    - name: title\n      note: n\n", "profile-placeholder-name"},
		"zweimal":               {"profile:\n  purpose: x\n  adapt:\n    - name: a\n      note: n\n    - name: a\n      note: n\n", "profile-placeholder-twice"},
		"kein Block":            {"profile: Text\n", "profile-yaml"},
	} {
		ds := Check("---\ndoc_id: SOP-X-001\n"+tc.profile+"---\n# T\n{{a}}\n", Options{})
		if !slices.Contains(codes(ds, Error), tc.code) {
			t.Errorf("%s: %v, erwartet %s", name, codes(ds, Error), tc.code)
		}
	}
}
