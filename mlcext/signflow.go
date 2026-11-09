package mlcext

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// SignatureStep is one signature in a form's signing flow, from the front
// matter block "signatures:" — the same form as the Sensor Monitor's catalog:
//
//	signatures:
//	  - field: sig_qmb          # the signature field [@#sig_qmb Unterschrift: …]
//	    who: "Qualitätsbeauftragte/r (QMB)"
//	    right: K                # a right in the Sensor Monitor; "" = anyone signed in
//	    after: [sig_entdecker]  # signed before
//	    requires: [ursache, risiko_kritisch|risiko_major]  # filled in before; a|b = one of them
//	    not_same: [sig_leitung] # four-eyes principle
//	    when: "nur bei einer Datenlücke"  # conditional: may be skipped with a reason
//
// This package describes and checks the flow; enforcing it is up to the
// tool that collects the signatures. Right is a term of the Sensor Monitor
// and only text here.
type SignatureStep struct {
	Field    string   `yaml:"field" json:"field"`
	Who      string   `yaml:"who" json:"who"`
	Right    string   `yaml:"right,omitempty" json:"right,omitempty"`
	After    []string `yaml:"after,omitempty" json:"after,omitempty"`
	Requires []string `yaml:"requires,omitempty" json:"requires,omitempty"`
	NotSame  []string `yaml:"not_same,omitempty" json:"notSame,omitempty"`
	When     string   `yaml:"when,omitempty" json:"when,omitempty"`
	// Line of the step in the file (1-based), 0 when not read from a file.
	Line int `yaml:"-" json:"line,omitempty"`
}

// signatureStepKeys are the keys a step may have.
var signatureStepKeys = []string{"field", "who", "right", "after", "requires", "not_same", "when"}

// SignatureFlow returns the signing flow of md from its front matter, nil
// when it has none. The error says what cannot be read; Check reports the
// details.
func SignatureFlow(md string) ([]SignatureStep, error) {
	v, yamlLine := signaturesNode(md)
	if v == nil {
		return nil, nil
	}
	var steps []SignatureStep
	if err := v.Decode(&steps); err != nil {
		return nil, fmt.Errorf("signatures: %w", err)
	}
	for i := range steps {
		steps[i].Line = v.Content[i].Line + yamlLine - 1
	}
	return steps, nil
}

// IsSignatureFlowNode tells the two meanings of "signatures:" in the front
// matter apart. mlcdoc uses the key for printed signature lines (a list of
// role, name, hint, caption); this package for the signing flow. A list
// whose items are all blocks without a step key is mlcdoc's; everything
// else is a flow (and is checked as one). mlcdoc rejects a flow, so
// mlcdoc.Prepare removes it before rendering.
func IsSignatureFlowNode(v *yaml.Node) bool {
	if v == nil || v.Kind != yaml.SequenceNode || len(v.Content) == 0 {
		return v != nil
	}
	for _, item := range v.Content {
		if item.Kind != yaml.MappingNode {
			return true
		}
		for j := 0; j < len(item.Content); j += 2 {
			if slices.Contains(signatureStepKeys, item.Content[j].Value) {
				return true
			}
		}
	}
	return false
}

// signaturesNode is the value of "signatures:" in the front matter and the
// file line of the first YAML line; nil when there is none or it is
// mlcdoc's signature lines (IsSignatureFlowNode).
func signaturesNode(md string) (*yaml.Node, int) {
	_, _, node, err := parseFrontMatter(md)
	if err != nil || node == nil || len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return nil, 0
	}
	m := node.Content[0]
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == "signatures" {
			if v := m.Content[i+1]; IsSignatureFlowNode(v) {
				return v, frontMatterBlock(md).yamlLine
			}
			return nil, 0
		}
	}
	return nil, 0
}

