package mlcext

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// FieldSchema describes the properties a form may use: the field names with
// a dot. It comes from the document ("schema:" with "fields:") or, by its ID,
// from whoever checks the document (Options.Schemas). It describes and grants
// nothing: who fills in values decides that from their own list.
type FieldSchema struct {
	// ID is an identifier, never fetched: an https address ending in a
	// version ("https://schemas.example/forms/x/v1"). A published schema only
	// grows, because a signature covers the ID and not the content. Empty for
	// a schema written in the document.
	ID     string                 `yaml:"id,omitempty" json:"id,omitempty"`
	Title  map[string]string      `yaml:"title,omitempty" json:"title,omitempty"`
	Fields map[string]SchemaField `yaml:"fields" json:"fields"`
	// Legacy maps old names without a dot to the property replacing them.
	// They stay valid, since released documents are signed.
	Legacy map[string]string `yaml:"legacy,omitempty" json:"legacy,omitempty"`
}

// SchemaField is one property. Names are not translated, descriptions are.
type SchemaField struct {
	// Type is one of SchemaFieldTypes.
	Type string `yaml:"type" json:"type"`
	// Requires says when a value exists; its words belong to the schema's
	// owner, so they are not checked here.
	Requires    string            `yaml:"requires,omitempty" json:"requires,omitempty"`
	Description map[string]string `yaml:"description,omitempty" json:"description,omitempty"`
	// Fixed: the value cannot be changed by the person filling in the form.
	Fixed bool `yaml:"fixed,omitempty" json:"fixed,omitempty"`
	// Right is what the filling tool requires for a value; empty: nothing.
	Right string `yaml:"right,omitempty" json:"right,omitempty"`
}

// SchemaFieldTypes are the known types of a property.
var SchemaFieldTypes = []string{"text", "date", "time", "datetime", "number"}

var schemaID = regexp.MustCompile(`^https://[^/\s]+(/\S*)?/v[0-9]+$`)

// ParseFieldSchema reads a schema file and validates it.
func ParseFieldSchema(data []byte) (FieldSchema, error) {
	var s FieldSchema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return FieldSchema{}, fmt.Errorf("schema lesen: %w", err)
	}
	return s, s.Validate()
}

// Validate reports the first problem of s, in a stable order. Every property
// must fit the grammar and none may be a level of another, because a PDF field
// is either a value or a level.
func (s FieldSchema) Validate() error {
	if s.ID != "" && !schemaID.MatchString(s.ID) {
		return fmt.Errorf("Schema-Kennung %q: erwartet wird eine https-Adresse, die mit der Version endet (…/v1).", s.ID)
	}
	if len(s.Fields) == 0 {
		return errors.New("Das Schema hat keine `fields`.")
	}
	names := s.names()
	for i, name := range names {
		if !propertyName.MatchString(name) {
			return fmt.Errorf("Schema-Eigenschaft %q: erwartet werden kleingeschriebene Wörter mit Punkt wie `probe.name`.", name)
		}
		if !slices.Contains(SchemaFieldTypes, s.Fields[name].Type) {
			return fmt.Errorf("Schema-Eigenschaft `%s`: unbekannter Typ %q; bekannt sind %s.", name, s.Fields[name].Type, strings.Join(SchemaFieldTypes, ", "))
		}
		// Sorted, so a level comes right before the names below it.
		if i+1 < len(names) && isFieldPrefix(name, names[i+1]) {
			return fmt.Errorf("Schema-Eigenschaften `%s` und `%s`: Ein Name kann nicht zugleich Feld und Ebene eines anderen sein.", name, names[i+1])
		}
	}
	for _, old := range sortedKeys(s.Legacy) {
		if _, ok := s.Fields[s.Legacy[old]]; !ok {
			return fmt.Errorf("Schema, `legacy`: `%s` verweist auf die unbekannte Eigenschaft `%s`.", old, s.Legacy[old])
		}
	}
	return nil
}

// LoadSchemas reads the schema files (*.yaml, *.yml) of dir by their ID, in
// name order. A missing folder is no error. A bad file, one without an ID or
// a second file with the same ID is left out and reported; the others still
// count.
func LoadSchemas(dir string) (map[string]FieldSchema, []error) {
	out := map[string]FieldSchema{}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return out, nil
	}
	if err != nil {
		return out, []error{fmt.Errorf("schemas lesen: %w", err)}
	}
	var errs []error
	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if e.IsDir() || (ext != ".yaml" && ext != ".yml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e.Name(), err))
			continue
		}
		s, err := ParseFieldSchema(data)
		switch _, seen := out[s.ID]; {
		case err != nil:
			errs = append(errs, fmt.Errorf("%s: %w", e.Name(), err))
		case s.ID == "":
			errs = append(errs, fmt.Errorf("%s: Die Datei hat keine `id`.", e.Name()))
		case seen:
			errs = append(errs, fmt.Errorf("%s: Die Kennung %s gibt es schon in einer anderen Datei.", e.Name(), s.ID))
		default:
			out[s.ID] = s
		}
	}
	return out, errs
}

