package mlcext

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"gopkg.in/yaml.v3"
)

// FormRef pins the form a record was filled from: the exact version, so a
// record can always be compared with the form it came from.
type FormRef struct {
	DocID   string `yaml:"doc_id,omitempty" json:"docId"`
	Version string `yaml:"version,omitempty" json:"version"`
	// DocNo is the Sensor Monitor's document number (SOP-2026-0031), when
	// the form came from there.
	DocNo string `yaml:"doc_no,omitempty" json:"docNo"`
	// SHA256 of the form's Markdown as it was filled in.
	SHA256 string `yaml:"sha256,omitempty" json:"sha256"`
}

// FormRecord is the "meta:" block of a filled-in form (FormRecordSchema).
type FormRecord struct {
	Schema string  `yaml:"schema" json:"schema"`
	Form   FormRef `yaml:"form" json:"form"`
	// Context: what the record belongs to ({event_no: EV-…} or {zone: …}).
	Context map[string]string `yaml:"context,omitempty" json:"context"`
	// FilledBy is a display name, never an e-mail address.
	FilledBy    string `yaml:"filled_by,omitempty" json:"filledBy"`
	CompletedAt string `yaml:"completed_at,omitempty" json:"completedAt"`
}

// FormRefOf is the form md is filled from: the "meta.form" of a record, else
// md itself (doc_id and version from its front matter, the SHA-256 of md).
func FormRefOf(md string) FormRef {
	if r, ok := FormRecordOf(md); ok {
		return r.Form
	}
	m := DocumentMeta(md)
	sum := sha256.Sum256([]byte(md))
	return FormRef{DocID: m.DocID, Version: m.Version, SHA256: hex.EncodeToString(sum[:])}
}

// FormRecordOf reads the "meta:" block of a filled-in form; false when md
// is none (IsFormRecord).
func FormRecordOf(md string) (FormRecord, bool) {
	_, _, node, err := parseFrontMatter(md)
	if err != nil || node == nil || len(node.Content) == 0 {
		return FormRecord{}, false
	}
	var fm struct {
		Meta FormRecord `yaml:"meta"`
	}
	if node.Content[0].Decode(&fm) != nil || fm.Meta.Schema != FormRecordSchema {
		return FormRecord{}, false
	}
	return fm.Meta, true
}

// SetFormRecord writes r as the "meta:" block into the front matter of md
// (Schema is set to FormRecordSchema), replacing an existing one and keeping
// every other key. A signature of the form is removed: it sealed the blank
// form, not this record.
func SetFormRecord(md string, r FormRecord) (string, error) {
	r.Schema = FormRecordSchema
	_, md = SplitSignature(md)
	b := frontMatterBlock(md)
	mapping := &yaml.Node{Kind: yaml.MappingNode}
	body := md
	if b.present {
		var root yaml.Node
		if err := yaml.Unmarshal([]byte(b.raw), &root); err != nil {
			return "", err
		}
		if len(root.Content) > 0 && root.Content[0].Kind == yaml.MappingNode {
			mapping = root.Content[0]
		}
		body = b.body
	}
	var meta yaml.Node
	if err := meta.Encode(r); err != nil {
		return "", err
	}
	kept := mapping.Content[:0:0]
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value != "meta" {
			kept = append(kept, mapping.Content[i], mapping.Content[i+1])
		}
	}
	mapping.Content = append(kept, &yaml.Node{Kind: yaml.ScalarNode, Value: "meta"}, &meta)
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	if err := enc.Encode(mapping); err != nil {
		return "", err
	}
	_ = enc.Close()
	return "---\n" + out.String() + "---\n\n" + strings.TrimLeft(body, "\r\n"), nil
}
