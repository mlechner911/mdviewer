package mlcext

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ImageSize is the size of an image, written after "|" in its alt text as in
// Obsidian, but in the plain Markdown form:
//
//	![Foto|200](foto.jpg)      width 200 px
//	![Foto|200x100](foto.jpg)  fitted into 200 × 100, aspect ratio kept
//	![Foto|50%](foto.jpg)      half the width of the text column
//
// Px are CSS pixels (0.75 pt in the PDF). Other Markdown tools show
// "Foto|200" as alt text; the image itself shows everywhere.
type ImageSize struct {
	// Width and Height in px; Height only with a box ("200x100").
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
	// Percent of the text column's width, 1 to 100; then Width and Height
	// are 0.
	Percent int `json:"percent,omitempty"`
}

// IsZero: no size.
func (s ImageSize) IsZero() bool { return s == ImageSize{} }

// String is the written form: "200", "200x100", "50%", "" for no size.
func (s ImageSize) String() string {
	switch {
	case s.Percent > 0:
		return strconv.Itoa(s.Percent) + "%"
	case s.Width > 0 && s.Height > 0:
		return fmt.Sprintf("%dx%d", s.Width, s.Height)
	case s.Width > 0:
		return strconv.Itoa(s.Width)
	}
	return ""
}

var reImageSize = regexp.MustCompile(`^(?:(\d+)|(\d+)x(\d+)|(\d+)%)$`)

// ParseImageSize reads "200", "200x100" or "50%". A percentage must lie
// between 1 and 100, a size in px must not be 0.
func ParseImageSize(s string) (ImageSize, error) {
	m := reImageSize.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return ImageSize{}, fmt.Errorf("Bildgröße %q: erwartet Breite in px (`200`), ein Feld (`200x100`) oder Prozent der Textbreite (`50%%`)", s)
	}
	n := func(i int) int { v, _ := strconv.Atoi(m[i]); return v }
	var size ImageSize
	switch {
	case m[1] != "":
		size.Width = n(1)
	case m[2] != "":
		size.Width, size.Height = n(2), n(3)
	default:
		size.Percent = n(4)
		if size.Percent < 1 || size.Percent > 100 {
			return ImageSize{}, fmt.Errorf("Bildgröße %q: Prozent der Textbreite liegt zwischen 1 und 100", s)
		}
		return size, nil
	}
	if size.Width == 0 || (m[2] != "" && size.Height == 0) {
		return ImageSize{}, fmt.Errorf("Bildgröße %q: 0 px zeigt nichts", s)
	}
	return size, nil
}

// SplitImageAlt splits an alt text into the text and the size after its
// last "|". Text after "|" that does not start with a digit is no size and
// stays part of the text ("A|B"). A size that cannot be read comes back as
// err, with the text before "|".
func SplitImageAlt(alt string) (text string, size ImageSize, err error) {
	k := strings.LastIndexByte(alt, '|')
	if k < 0 {
		return alt, ImageSize{}, nil
	}
	raw := strings.TrimSpace(alt[k+1:])
	if raw == "" || raw[0] < '0' || raw[0] > '9' {
		return alt, ImageSize{}, nil
	}
	size, err = ParseImageSize(raw)
	return strings.TrimRight(alt[:k], " "), size, err
}

// ImageAlt writes text and size as an alt text ("Foto|200"); without a
// size it is text.
func ImageAlt(text string, size ImageSize) string {
	if size.IsZero() {
		return text
	}
	return text + "|" + size.String()
}

// imageSizeDiagnostics reports image sizes that cannot be read, in every
// Markdown document (not only those of the extension). With pdf it also
// notes that mlcdoc does not apply a size yet (I-MD2PD-20261028-01).
func imageSizeDiagnostics(md string, pdf bool) []Diagnostic {
	var ds []Diagnostic
	for _, m := range imageMatches(md) {
		_, size, err := SplitImageAlt(md[m[2]:m[3]])
		line, col := position(md, m[0])
		switch {
		case err != nil:
			ds = append(ds, Diagnostic{line, col, Warning, "image-size", err.Error() + "; das Bild wird in seiner eigenen Größe gezeigt."})
		case pdf && !size.IsZero():
			ds = append(ds, Diagnostic{line, col, Info, "image-size-pdf", fmt.Sprintf("Bildgröße `%s` gilt in der Vorschau; im PDF erscheint das Bild noch in seiner eigenen Größe.", size)})
		}
	}
	return ds
}
