package mlcext

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// The PDF (mlcdoc) has two fonts: Liberation Sans and DejaVu Sans as a
// fallback (pdfGlyphRanges, generated from them). Colored emoji are not in
// them; mlcdoc replaces a few by colored symbols (PDFSubstitutes), every
// other missing character prints as an empty box.

// PDFSubstitute is an emoji mlcdoc prints as a colored symbol.
type PDFSubstitute struct {
	Emoji  string `json:"emoji"`
	Symbol string `json:"symbol"`
	Color  string `json:"color"`
}

// PDFSubstitutes as in mlcmd2pdf (src/markdown.rs, EMOJI_SUBSTITUTES).
var PDFSubstitutes = []PDFSubstitute{
	{"✅", "✔", "#237A57"}, {"❌", "✘", "#C0392B"}, {"❎", "✘", "#C0392B"},
	{"🔴", "●", "#D92D20"}, {"🟠", "●", "#F79009"}, {"🟡", "●", "#EAAA08"},
	{"🟢", "●", "#12B76A"}, {"🔵", "●", "#1570EF"}, {"🟣", "●", "#7A5AF8"},
	{"🟤", "●", "#93370D"},
}

// HasPDFGlyph reports whether r prints in the PDF: the fonts have it, or
// mlcdoc substitutes it. Variation selectors and joiners are dropped by
// mlcdoc and count as printable.
func HasPDFGlyph(r rune) bool {
	if r < 0x20 || r == 0x200D || r >= 0xFE00 && r <= 0xFE0F || r == 0x20E3 {
		return true
	}
	for _, s := range PDFSubstitutes {
		if []rune(s.Emoji)[0] == r {
			return true
		}
	}
	i := sort.Search(len(pdfGlyphRanges), func(i int) bool { return pdfGlyphRanges[i][1] >= r })
	return i < len(pdfGlyphRanges) && pdfGlyphRanges[i][0] <= r
}

// SymbolGroup is a group of the editor's symbol palette.
type SymbolGroup struct {
	Name    string   `json:"name"`
	Symbols []string `json:"symbols"`
}

// PDFSymbols is the palette of symbols that print in the PDF (tested
// against the fonts), for editors.
var PDFSymbols = []SymbolGroup{
	{"Status", strings.Split("✅ ❌ ✔ ✘ ✓ ✗ ☐ ☑ ☒ ⚠ ℹ ★ ☆ ⚑ ⚐", " ")},
	{"Farben", strings.Split("🔴 🟠 🟡 🟢 🔵 🟣 🟤 ● ○ ◆ ◇ ■ □ ▲ ▼", " ")},
	{"Pfeile", strings.Split("→ ← ↑ ↓ ↔ ↕ ⇒ ⇐ ⇔ ➜ ➔ ↻ ↺ ↩ ↪", " ")},
	{"Technik", strings.Split("⚙ ⚡ ☎ ✉ ✎ ✏ ✂ ♻ ☢ ☣ ⚕ ⚖ ⚗ ⚛ ⌘", " ")},
	{"Wetter", strings.Split("☀ ☁ ☂ ❄ ☃ ☔ ☕ ♨", " ")},
	{"Zeichen", strings.Split("° ± × ÷ ≈ ≠ ≤ ≥ € § © ® ™ • … ‰ µ Ω", " ")},
	{"Nummern", strings.Split("① ② ③ ④ ⑤ ⑥ ⑦ ⑧ ⑨ ⑩", " ")},
	{"Gesichter", strings.Split("😀 😃 😊 😉 😐 😕 😟 😢 😠 😮 ☺ ☹", " ")},
}

// checkGlyphs warns about characters that print as an empty box in the PDF
// (each character once, at its first place; code and comments included,
// front matter not).
func checkGlyphs(md string, add func(int, int, Severity, string, string)) {
	start := len(md) - len(frontMatterBlock(md).body)
	end := endSignatureAt(md)
	if end < start {
		return
	}
	seen := map[rune]bool{}
	for i, r := range md[start:end] {
		if r < 0x80 || seen[r] || HasPDFGlyph(r) || unicode.IsSpace(r) {
			continue
		}
		seen[r] = true
		line, col := position(md, start+i)
		add(line, col, Warning, "pdf-glyph-missing", fmt.Sprintf("„%c“ (U+%04X) fehlt in der Schrift des PDFs und erscheint dort als leeres Kästchen.", r, r))
		if len(seen) >= 20 {
			return
		}
	}
}
