package mlcext

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile is the "Steckbrief" of a shipped template (front matter key
// "profile"): what it is for, when it appears, what the organisation must
// adapt before approving it, what the system fills in, and what it is
// oriented on. The Sensor Monitor shows it in its library and refuses to
// approve a template whose placeholders are still open.
//
//	profile:
//	  purpose: Was tun, wenn der Kühlschrank zu warm ist.
//	  appears: Bei Grenzwert-Ereignissen an Kühlgeräten.
//	  adapt:
//	    - name: verantwortlich
//	      note: Wer entscheidet über die Ware (Name oder Funktion).
//	  keep:
//	    - name: Ereignis-Nr.
//	      note: Füllt das System aus dem Ereignis.
//	  orientation:
//	    - ApBetrO § 2a (QM-System)
//
// A name in "adapt" may stand in the text as {{name}}: it is an open
// placeholder, reported as a warning while the template is a draft. Any
// other unknown {{…}} stays an error.
type Profile struct {
	Purpose     string        `yaml:"purpose" json:"purpose"`
	Appears     string        `yaml:"appears" json:"appears,omitempty"`
	Adapt       []ProfileItem `yaml:"adapt" json:"adapt,omitempty"`
	Keep        []ProfileItem `yaml:"keep" json:"keep,omitempty"`
	Orientation []string      `yaml:"orientation" json:"orientation,omitempty"`
}

// ProfileItem is one entry of "adapt" (a placeholder name) or "keep" (a
// value the system fills in, by its label).
type ProfileItem struct {
	Name string `yaml:"name" json:"name"`
	Note string `yaml:"note" json:"note"`
}

// ProfileKeys are the keys allowed in the "profile" block.
var ProfileKeys = []string{"purpose", "appears", "adapt", "keep", "orientation"}

// rePlaceholder: the name of a placeholder, as in {{verantwortlich}} —
// lowercase letters, digits and "_", starting with a letter.
var rePlaceholder = regexp.MustCompile(`^[a-z][a-z0-9_]{0,39}$`)

// IsPlaceholderName reports whether name may be declared in "adapt": the
// document variables ({{doc_id}}, {{version}}, {{title}}) are not
// placeholders, the system resolves them.
func IsPlaceholderName(name string) bool {
	return rePlaceholder.MatchString(name) && !isVariable(name)
}

// ParseProfile reads the "profile" block of md. ok is false without a
// block or when it cannot be read; the rest of the front matter is read
// separately, so a broken profile never hides doc_id or version.
func ParseProfile(md string) (p Profile, ok bool) {
	b := frontMatterBlock(md)
	if !b.present {
		return Profile{}, false
	}
	var v struct {
		Profile *Profile `yaml:"profile"`
	}
	if err := yaml.NewDecoder(bytes.NewReader([]byte(b.raw))).Decode(&v); err != nil || v.Profile == nil {
		return Profile{}, false
	}
	return *v.Profile, true
}

// Placeholders: the names declared in "adapt".
func (p Profile) Placeholders() []string {
	out := make([]string, 0, len(p.Adapt))
	for _, a := range p.Adapt {
		if n := strings.TrimSpace(a.Name); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// OpenPlaceholders lists the declared names that still stand in md as
// {{name}} (outside code and the signature), in the order of declared. md
// may be the stored text without front matter: the Sensor Monitor keeps
// the declared names with the document and asks before approving.
func OpenPlaceholders(md string, declared []string) []string {
	seen := map[string]bool{}
	for _, r := range varRefs(md) {
		seen[r.Name] = true
	}
	var out []string
	for _, n := range declared {
		if seen[n] && !slices.Contains(out, n) {
			out = append(out, n)
		}
	}
	return out
}

// checkProfile checks the "profile" block (node v, front matter starting at
// yamlLine) and returns the declared placeholder names.
func checkProfile(md string, v *yaml.Node, yamlLine int, add func(int, int, Severity, string, string)) map[string]bool {
	declared := map[string]bool{}
	at := func(n *yaml.Node) (int, int) { return n.Line + yamlLine - 1, n.Column }
	if v.Kind != yaml.MappingNode {
		l, c := at(v)
		add(l, c, Error, "profile-yaml", "`profile` (Steckbrief) muss ein Block aus `schlüssel: wert`-Zeilen sein: purpose, appears, adapt, keep, orientation.")
		return declared
	}
	hasPurpose := false
	for i := 0; i+1 < len(v.Content); i += 2 {
		k, val := v.Content[i], v.Content[i+1]
		l, c := at(k)
		switch k.Value {
		case "purpose":
			hasPurpose = val.Kind == yaml.ScalarNode && strings.TrimSpace(val.Value) != ""
			if !hasPurpose {
				add(l, c, Error, "profile-purpose", "`profile.purpose` ist ein Satz: wozu die Vorlage da ist.")
			}
		case "appears":
			if val.Kind != yaml.ScalarNode {
				add(l, c, Error, "profile-appears", "`profile.appears` ist ein Satz: wann und wo die Vorlage erscheint.")
			}
		case "orientation":
			if val.Kind != yaml.SequenceNode {
				add(l, c, Error, "profile-orientation", "`profile.orientation` ist eine Liste der Werke, an denen sich die Vorlage orientiert.")
			}
		case "adapt", "keep":
			if val.Kind != yaml.SequenceNode {
				add(l, c, Error, "profile-list", fmt.Sprintf("`profile.%s` ist eine Liste von Einträgen mit `name` und `note`.", k.Value))
				continue
			}
			for _, it := range val.Content {
				var e ProfileItem
				il, ic := at(it)
				if it.Kind != yaml.MappingNode || it.Decode(&e) != nil || strings.TrimSpace(e.Name) == "" {
					add(il, ic, Error, "profile-item", fmt.Sprintf("Eintrag in `profile.%s` braucht `name` und `note`.", k.Value))
					continue
				}
				if strings.TrimSpace(e.Note) == "" {
					add(il, ic, Warning, "profile-item-note", fmt.Sprintf("`%s` in `profile.%s` ohne `note`: Wer die Vorlage übernimmt, braucht die Erklärung.", e.Name, k.Value))
				}
				if k.Value == "keep" {
					continue
				}
				switch {
				case !IsPlaceholderName(e.Name):
					add(il, ic, Error, "profile-placeholder-name", fmt.Sprintf("Platzhalter `%s`: nur Kleinbuchstaben, Ziffern und `_`, beginnend mit einem Buchstaben; nicht doc_id, version oder title.", e.Name))
				case declared[e.Name]:
					add(il, ic, Error, "profile-placeholder-twice", fmt.Sprintf("Platzhalter `%s` steht zweimal in `profile.adapt`.", e.Name))
				default:
					declared[e.Name] = true
				}
			}
		default:
			add(l, c, Error, "profile-unknown-key", fmt.Sprintf("Unbekannter Schlüssel `profile.%s`; erlaubt sind %s.", k.Value, strings.Join(ProfileKeys, ", ")))
		}
	}
	if !hasPurpose && !slices.ContainsFunc(v.Content, func(n *yaml.Node) bool { return n.Value == "purpose" }) {
		l, c := at(v)
		add(l, c, Error, "profile-purpose", "`profile.purpose` fehlt: Ein Steckbrief sagt zuerst, wozu die Vorlage da ist.")
	}
	used := map[string]bool{}
	for _, r := range varRefs(md) {
		used[r.Name] = true
	}
	names := make([]string, 0, len(declared))
	for n := range declared {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, n := range names {
		if !used[n] {
			l, c := at(v)
			add(l, c, Warning, "profile-placeholder-unused", fmt.Sprintf("Platzhalter `%s` steht in `profile.adapt`, aber nicht als `{{%s}}` im Text.", n, n))
		}
	}
	return declared
}
