package mlcext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
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

func TestLoadSchemas(t *testing.T) {
	dir := t.TempDir()
	const good = "id: https://schemas.example/forms/x/v1\nfields:\n  a.b: { type: text }\n"
	for name, content := range map[string]string{
		"a.yaml":    good,
		"b.yml":     good,                                      // same ID
		"c.yaml":    "fields:\n  a.b: { type: text }\n",        // no ID
		"d.yaml":    "id: https://x.example/f/v1\nfields: [ a", // broken
		"notes.txt": "ignored",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, errs := LoadSchemas(dir)
	if _, ok := got["https://schemas.example/forms/x/v1"]; !ok || len(got) != 1 {
		t.Errorf("schemas: %v", got)
	}
	if len(errs) != 3 {
		t.Errorf("want 3 problems, got %v", errs)
	}
	if got, errs := LoadSchemas(filepath.Join(dir, "missing")); len(got) != 0 || errs != nil {
		t.Errorf("missing folder: %v %v", got, errs)
	}
}

func TestAnalyzeUsesSchemas(t *testing.T) {
	const id = "https://schemas.example/forms/x/v1"
	md := "---\ndoc_id: FB-1\nversion: \"1\"\nschema: " + id + "\n---\n[@#a.typo ____]\n"
	resolve := func(string) (FieldSchema, bool) {
		return FieldSchema{ID: id, Fields: map[string]SchemaField{"a.b": {Type: "text"}}}, true
	}
	found := func(a Analysis, code string) bool {
		for _, d := range a.Diagnostics {
			if d.Code == code {
				return true
			}
		}
		return false
	}
	if a := AnalyzeWithSchemas(md, resolve); !found(a, "field-unknown-property") {
		t.Errorf("with schemas: %+v", a.Diagnostics)
	}
	if a := Analyze(md); !found(a, "schema-unknown") {
		t.Errorf("without schemas: %+v", a.Diagnostics)
	}
	// Without doc_id and version the field alone makes it an MLC document.
	if a := AnalyzeWithSchemas("---\nschema: "+id+"\n---\n[@#a.typo ____]\n", resolve); !found(a, "field-unknown-property") {
		t.Errorf("no doc_id: %+v", a.Diagnostics)
	}
}

// The JSON schema is for editors; the Go code decides. They must not drift.
func TestJSONSchemaMatchesCode(t *testing.T) {
	data, err := os.ReadFile("mlc-field-schema.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Properties struct {
			ID struct {
				Pattern string `json:"pattern"`
			} `json:"id"`
			Fields struct {
				PatternProperties map[string]any `json:"patternProperties"`
			} `json:"fields"`
		} `json:"properties"`
		Defs struct {
			Field struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"field"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Properties.ID.Pattern != schemaID.String() {
		t.Errorf("id pattern: %q, code %q", doc.Properties.ID.Pattern, schemaID)
	}
	if _, ok := doc.Properties.Fields.PatternProperties[propertyName.String()]; !ok {
		t.Errorf("property pattern differs from %q: %v", propertyName, doc.Properties.Fields.PatternProperties)
	}
	var kind struct {
		Enum []string `json:"enum"`
	}
	if err := json.Unmarshal(doc.Defs.Field.Properties["type"], &kind); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(kind.Enum, SchemaFieldTypes) {
		t.Errorf("types: %v, code %v", kind.Enum, SchemaFieldTypes)
	}
}

func TestPatternAgainstSchemaType(t *testing.T) {
	head := "---\nschema: { fields: { a.text: { type: text }, a.date: { type: date }, a.time: { type: time }, a.at: { type: datetime } } }\n---\n"
	for field, warn := range map[string]bool{
		"[@#a.text TT.MM.JJJJ]":     true,
		"[@#a.text HH:MM]":          true,
		"[@#a.text ____]":           false,
		"[@#a.date TT.MM.JJJJ]":     false,
		"[@#a.date HH:MM]":          true,
		"[@#a.date ____]":           false,
		"[@#a.time HH:MM]":          false,
		"[@#a.at TT.MM.JJJJ HH:MM]": false,
		"[@#a.at TT.MM.JJJJ]":       true,
		"[@#a.text ABC-999]":        false,
	} {
		got := schemaCodes(head+field+"\n", Options{})
		if schemaHasCode(got, "field-type-pattern") != warn {
			t.Errorf("%s: warn %v, got %v", field, warn, got)
		}
	}
}

func TestAnalysisCarriesTheSchema(t *testing.T) {
	if a := Analyze(inlineSchema + "[@#patient.name ____]\n"); a.Schema == nil || a.Schema.Fields["patient.name"].Description["de"] != "Name" {
		t.Errorf("inline: %+v", a.Schema)
	}
	if a := Analyze("---\nschema: https://x.example/f/v1\n---\n[@#a.b ____]\n"); a.Schema != nil {
		t.Errorf("unknown: %+v", a.Schema)
	}
	if a := Analyze("---\nschema: nonsense\n---\n[@#a.b ____]\n"); a.Schema != nil {
		t.Errorf("invalid: %+v", a.Schema)
	}
}

func TestExampleIsRead(t *testing.T) {
	s, err := ParseFieldSchema([]byte("fields:\n  a.b: { type: text, example: { de: Kühlraum A } }\n"))
	if err != nil || s.Fields["a.b"].Example["de"] != "Kühlraum A" {
		t.Errorf("got %+v, %v", s, err)
	}
	if out, set := FillFields("[@#a.b TT.MM.JJJJ] [@#a.c ______]\n", map[string]string{"a.b": "16.10.1980", "a.c": "x"}); len(set) != 2 || !strings.Contains(out, `[@#a.b TT.MM.JJJJ=16\.10\.1980]`) {
		t.Errorf("fill: %q %v", out, set)
	}
}
