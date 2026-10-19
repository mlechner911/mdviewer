package mlcext

// EventKinds are the event kinds a document may name in its front matter
// (`event_kinds: [OUT_OF_RANGE, NO_DATA]`): the document is shown first for
// these events. Single source for MarkSafe and the Sensor Monitor; add a new
// kind here first, then the Sensor Monitor migration.
var EventKinds = []string{
	"OUT_OF_RANGE",
	"DEVIATION",
	"NO_DATA",
	"CALIBRATION_DUE",
	"CERT_EXPIRING",
	"LOW_BATTERY",
}

// IsEventKind reports whether k is a known event kind.
func IsEventKind(k string) bool {
	for _, e := range EventKinds {
		if e == k {
			return true
		}
	}
	return false
}

// PrefillFields are field names the Sensor Monitor fills from an event
// ("[@#ort ____]" gets "Zone › Gerät › Probe"). Using them is intended, so
// the check reports them as information, not as a problem.
var PrefillFields = []string{
	"ereignis",
	"entdeckt_datum",
	"entdeckt_uhrzeit",
	"ort",
	"betroffen",
	"vorfall",
	"sofortmassnahmen",
}

// IsPrefillField reports whether name is filled from an event.
func IsPrefillField(name string) bool {
	for _, p := range PrefillFields {
		if p == name {
			return true
		}
	}
	return false
}

// FrontMatterKeys are the keys a document may carry in its front matter.
// Everything else is rejected: validity, review dates and status are set by
// the approval in the Sensor Monitor, never by the author.
var FrontMatterKeys = []string{"title", "version", "description", "event_kinds"}

// serverOwnedKeys are keys authors tend to write that belong to the
// approval; they get a dedicated message.
var serverOwnedKeys = map[string]bool{
	"valid_from": true, "valid_until": true, "review_due": true, "status": true,
	"approved": true, "approved_by": true, "gueltig_ab": true, "gültig_ab": true,
}
