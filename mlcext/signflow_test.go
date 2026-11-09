package mlcext

import (
	"slices"
	"strings"
	"testing"
)

const flowBody = `
# CAPA

Ursache: [@#ursache ________]
Risiko: [@#risiko_major] major [@#risiko_minor] minor

[@#sig_entdecker Unterschrift: Entdecker]
[@#sig_qmb Unterschrift: QMB]
[@#sig_leitung Unterschrift: Leitung]
`

func flowDoc(yaml string) string {
	return "---\ndoc_id: FB-QM-002\nversion: \"2.2\"\nsignatures:\n" + yaml + "---\n" + flowBody
}

const validFlow = `  - field: sig_entdecker
    who: "wer die Abweichung bemerkt hat"
    right: ""
  - field: sig_qmb
    who: "Qualitätsbeauftragte/r (QMB)"
    right: K
    after: [sig_entdecker]
    requires: [ursache, risiko_major|risiko_minor]
  - field: sig_leitung
    who: Leitung
    after: [sig_qmb]
    not_same: [sig_qmb]
    when: "nur bei kritischem Risiko"
`

func TestSignatureFlowRead(t *testing.T) {
	md := flowDoc(validFlow)
	steps, err := SignatureFlow(md)
	if err != nil || len(steps) != 3 {
		t.Fatalf("%v %+v", err, steps)
	}
	q := steps[1]
	if q.Field != "sig_qmb" || q.Right != "K" || !slices.Equal(q.After, []string{"sig_entdecker"}) ||
		!slices.Equal(q.Requires, []string{"ursache", "risiko_major|risiko_minor"}) || q.Line != 8 {
		t.Errorf("%+v", q)
	}
	if l := steps[2]; !slices.Equal(l.NotSame, []string{"sig_qmb"}) || l.When == "" {
		t.Errorf("%+v", l)
	}
	if ds := Check(md, Options{}); len(codes(ds, Error))+len(codes(ds, Warning)) > 0 {
		t.Fatalf("valid flow: %+v", ds)
	}
	if s, err := SignatureFlow("# ohne\n"); s != nil || err != nil {
		t.Errorf("no flow: %v %v", s, err)
	}
}

func TestSignatureFlowCheck(t *testing.T) {
	for _, c := range []struct{ name, yaml, code string }{
		{"not a list", "  field: sig_qmb\n", "signatures-yaml"},
		{"unknown key", strings.Replace(validFlow, "    right: K\n", "    recht: K\n", 1), "signatures-unknown-key"},
		{"no signature field", strings.Replace(validFlow, "field: sig_leitung", "field: ursache", 1), "signatures-field"},
		{"twice", validFlow + "  - field: sig_qmb\n    who: x\n", "signatures-duplicate"},
		{"who missing", strings.Replace(validFlow, "    who: Leitung\n", "", 1), "signatures-who"},
		{"after names no step", strings.Replace(validFlow, "after: [sig_qmb]", "after: [sig_chef]", 1), "signatures-ref"},
		{"not_same itself", strings.Replace(validFlow, "not_same: [sig_qmb]", "not_same: [sig_leitung]", 1), "signatures-ref"},
		{"requires a signature", strings.Replace(validFlow, "requires: [ursache,", "requires: [sig_entdecker,", 1), "signatures-requires"},
		{"requires unknown alternative", strings.Replace(validFlow, "risiko_major|risiko_minor", "risiko_major|risiko_x", 1), "signatures-requires"},
		{"loop", strings.Replace(validFlow, "    right: \"\"\n", "    after: [sig_leitung]\n", 1), "signatures-loop"},
		{"field without step", strings.SplitAfter(validFlow, "requires: [ursache, risiko_major|risiko_minor]\n")[0], "signatures-missing"},
	} {
		t.Run(c.name, func(t *testing.T) {
			ds := Check(flowDoc(c.yaml), Options{})
			if has(ds, c.code) == nil {
				t.Errorf("want %s, got %+v", c.code, ds)
			}
		})
	}
}

// mlcdoc's printed signature lines use the same key; they are no flow.
func TestSignatureFlowLeavesMlcdocLines(t *testing.T) {
	md := "---\ntitle: T\nsignatures:\n  - role: QMB\n    name: A. Muster\n---\n# T\n\n[@#sig_a Unterschrift: A]\n"
	if s, err := SignatureFlow(md); s != nil || err != nil {
		t.Errorf("%v %v", s, err)
	}
	if ds := Check(md, Options{}); len(codes(ds, Error)) > 0 || has(ds, "signatures-missing") != nil {
		t.Errorf("%+v", ds)
	}
	if StoredText(md) != md {
		t.Errorf("StoredText dropped the layout front matter")
	}
}