func (s FieldSchema) names() []string { return sortedKeys(s.Fields) }

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// SchemaRef is the "schema:" key of a document: an ID, or a schema written in
// the document. Err says why the key cannot be used.
type SchemaRef struct {
	Present bool
	ID      string
	Inline  *FieldSchema
	Err     error
}

// DocumentSchema reads the "schema:" key of md.
func DocumentSchema(md string) SchemaRef {
	b := frontMatterBlock(md)
	if !b.present {
		return SchemaRef{}
	}
	var node yaml.Node
	if yaml.Unmarshal([]byte(b.raw), &node) != nil || len(node.Content) == 0 {
		return SchemaRef{}
	}
	m := node.Content[0]
	if m.Kind != yaml.MappingNode {
		return SchemaRef{}
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == "schema" {
			return schemaRefOf(m.Content[i+1])
		}
	}
	return SchemaRef{}
}

func schemaRefOf(v *yaml.Node) SchemaRef {
	r := SchemaRef{Present: true}
	switch v.Kind {
	case yaml.ScalarNode:
		r.ID = v.Value
		if !schemaID.MatchString(v.Value) {
			r.Err = fmt.Errorf("Schema-Kennung %q: erwartet wird eine https-Adresse, die mit der Version endet (…/v1).", v.Value)
		}
	case yaml.MappingNode:
		var s FieldSchema
		if err := v.Decode(&s); err != nil {
			r.Err = fmt.Errorf("Das Schema im Dokument lässt sich nicht lesen: %w", err)
			return r
		}
		r.Inline = &s
		r.Err = s.Validate()
	default:
		r.Err = errors.New("`schema` ist eine https-Adresse oder ein Block mit `fields`.")
	}
	return r
}

// checkSchema reports what the schema says about the fields of a document.
// line is the line of the "schema:" key.
func checkSchema(md string, line int, fields []Field, opt Options, add func(int, int, Severity, string, string)) {
	ref := DocumentSchema(md)
	if ref.Err != nil {
		add(line, 1, Error, "frontmatter-schema", ref.Err.Error())
		return
	}
	schema := ref.Inline
	if schema == nil && ref.Present && opt.Schemas != nil {
		if s, ok := opt.Schemas(ref.ID); ok {
			schema = &s
		}
	}
	if schema == nil && ref.Present {
		add(line, 1, Info, "schema-unknown", fmt.Sprintf("Schema `%s` ist hier nicht bekannt; die Felder mit Punkt werden nicht geprüft.", ref.ID))
		return
	}
	for _, f := range fields {
		if !f.Named {
			continue
		}
		if !strings.Contains(f.Name, ".") {
			if schema != nil && schema.Legacy[f.Name] != "" {
				add(f.Line, f.Col, Info, "field-legacy", fmt.Sprintf("Feld `%s` ist die alte Form von `%s`; es bleibt gültig.", f.Name, schema.Legacy[f.Name]))
			}
			continue
		}
		if schema == nil {
			add(f.Line, f.Col, Info, "field-no-schema", fmt.Sprintf("Feld `%s`: Das Dokument nennt kein Schema, die Eigenschaft wird nicht geprüft.", f.Name))
			continue
		}
		if _, ok := schema.Fields[f.Name]; ok {
			continue
		}
		msg := fmt.Sprintf("Feld `%s` steht nicht im Schema.", f.Name)
		if near := nearestName(f.Name, schema.names()); near != "" {
			msg += fmt.Sprintf(" Gemeint ist vielleicht `%s`.", near)
		}
		add(f.Line, f.Col, Error, "field-unknown-property", msg)
	}
}

// nearestName is the name closest to name, if it is close enough to be a typo.
func nearestName(name string, names []string) string {
	best, bestDist := "", max(len(name)/3, 2)+1
	for _, n := range names {
		if d := editDistance(name, n); d < bestDist {
			best, bestDist = n, d
		}
	}
	return best
}

// editDistance is the Levenshtein distance in bytes; names are ASCII.
func editDistance(a, b string) int {
	row := make([]int, len(b)+1)
	for j := range row {
		row[j] = j
	}
	for i := 1; i <= len(a); i++ {
		prev := row[0]
		row[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			prev, row[j] = row[j], min(row[j]+1, row[j-1]+1, prev+cost)
		}
	}
	return row[len(b)]
}
