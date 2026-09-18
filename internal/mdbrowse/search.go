package mdbrowse

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// On-demand full-text search over the markdown tree.
//
// Scoring (higher wins, additive across categories):
//   - filename exact match:      100
//   - filename contains query:    80
//   - filename fuzzy match:    40..60 (by quality)
//   - title exact match:          90
//   - title contains query:       70
//   - title fuzzy match:       35..50 (by quality)
//   - body exact: 15 for the first hit + 3 per extra hit (cap 40)
//   - body fuzzy word:          8..20 (by quality)
//
// Every result carries a ~100-rune snippet around the first body hit
// with the match wrapped in <mark> so the UI can show WHY a page was
// suggested. Title/filename-only hits get a leading excerpt instead.

// SearchResult is one ranked hit of /api/search.
type SearchResult struct {
	Path    string  `json:"path"`
	Title   string  `json:"title"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet"` // HTML: escaped text + <mark> hits
}

const (
	maxSearchResults = 20
	maxSearchBytes   = 400 * 1024 // bodies above this are truncated for search
	snippetRunes     = 100
	fuzzyThreshold   = 0.5
)

// by score (best first, capped at maxSearchResults).
func (r *Root) Search(q string) ([]SearchResult, error) {
	if strings.TrimSpace(q) == "" {
		return []SearchResult{}, nil
	}
	files, err := r.GetMarkdownFiles()
	if err != nil {
		return nil, err
	}
	return searchFiles(r.Dir, files, q), nil
}

// searchFiles scores files against q and returns hits sorted by score.
func searchFiles(rootDir string, files []string, q string) []SearchResult {
	lq := strings.ToLower(q)
	results := make([]SearchResult, 0, len(files))

	for _, absPath := range files {
		relPath, err := RelativePath(rootDir, absPath)
		if err != nil {
			continue
		}
		title, err := ExtractTitle(absPath)
		if err != nil || title == "" {
			title = FormatTitle(strings.TrimSuffix(filepath.Base(absPath), ".md"))
		}
		raw, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}
		content := string(raw)
		// Whitespace-normalized once: all offsets below are rune offsets
		// into nbody, so snippets line up exactly.
		nbody := normSpace(stripFrontmatter(content))
		if len(nbody) > maxSearchBytes {
			nbody = nbody[:maxSearchBytes]
		}

		score := 0.0
		matched := false
		snippet := ""

		// --- filename ---
		base := strings.ToLower(strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath)))
		switch {
		case base == lq:
			score += 100
			matched = true
		case strings.Contains(base, lq):
			score += 80
			matched = true
		default:
			if fq := fuzzyQuality(base, lq); fq > 0 {
				score += 40 + 20*fq
				matched = true
			}
		}

		// --- title (H1) ---
		lt := strings.ToLower(title)
		switch {
		case lt == lq:
			score += 90
			matched = true
		case strings.Contains(lt, lq):
			score += 70
			matched = true
		default:
			if fq := fuzzyQuality(lt, lq); fq > 0 {
				score += 35 + 15*fq
				matched = true
			}
		}

		// --- body ---
		lb := strings.ToLower(nbody)
		if n := strings.Count(lb, lq); n > 0 {
			add := 15 + 3*(n-1)
			if add > 40 {
				add = 40
			}
			score += float64(add)
			matched = true
			ri := len([]rune(nbody[:strings.Index(lb, lq)]))
			snippet = snippetExact(nbody, ri, len([]rune(lq)), q)
		} else if wi, wl, wq := bestFuzzyWord(nbody, lq); wq > 0 {
			score += 8 + 12*wq
			matched = true
			snippet = snippetMarked(nbody, wi, wi+wl)
		}

		if !matched {
			continue
		}
		if snippet == "" {
			snippet = leadingExcerpt(nbody)
		}
		results = append(results, SearchResult{
			Path:    relPath,
			Title:   truncateTitle(title),
			Score:   math.Round(score*10) / 10,
			Snippet: snippet,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Path < results[j].Path
	})
	if len(results) > maxSearchResults {
		results = results[:maxSearchResults]
	}
	return results
}

// normSpace collapses all whitespace to single spaces.
func normSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// fuzzyQuality rates how well pattern matches text as a subsequence:
// 1.0 = contiguous, decreasing with gaps. 0 = no usable match.
func fuzzyQuality(text, pattern string) float64 {
	tr, pr := []rune(text), []rune(pattern)
	if len(pr) == 0 || len(pr) > len(tr) {
		return 0
	}
	best := 0.0
	for i := 0; i < len(tr); i++ {
		if tr[i] != pr[0] {
			continue
		}
		j, k := i, 0
		for j < len(tr) && k < len(pr) {
			if tr[j] == pr[k] {
				k++
			}
			j++
		}
		if k != len(pr) {
			continue
		}
		qual := float64(len(pr)) / float64(j-i)
		if i == 0 || isWordBoundary(tr[i-1]) {
			qual += 0.1
		}
		if qual > 1 {
			qual = 1
		}
		if qual > best {
			best = qual
		}
	}
	if best < fuzzyThreshold {
		return 0
	}
	return best
}