// checkSignatureFlow checks the flow against the document's fields, with
// the rules of the Sensor Monitor (ValidateFlow): every step names a named
// signature field, each once, and says who signs; after and not_same name
// other steps; requires names fields that are no signature ("a|b" = one of
// them); no loop. A signature field without a step is a warning.
func checkSignatureFlow(md string, fields []Field, add func(int, int, Severity, string, string)) {
	v, yamlLine := signaturesNode(md)
	if v == nil {
		return
	}
	at := func(n *yaml.Node) int { return n.Line + yamlLine - 1 }
	if v.Kind != yaml.SequenceNode {
		add(at(v), v.Column, Error, "signatures-yaml", "`signatures` muss eine Liste von Schritten sein (`- field: sig_…`).")
		return
	}
	for _, item := range v.Content {
		if item.Kind != yaml.MappingNode {
			add(at(item), item.Column, Error, "signatures-yaml", "Ein Schritt in `signatures` ist ein Block aus `schlüssel: wert`-Zeilen (`field`, `who`, …).")
			return
		}
		for j := 0; j < len(item.Content); j += 2 {
			if k := item.Content[j]; !slices.Contains(signatureStepKeys, k.Value) {
				add(at(k), k.Column, Error, "signatures-unknown-key", fmt.Sprintf("Unbekannter Schlüssel `%s` in einem Schritt; erlaubt sind %s.", k.Value, strings.Join(signatureStepKeys, ", ")))
			}
		}
	}
	steps, err := SignatureFlow(md)
	if err != nil {
		add(at(v), v.Column, Error, "signatures-yaml", "`signatures` ist nicht lesbar: "+err.Error())
		return
	}
	sigs, others := map[string]bool{}, map[string]bool{}
	for _, f := range fields {
		switch {
		case !f.Named:
		case f.Kind == SignatureField:
			sigs[f.Name] = true
		default:
			others[f.Name] = true
		}
	}
	inFlow := map[string]bool{}
	for _, st := range steps {
		switch {
		case st.Field == "":
			add(st.Line, 1, Error, "signatures-field", "Schritt ohne `field`: Welches Unterschriftsfeld wird hier geleistet?")
		case !sigs[st.Field]:
			add(st.Line, 1, Error, "signatures-field", fmt.Sprintf("`%s` ist kein benanntes Unterschriftsfeld des Formblatts (`[@#%s Unterschrift: …]`).", st.Field, st.Field))
		case inFlow[st.Field]:
			add(st.Line, 1, Error, "signatures-duplicate", fmt.Sprintf("`%s` steht zweimal im Ablauf.", st.Field))
		}
		inFlow[st.Field] = true
		if strings.TrimSpace(st.Who) == "" {
			add(st.Line, 1, Error, "signatures-who", fmt.Sprintf("Schritt `%s`: `who` fehlt — wer unterschreibt hier?", st.Field))
		}
	}
	for _, st := range steps {
		for _, a := range slices.Concat(st.After, st.NotSame) {
			if !inFlow[a] || a == st.Field {
				add(st.Line, 1, Error, "signatures-ref", fmt.Sprintf("Schritt `%s` verweist auf `%s` — kein anderer Schritt des Ablaufs.", st.Field, a))
			}
		}
		for _, req := range st.Requires {
			for _, alt := range strings.Split(req, "|") {
				if alt = strings.TrimSpace(alt); !others[alt] {
					add(st.Line, 1, Error, "signatures-requires", fmt.Sprintf("Schritt `%s` verlangt `%s` — kein Feld des Formblatts, das ausgefüllt wird.", st.Field, alt))
				}
			}
		}
	}
	if st, ok := flowLoop(steps); ok {
		add(st.Line, 1, Error, "signatures-loop", fmt.Sprintf("Schritt `%s` kommt nie an die Reihe: Schleife in `after`.", st.Field))
	}
	for _, f := range fields {
		if f.Named && f.Kind == SignatureField && !inFlow[f.Name] {
			add(f.Line, f.Col, Warning, "signatures-missing", fmt.Sprintf("Unterschriftsfeld `%s` hat keinen Schritt in `signatures`.", f.Name))
		}
	}
}

// flowLoop returns the first step that never comes up because "after"
// forms a loop. A reference to no step is reported elsewhere and ignored.
func flowLoop(steps []SignatureStep) (SignatureStep, bool) {
	known := map[string]bool{}
	for _, st := range steps {
		known[st.Field] = true
	}
	done := map[string]bool{}
	for changed := true; changed; {
		changed = false
		for _, st := range steps {
			if done[st.Field] {
				continue
			}
			if !slices.ContainsFunc(st.After, func(p string) bool { return known[p] && p != st.Field && !done[p] }) {
				done[st.Field], changed = true, true
			}
		}
	}
	for _, st := range steps {
		if !done[st.Field] {
			return st, true
		}
	}
	return SignatureStep{}, false
}
