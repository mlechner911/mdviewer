package mlcext

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Templates for new documents: data, not code. A template is a schema
// "<id>.json" (name, input fields, output file name) plus a Markdown body
// with {{placeholders}}. New kinds, fields or choice lists need a new file,
// never a new program. The JSON Schema of the schema file is
// mlc-template.schema.json; editors validate templates with it.
//
// Built-in templates ship embedded as a starting set: the documents of the
// MLC extension (SOP, VA, form, info) and general ones (README, CHANGELOG,
// empty document, minutes). SeedTemplates copies them into a user folder
// where they can be changed without a new release; a template of the same
// id in a later folder replaces an earlier one.
//
// Documents of the MLC extension keep id, version and title in the front
// matter: fields with "target": "frontmatter" are written there, and the
// body uses the document variables {{doc_id}}, {{version}}, {{title}},
// which stay in the file (see Variables). Other placeholders are replaced
// when the document is created.
//
//go:embed templates/*.json templates/*.md
var builtinTemplates embed.FS

// TemplateSchema is one template file.
type TemplateSchema struct {
	Schema      string `json:"$schema,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Body        string `json:"body"`           // Markdown file next to the schema
	File        string `json:"file,omitempty"` // output file name, with placeholders
	// Category: "mlc" (document of the MLC extension) or "general";
	// default "mlc" when a field writes doc_id or is a refno, else
	// "general".
	Category string          `json:"category,omitempty"`
	Fields   []TemplateField `json:"fields"`
}

// TemplateField is one input of a template.
//
//	text, multiline  free text
//	number           optional min, max, step
//	date             TT.MM.JJJJ (format "JJJJ-MM-TT" for ISO)
//	time             HH:MM ("hour" is accepted as the same type)
//	select           one of options
//	multiselect      several of options
//	refno            document ID with prefix ("FB") and an area field;
//	                 the editor proposes the next free number
//	docref           reference to other documents of the document folder
//	                 (optional prefix filter, "multiple" for several);
//	                 {{id}} writes the IDs, {{id|ref}} "ID (Titel)"
//
// options is a list of strings or the name of a shared list
// ("event_kinds").
type TemplateField struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required,omitempty"`
	Default     any      `json:"default,omitempty"`
	Help        string   `json:"help,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Options     any      `json:"options,omitempty"`
	Prefix      string   `json:"prefix,omitempty"`
	Area        string   `json:"area,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	Step        *float64 `json:"step,omitempty"`
	Format      string   `json:"format,omitempty"`
	Multiple    bool     `json:"multiple,omitempty"`
	// Target "frontmatter": the value goes into the front matter under the
	// field's id (doc_id, version, title, description, event_kinds for
	// documents of the MLC extension) instead of a placeholder.
	Target string `json:"target,omitempty"`
}

// FieldTypes are the known field types.
var FieldTypes = []string{"text", "multiline", "number", "date", "time", "select", "multiselect", "refno", "docref"}

// TemplateInfo is a loaded template, ready for a form: options resolved,
// problems listed (a template with problems cannot be used).
type TemplateInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Source      string          `json:"source"` // "builtin" or the folder
	Category    string          `json:"category"`
	Fields      []TemplateInput `json:"fields"`
	Problems    []string        `json:"problems"`
	// Warnings: the template works but should be updated (old format).
	Warnings []string `json:"warnings"`

	schema TemplateSchema
	body   string
	titles map[string]string // docref: RefNo → title, from WithDocuments
}

// TemplateInput is a field as a form needs it.
type TemplateInput struct {
	TemplateField
	OptionList []string `json:"optionList"`
	// OptionLabels are display texts parallel to OptionList (docref:
	// "SOP-QM-006 · CAPA-Prozess"); empty means the options themselves.
	OptionLabels []string `json:"optionLabels"`
	DefaultStr   string   `json:"defaultValue"`
}

// DocumentRef is a document a docref field can point to.
type DocumentRef struct {
	RefNo string `json:"refNo"`
	Title string `json:"title"`
}

// WithDocuments fills the choices of docref fields from docs (documents of
// the document folder; those without a RefNo are skipped).
func (t TemplateInfo) WithDocuments(docs []DocumentRef) TemplateInfo {
	t.titles = map[string]string{}
	for _, d := range docs {
		if d.RefNo != "" {
			t.titles[d.RefNo] = d.Title
		}
	}
	fields := make([]TemplateInput, len(t.Fields))
	for i, f := range t.Fields {
		if f.Type == "docref" {
			f.OptionList, f.OptionLabels = []string{}, []string{}
			for _, d := range docs {
				if d.RefNo == "" || f.Prefix != "" && !strings.EqualFold(strings.SplitN(d.RefNo, "-", 2)[0], f.Prefix) {
					continue
				}
				f.OptionList = append(f.OptionList, d.RefNo)
				f.OptionLabels = append(f.OptionLabels, strings.TrimSpace(d.RefNo+" · "+d.Title))
			}
		}
		fields[i] = f
	}
	t.Fields = fields
	return t
}

