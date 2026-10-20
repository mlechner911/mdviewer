package mlcext

// Analysis is everything an editor needs to support the MLC Markdown
// Extension without re-implementing its rules: findings, field and hint
// positions, and the shared lists. Positions are 1-based lines and columns
// counted in characters (Unicode code points); End positions are exclusive.
type Analysis struct {
	// Detected: the document uses the extension. Plain Markdown gets an
	// empty analysis.
	Detected    bool         `json:"detected"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	Fields      []FieldSpan  `json:"fields"`
	Hints       []HintSpan   `json:"hints"`
	Lists       Lists        `json:"lists"`
	// Document is how the Sensor Monitor reads the document (front matter,
	// legacy header lines as a fallback), for a "how the server sees it"
	// display.
	Document DocumentInfo `json:"document"`
}

// DocumentInfo is what the Sensor Monitor takes from a document.
type DocumentInfo struct {
	RefNo         string   `json:"refNo"`
	Kind          string   `json:"kind"`
	KindLabel     string   `json:"kindLabel"`
	Version       string   `json:"version"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	EventKinds    []string `json:"eventKinds"`
	EventKindsSet bool     `json:"eventKindsSet"`
	Fields        int      `json:"fields"`
	PrefillFields int      `json:"prefillFields"`
	FieldErrors   int      `json:"fieldErrors"`
	// HeaderInBody: id or version come from the legacy header lines.
	HeaderInBody bool `json:"headerInBody"`
	// Signature of the file, nil when unsigned (not verified here).
	Signature *Signature `json:"signature"`
}

// FieldSpan locates one form field.
type FieldSpan struct {
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	EndLine int    `json:"endLine"`
	EndCol  int    `json:"endCol"`
	Name    string `json:"name"`
	Named   bool   `json:"named"`
	Kind    string `json:"kind"` // check | text | multiline
	Pattern string `json:"pattern"`
	// Prefill: the Sensor Monitor fills this field from an event.
	Prefill bool `json:"prefill"`
}

// HintSpan locates a tooltip "<!--* … -->" or fill hint "<!--! … -->".
type HintSpan struct {
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	EndLine int    `json:"endLine"`
	EndCol  int    `json:"endCol"`
	Kind    string `json:"kind"` // tooltip | fill-hint
}

// Lists are the shared vocabularies (for completion in editors).
type Lists struct {
	EventKinds      []string       `json:"eventKinds"`
	Prefill         []PrefillField `json:"prefill"`
	FrontMatterKeys []string       `json:"frontMatterKeys"`
}

// SharedLists returns the shared vocabularies.
func SharedLists() Lists {
	return Lists{EventKinds: EventKinds, Prefill: Prefill, FrontMatterKeys: FrontMatterKeys}
}

// Analyze checks md like Check with RequireHeader and also returns the
// positions of fields and hints. For plain Markdown (Detect is false) it
// returns only the lists.
func Analyze(md string) Analysis {
	a := Analysis{Diagnostics: []Diagnostic{}, Fields: []FieldSpan{}, Hints: []HintSpan{}, Lists: SharedLists()}
	if !Detect(md) {
		return a
	}
	a.Detected = true
	a.Diagnostics = Check(md, Options{RequireHeader: true})
	s := &scanner{src: md, section: "feld"}
	s.run()
	fm, _ := ParseFrontMatter(md)
	m := DocumentMeta(md)
	sig, _ := SplitSignature(md)
	d := DocumentInfo{
		RefNo: m.DocID, Kind: m.Kind, KindLabel: KindLabel(m.Kind), Version: m.Version, Title: m.Title,
		Description: fm.Description, EventKinds: fm.EventKinds, EventKindsSet: fm.EventKindsSet,
		Fields: len(s.fields), FieldErrors: len(s.errs),
		HeaderInBody: m.DocIDSource == "header" || m.VersionSource == "header",
		Signature:    sig,
	}
	if d.EventKinds == nil {
		d.EventKinds = []string{}
	}
	for _, f := range s.fields {
		if f.Named && IsPrefillField(f.Name) {
			d.PrefillFields++
		}
	}
	a.Document = d
	for _, f := range s.fields {
		el, ec := position(md, f.End)
		a.Fields = append(a.Fields, FieldSpan{
			Line: f.Line, Col: f.Col, EndLine: el, EndCol: ec,
			Name: f.Name, Named: f.Named, Kind: f.Kind.String(), Pattern: f.Pattern,
			Prefill: f.Named && IsPrefillField(f.Name),
		})
	}
	for _, h := range s.hints {
		a.Hints = append(a.Hints, HintSpan{Line: h.Line, Col: h.Col, EndLine: h.EndLine, EndCol: h.EndCol, Kind: h.Kind})
	}
	return a
}
