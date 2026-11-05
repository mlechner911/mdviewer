package mlcext

import (
	"strings"
	"testing"
)

const inlineSchema = `---
schema:
  fields:
    patient.name: { type: text, description: { de: Name, en: Name } }
    patient.birth_date: { type: date }
  legacy:
    name: patient.name
---
`

func schemaCodes(md string, opt Options) []string {
	var out []string
	for _, d := range Check(md, opt) {
		out = append(out, d.Code)
	}
	return out
}

func schemaHasCode(list []string, code string) bool {
	return strings.Contains(" "+strings.Join(list, " ")+" ", " "+code+" ")
}

func TestSchemaInDocument(t *testing.T) {
	ok := schemaCodes(inlineSchema+"[@#patient.name ____] [@#eigen ____]\n", Options{})
	if schemaHasCode(ok, "field-unknown-property") || schemaHasCode(ok, "frontmatter-unknown-key") {
		t.Errorf("known property rejected: %v", ok)
	}
	ds := Check(inlineSchema+"[@#patient.nmae ____]\n", Options{})
	found := false
	for _, d := range ds {
		if d.Code == "field-unknown-property" {
			found = d.Severity == Error && strings.Contains(d.Message, "`patient.name`")
		}
	}
	if !found {
		t.Errorf("no error with suggestion: %+v", ds)
	}
	if got := schemaCodes(inlineSchema+"[@#name ____]\n", Options{}); !schemaHasCode(got, "field-legacy") {
		t.Errorf("legacy name not reported: %v", got)
	}
}

func TestPropertyWithoutSchemaIsOnlyAHint(t *testing.T) {
	got := schemaCodes("[@#probe.name ____] [@#eigen ____]\n", Options{})
	if !schemaHasCode(got, "field-no-schema") || len(got) != 1 {
		t.Errorf("got %v", got)
	}
}

func TestSchemaByName(t *testing.T) {
	const id = "https://schemas.example/forms/x/v1"
	md := "---\nschema: " + id + "\n---\n[@#a.typo ____]\n"
	if got := schemaCodes(md, Options{}); !schemaHasCode(got, "schema-unknown") || schemaHasCode(got, "field-unknown-property") {
		t.Errorf("unresolved: %v", got)
	}
	asked := ""
	resolve := func(s string) (FieldSchema, bool) {
		asked = s
		return FieldSchema{ID: id, Fields: map[string]SchemaField{"a.b": {Type: "text"}}}, true
	}
	if got := schemaCodes(md, Options{Schemas: resolve}); !schemaHasCode(got, "field-unknown-property") || asked != id {
		t.Errorf("resolved: %v, asked %q", got, asked)
	}
}

func TestInvalidSchemas(t *testing.T) {
	for name, fm := range map[string]string{
		"not https":   "schema: http://x.example/f/v1",
		"no version":  "schema: https://x.example/forms/x",
		"list":        "schema: [a]",
		"no fields":   "schema: { title: { de: T } }",
		"bad name":    "schema: { fields: { Probe.name: { type: text } } }",
		"no dot":      "schema: { fields: { name: { type: text } } }",
		"bad type":    "schema: { fields: { a.b: { type: datum } } }",
		"both levels": "schema: { fields: { a.b: { type: text }, a.b.c: { type: text } } }",
		"legacy":      "schema: { fields: { a.b: { type: text } }, legacy: { x: a.c } }",
	} {
		if got := schemaCodes("---\n"+fm+"\n---\nText\n", Options{}); !schemaHasCode(got, "frontmatter-schema") {
			t.Errorf("%s: got %v", name, got)
		}
	}
}

func TestNoSchemaNoFindings(t *testing.T) {
	if got := schemaCodes("[@#eigen ____]\n", Options{}); len(got) != 0 {
		t.Errorf("got %v", got)
	}
}