var (
	fieldIDRe     = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	placeholderRe = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*(?:\|\s*([a-z]+)\s*)?\}\}`)
	sharedLists   = map[string]func() []string{"event_kinds": func() []string { return EventKinds }}
)

// LoadTemplates returns the built-in templates and those in dirs; a later
// template with the same id (file name without .json) replaces an earlier
// one. Sorted by name.
func LoadTemplates(dirs ...string) []TemplateInfo {
	return LoadTemplatesFrom(true, dirs...)
}

// LoadTemplatesFrom is LoadTemplates with the built-in set optional: a
// seeded user folder (SeedTemplates) replaces it, so that a template the
// user deleted stays deleted.
func LoadTemplatesFrom(builtin bool, dirs ...string) []TemplateInfo {
	byID := map[string]TemplateInfo{}
	if builtin {
		sub, _ := fs.Sub(builtinTemplates, "templates")
		for _, t := range loadFrom(sub, "builtin") {
			byID[t.ID] = t
		}
	}
	for _, d := range dirs {
		if d == "" {
			continue
		}
		for _, t := range loadFrom(os.DirFS(d), d) {
			byID[t.ID] = t
		}
	}
	out := make([]TemplateInfo, 0, len(byID))
	for _, t := range byID {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category == "mlc" // MLC documents first
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func loadFrom(fsys fs.FS, source string) []TemplateInfo {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil
	}
	var out []TemplateInfo
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".json" || strings.HasSuffix(name, ".schema.json") || strings.HasPrefix(name, ".") {
			continue
		}
		t := TemplateInfo{ID: strings.TrimSuffix(name, ".json"), Source: source, Problems: []string{}, Warnings: []string{}, Category: "general"}
		raw, err := fs.ReadFile(fsys, name)
		if err == nil {
			dec := json.NewDecoder(strings.NewReader(string(raw)))
			dec.DisallowUnknownFields()
			err = dec.Decode(&t.schema)
		}
		if err != nil {
			t.Name = t.ID
			t.Problems = append(t.Problems, fmt.Sprintf("%s: kein gültiges Vorlagen-Schema: %v", name, err))
			out = append(out, t)
			continue
		}
		t.Name, t.Description = t.schema.Name, t.schema.Description
		if t.Name == "" {
			t.Name = t.ID
		}
		if t.schema.Body != "" {
			if b, err := fs.ReadFile(fsys, filepath.ToSlash(t.schema.Body)); err == nil {
				t.body = string(b)
			} else {
				t.Problems = append(t.Problems, fmt.Sprintf("Text %q nicht gefunden", t.schema.Body))
			}
		}
		t.Category = templateCategory(t.schema)
		t.Problems = append(t.Problems, validateSchema(t.schema, t.body)...)
		t.Warnings = append(t.Warnings, templateWarnings(t.schema, t.body, t.Category)...)
		for _, f := range t.schema.Fields {
			in := TemplateInput{TemplateField: f, OptionList: optionList(f.Options), OptionLabels: []string{}, DefaultStr: defaultString(f.Default)}
			if in.Type == "hour" {
				in.Type = "time"
			}
			t.Fields = append(t.Fields, in)
		}
		out = append(out, t)
	}
	return out
}

func optionList(o any) []string {
	switch v := o.(type) {
	case string:
		if f, ok := sharedLists[v]; ok {
			return append([]string(nil), f()...)
		}
	case []any:
		out := make([]string, 0, len(v))
		for _, x := range v {
			out = append(out, fmt.Sprint(x))
		}
		return out
	}
	return []string{}
}

func defaultString(d any) string {
	switch v := d.(type) {
	case nil:
		return ""
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case []any:
		parts := make([]string, len(v))
		for i, x := range v {
			parts[i] = fmt.Sprint(x)
		}
		return strings.Join(parts, ",")
	}
	return fmt.Sprint(d)
}

// validateSchema reports problems that make a template unusable.
func validateSchema(s TemplateSchema, body string) []string {
	var p []string
	if strings.TrimSpace(s.Name) == "" {
		p = append(p, "„name“ fehlt")
	}
	if s.Body == "" {
		p = append(p, "„body“ fehlt (Markdown-Datei mit dem Dokumenttext)")
	}
	ids := map[string]TemplateField{}
	for i, f := range s.Fields {
		where := fmt.Sprintf("Feld %d (%s)", i+1, f.ID)
		if !fieldIDRe.MatchString(f.ID) {
			p = append(p, where+": „id“ nur Kleinbuchstaben, Ziffern und _ , beginnend mit einem Buchstaben")
		}
		if _, dup := ids[f.ID]; dup {
			p = append(p, where+": „id“ kommt zweimal vor")
		}
		ids[f.ID] = f
		typ := f.Type
		if typ == "hour" {
			typ = "time"
		}
		if !contains(FieldTypes, typ) {
			p = append(p, fmt.Sprintf("%s: unbekannter Typ %q; bekannt: %s", where, f.Type, strings.Join(FieldTypes, ", ")))
		}
		if strings.TrimSpace(f.Label) == "" {
			p = append(p, where+": „label“ fehlt")
		}
		if (typ == "select" || typ == "multiselect") && len(optionList(f.Options)) == 0 {
			p = append(p, where+": „options“ fehlt (Liste oder Name einer gemeinsamen Liste wie „event_kinds“)")
		}
		if typ == "refno" && strings.TrimSpace(f.Prefix) == "" {
			p = append(p, where+": „prefix“ fehlt (z. B. „FB“)")
		}
		if typ == "date" && f.Format != "" && f.Format != "TT.MM.JJJJ" && f.Format != "JJJJ-MM-TT" {
			p = append(p, where+": „format“ ist TT.MM.JJJJ oder JJJJ-MM-TT")
		}
		if f.Target != "" && f.Target != "frontmatter" {
			p = append(p, where+": „target“ ist „frontmatter“ oder fehlt")
		}
	}
	mlc := templateCategory(s) == "mlc"
	if s.Category != "" && s.Category != "mlc" && s.Category != "general" {
		p = append(p, "„category“ ist „mlc“ oder „general“")
	}
	if mlc {
		for _, f := range s.Fields {
			switch {
			case f.Target == "frontmatter" && !slices.Contains(FrontMatterKeys, f.ID):
				p = append(p, fmt.Sprintf("Feld %s: kein erlaubter Front-Matter-Schlüssel (erlaubt: %s)", f.ID, strings.Join(FrontMatterKeys, ", ")))
			case f.Target != "frontmatter" && isVariable(f.ID):
				p = append(p, fmt.Sprintf("Feld %s: braucht „target“: „frontmatter“, denn {{%s}} ist eine Variable des Dokuments", f.ID, f.ID))
			}
		}
	}
	for _, f := range s.Fields {
		if f.Area != "" {
			if _, ok := ids[f.Area]; !ok {
				p = append(p, fmt.Sprintf("Feld %s: „area“ verweist auf unbekanntes Feld %q", f.ID, f.Area))
			}
		}
	}
	for i, src := range []string{body, s.File} {
		for _, m := range placeholderRe.FindAllStringSubmatch(src, -1) {
			if mlc && i == 0 && isVariable(m[1]) && m[2] == "" {
				continue // document variable, stays in the file
			}
			if _, ok := ids[m[1]]; !ok {
				p = append(p, fmt.Sprintf("Platzhalter {{%s}} hat kein Feld", m[1]))
			}
			if m[2] != "" && m[2] != "yaml" && m[2] != "list" && m[2] != "ref" {
				p = append(p, fmt.Sprintf("Platzhalter {{%s|%s}}: Filter ist yaml, list oder ref", m[1], m[2]))
			}
		}
	}
	return p
}

// templateCategory is the schema's category, or the default from its
// fields: "mlc" when one writes doc_id or is a refno.
func templateCategory(s TemplateSchema) string {
	if s.Category != "" {
		return s.Category
	}
	for _, f := range s.Fields {
		if f.Type == "refno" || f.ID == "doc_id" || f.ID == "event_kinds" {
			return "mlc"
		}
	}
	return "general"
}

// templateWarnings: usable, but in the old form (id and version as a
// header line in the text instead of the front matter).
func templateWarnings(s TemplateSchema, body, category string) []string {
	if category != "mlc" {
		return nil
	}
	for _, f := range s.Fields {
		if f.Target == "frontmatter" {
			return nil
		}
	}
	if strings.Contains(body, "**Dokumenten-ID:**") || strings.Contains(body, "**Version:**") {
		return []string{"altes Vorlagenformat: Kennung und Version stehen als Kopfzeile im Text; besser Felder mit „target“: „frontmatter“ (doc_id, version, title) und im Text {{doc_id}}, {{version}}, {{title}}"}
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// RenderResult is a document created from a template.
type RenderResult struct {
	Content  string `json:"content"`
	FileName string `json:"fileName"`
	// FieldErrors maps a field id to what is wrong with its value.
	FieldErrors map[string]string `json:"fieldErrors"`
	// Diagnostics of the created document (Check); errors block creation
	// in the editor.
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Render fills t with values (strings; multiselect as a list or a
// comma-separated string) and checks the result.
func (t TemplateInfo) Render(values map[string]any) RenderResult {
	res := RenderResult{FieldErrors: map[string]string{}, Diagnostics: []Diagnostic{}}
	if len(t.Problems) > 0 {
		res.FieldErrors[""] = "Vorlage fehlerhaft: " + strings.Join(t.Problems, "; ")
		return res
	}
	filled := map[string][]string{}
	for _, f := range t.Fields {
		vals, msg := normalize(f, values[f.ID])
		if msg != "" {
			res.FieldErrors[f.ID] = msg
		}
		filled[f.ID] = vals
	}
	if len(res.FieldErrors) > 0 {
		return res
	}
	mlc := t.Category == "mlc"
	subst := func(src string, keepVariables bool) string {
		return placeholderRe.ReplaceAllStringFunc(src, func(m string) string {
			sm := placeholderRe.FindStringSubmatch(m)
			if keepVariables && isVariable(sm[1]) && sm[2] == "" {
				return m // document variable: replaced when rendered, not now
			}
			v, ok := filled[sm[1]]
			if !ok {
				return m
			}
			switch sm[2] {
			case "ref":
				items := make([]string, len(v))
				for i, x := range v {
					items[i] = x
					if title := t.titles[x]; title != "" {
						items[i] = x + " (" + title + ")"
					}
				}
				return strings.Join(items, ", ")
			case "list":
				items := make([]string, len(v))
				for i, x := range v {
					items[i] = yamlScalar(x)
				}
				return "[" + strings.Join(items, ", ") + "]"
			case "yaml":
				return yamlScalar(strings.Join(v, ", "))
			}
			return strings.Join(v, ", ")
		})
	}
	res.Content = withFrontMatter(subst(t.body, mlc), t.frontMatterLines(filled, mlc))
	if t.schema.File != "" {
		res.FileName = sanitizeFileName(subst(t.schema.File, false))
	}
	res.Diagnostics = Check(res.Content, Options{RequireHeader: Detect(res.Content)})
	return res
}

// frontMatterLines are "key: value" lines of the fields with target
// "frontmatter"; for MLC documents in the order of FrontMatterKeys. Empty
// optional values are left out, except lists ("event_kinds: []" means
// "applies always").
func (t TemplateInfo) frontMatterLines(filled map[string][]string, mlc bool) []string {
	var fields []TemplateInput
	for _, f := range t.Fields {
		if f.Target == "frontmatter" {
			fields = append(fields, f)
		}
	}
	if mlc {
		sort.SliceStable(fields, func(i, j int) bool {
			return slices.Index(FrontMatterKeys, fields[i].ID) < slices.Index(FrontMatterKeys, fields[j].ID)
		})
	}
	var lines []string
	for _, f := range fields {
		v := filled[f.ID]
		if multi(f) {
			items := make([]string, len(v))
			for i, x := range v {
				items[i] = yamlScalar(x)
			}
			lines = append(lines, f.ID+": ["+strings.Join(items, ", ")+"]")
			continue
		}
		s := strings.Join(v, ", ")
		if s == "" {
			continue
		}
		lines = append(lines, f.ID+": "+yamlScalar(s))
	}
	return lines
}

// withFrontMatter puts lines at the top of body's front matter, or starts
// one.
func withFrontMatter(body string, lines []string) string {
	if len(lines) == 0 {
		return body
	}
	block := strings.Join(lines, "\n") + "\n"
	if b := frontMatterBlock(body); b.present {
		at := lineOffset(body, b.yamlLine-1)
		return body[:at] + block + body[at:]
	}
	return "---\n" + block + "---\n" + strings.TrimLeft(body, "\r\n")
}

// normalize validates one value and returns it as a list of strings.
func normalize(f TemplateInput, raw any) ([]string, string) {
	vals := collectValues(f, raw)
	if len(vals) == 0 {
		switch {
		case f.Required:
			return nil, "Pflichtfeld"
		case multi(f):
			return []string{}, ""
		}
		return []string{""}, ""
	}
	if multi(f) {
		for _, x := range vals {
			if !contains(f.OptionList, x) {
				return nil, fmt.Sprintf("%q ist kein Wert aus der Liste", x)
			}
		}
		return vals, ""
	}
	if len(vals) > 1 {
		return nil, "nur ein Wert erlaubt"
	}
	v, msg := checkValue(f, vals[0])
	if msg != "" {
		return nil, msg
	}
	return []string{v}, ""
}

// multi: the field takes several values.
func multi(f TemplateInput) bool {
	return f.Type == "multiselect" || f.Type == "docref" && f.Multiple
}

// collectValues turns a form value (string, number or list) into trimmed
// strings; multiselect also accepts "A, B". Except in multiline fields,
// whitespace runs collapse to one space.
func collectValues(f TemplateInput, raw any) []string {
	var parts []string
	switch v := raw.(type) {
	case nil:
	case []any:
		for _, x := range v {
			parts = append(parts, fmt.Sprint(x))
		}
	case []string:
		parts = v
	case float64:
		parts = []string{strconv.FormatFloat(v, 'f', -1, 64)}
	default:
		s := fmt.Sprint(v)
		if multi(f) {
			parts = strings.Split(s, ",")
		} else {
			parts = []string{s}
		}
	}
	var vals []string
	for _, p := range parts {
		if f.Type != "multiline" {
			p = strings.Join(strings.Fields(p), " ")
		}
		if p = strings.TrimSpace(p); p != "" {
			vals = append(vals, p)
		}
	}
	return vals
}

// checkValue validates and formats a single value of its field type.
func checkValue(f TemplateInput, v string) (string, string) {
	switch f.Type {
	case "number":
		return checkNumber(f, v)
	case "date":
		d, ok := parseDate(v)
		if !ok {
			return "", "Datum als TT.MM.JJJJ"
		}
		if f.Format == "JJJJ-MM-TT" {
			return d.Format("2006-01-02"), ""
		}
		return d.Format("02.01.2006"), ""
	case "time":
		tm, err := time.Parse("15:04", v)
		if err != nil {
			return "", "Uhrzeit als HH:MM"
		}
		return tm.Format("15:04"), ""
	case "select", "docref":
		if !contains(f.OptionList, v) {
			return "", "kein Wert aus der Liste"
		}
	case "refno":
		v = strings.ToUpper(v)
		if m := refNoRe.FindStringSubmatch(v); m == nil || !strings.EqualFold(m[1], f.Prefix) {
			return "", fmt.Sprintf("Kennung als %s-BEREICH-NUMMER, z. B. %s-QM-001", f.Prefix, strings.ToUpper(f.Prefix))
		}
	}
	return v, ""
}

// checkNumber checks range and step and writes the German decimal comma,
// as in the documents.
func checkNumber(f TemplateInput, v string) (string, string) {
	n, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64)
	if err != nil {
		return "", "keine Zahl"
	}
	if f.Min != nil && n < *f.Min || f.Max != nil && n > *f.Max {
		return "", fmt.Sprintf("außerhalb von %s bis %s", fmtNum(f.Min), fmtNum(f.Max))
	}
	if f.Step != nil && *f.Step > 0 && f.Min != nil {
		if q := (n - *f.Min) / *f.Step; math.Abs(q-math.Round(q)) > 1e-9 {
			return "", fmt.Sprintf("nur in Schritten von %s", fmtNum(f.Step))
		}
	}
	return strings.ReplaceAll(strconv.FormatFloat(n, 'f', -1, 64), ".", ","), ""
}

func fmtNum(p *float64) string {
	if p == nil {
		return "…"
	}
	return strconv.FormatFloat(*p, 'f', -1, 64)
}

func parseDate(s string) (time.Time, bool) {
	for _, layout := range []string{"02.01.2006", "2.1.2006", "2006-01-02"} {
		if d, err := time.Parse(layout, s); err == nil {
			return d, true
		}
	}
	return time.Time{}, false
}

func yamlScalar(s string) string {
	b, err := yaml.Marshal(s)
	if err != nil {
		return `""`
	}
	return strings.TrimSpace(string(b))
}

var badFileChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func sanitizeFileName(s string) string {
	s = strings.TrimSpace(badFileChars.ReplaceAllString(s, "_"))
	if s == "" {
		return ""
	}
	if !strings.EqualFold(filepath.Ext(s), ".md") {
		s += ".md" // also "Messung 05.03.2026": ".2026" is no extension
	}
	return s
}