func isWordBoundary(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '/' || r == '_' ||
		r == '-' || r == '.' || r == '(' || r == '[' || r == '"' || r == '\''
}

// bestFuzzyWord finds the word with the best fuzzy quality. Returns its
// rune offset + length in the (already normalized) body, plus quality.
func bestFuzzyWord(nbody, pattern string) (off, ln int, qual float64) {
	off, pos := -1, 0
	for _, word := range strings.Split(nbody, " ") {
		if word != "" {
			if q := fuzzyQuality(strings.ToLower(word), pattern); q > qual {
				qual = q
				off = pos
				ln = len([]rune(word))
			}
		}
		pos += len([]rune(word)) + 1
	}
	return off, ln, qual
}

// windowAround cuts ~100 runes around [c0:c1), trimmed to word boundaries,
// and returns escaped (pre, marked, post) parts for <mark> assembly.
func windowAround(nbody string, c0, c1 int) (pre, marked, post string) {
	r := []rune(nbody)
	if c0 < 0 {
		c0 = 0
	}
	if c1 > len(r) {
		c1 = len(r)
	}
	lo := c0 - 45
	if lo < 0 {
		lo = 0
	}
	hi := c1 + 55
	if hi > len(r) {
		hi = len(r)
	}
	prefix, suffix := "", ""
	if lo > 0 {
		for lo < hi && r[lo] != ' ' {
			lo++
		}
		if lo < hi {
			lo++
		}
		prefix = "… "
	}
	if hi < len(r) {
		for hi > lo && r[hi-1] != ' ' {
			hi--
		}
		suffix = " …"
	}
	pre = prefix + escapeHTML(string(r[lo:c0]))
	marked = escapeHTML(string(r[c0:c1]))
	post = escapeHTML(string(r[c1:hi])) + suffix
	return pre, marked, post
}

// snippetExact marks every exact occurrence of q inside the window.
func snippetExact(nbody string, _, _ int, q string) string {
	r := []rune(nbody)
	ri := -1
	if i := strings.Index(strings.ToLower(nbody), strings.ToLower(q)); i >= 0 {
		ri = len([]rune(nbody[:i]))
	}
	if ri < 0 {
		return leadingExcerpt(nbody)
	}
	lo := ri - 45
	if lo < 0 {
		lo = 0
	}
	hi := ri + len([]rune(q)) + 55
	if hi > len(r) {
		hi = len(r)
	}
	for lo > 0 && r[lo] != ' ' {
		lo++
	}
	if lo > 0 {
		lo++
	}
	for hi < len(r) && r[hi-1] != ' ' {
		hi--
	}
	window := strings.TrimSpace(string(r[lo:hi]))
	if lo > 0 {
		window = "… " + window
	}
	if hi < len(r) {
		window = window + " …"
	}
	return highlightMark(escapeHTML(window), escapeHTML(q))
}

// snippetMarked marks one specific word range (fuzzy hits).
func snippetMarked(nbody string, c0, c1 int) string {
	pre, marked, post := windowAround(nbody, c0, c1)
	return pre + "<mark>" + marked + "</mark>" + post
}

// leadingExcerpt returns the first ~100 runes for title-only hits.
func leadingExcerpt(nbody string) string {
	r := []rune(nbody)
	if len(r) <= snippetRunes {
		return escapeHTML(strings.TrimSpace(nbody))
	}
	end := snippetRunes
	for end > 0 && r[end] != ' ' {
		end--
	}
	if end == 0 {
		end = snippetRunes
	}
	return escapeHTML(strings.TrimSpace(string(r[:end]))) + " …"
}

// highlightMark wraps every case-insensitive occurrence of needle in <mark>.
// Both inputs must already be HTML-escaped.
func highlightMark(text, needle string) string {
	if needle == "" {
		return text
	}
	lower, ln := strings.ToLower(text), strings.ToLower(needle)
	var b strings.Builder
	pos := 0
	for {
		i := strings.Index(lower[pos:], ln)
		if i < 0 {
			break
		}
		i += pos
		b.WriteString(text[pos:i])
		b.WriteString("<mark>")
		b.WriteString(text[i : i+len(needle)])
		b.WriteString("</mark>")
		pos = i + len(needle)
	}
	b.WriteString(text[pos:])
	return b.String()
}
