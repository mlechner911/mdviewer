package mlcext

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// reImage is an inline image: alt, target (plain or <with spaces>), title.
var reImage = regexp.MustCompile(`!\[([^\]\n]*)\]\(\s*(<[^>\n]+>|[^)\s]+)(\s+"[^"\n]*")?\s*\)`)

// imageMatches are the images of md outside code. Each entry holds the
// submatch offsets of reImage.
func imageMatches(md string) [][]int {
	var fences [][2]int
	offset, inFence, start := 0, "", 0
	for _, line := range strings.SplitAfter(md, "\n") {
		if f := codeFence(line); f != "" {
			switch {
			case inFence == "":
				inFence, start = f, offset
			case strings.HasPrefix(strings.TrimSpace(line), inFence):
				inFence = ""
				fences = append(fences, [2]int{start, offset + len(line)})
			}
		}
		offset += len(line)
	}
	if inFence != "" {
		fences = append(fences, [2]int{start, len(md)})
	}
	var out [][]int
next:
	for _, m := range reImage.FindAllStringSubmatchIndex(md, -1) {
		for _, f := range fences {
			if m[0] >= f[0] && m[0] < f[1] {
				continue next
			}
		}
		lineStart := strings.LastIndexByte(md[:m[0]], '\n') + 1
		if strings.Count(md[lineStart:m[0]], "`")%2 == 1 {
			continue
		}
		out = append(out, m)
	}
	return out
}

func imageTarget(md string, m []int) string {
	return strings.Trim(md[m[4]:m[5]], "<>")
}

// ImageTargets lists the images of md that are not embedded yet, once each,
// in document order.
func ImageTargets(md string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range imageMatches(md) {
		t := imageTarget(md, m)
		if strings.HasPrefix(t, "data:") || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

// EmbedImages writes the images load returns into md as data: URLs, with the
// file name as title unless the image has one. A target load has no data for,
// or whose data is no image, stays as it is and is listed in failed.
func EmbedImages(md string, load func(target string) ([]byte, bool)) (out string, embedded, failed []string) {
	var b strings.Builder
	last := 0
	done := map[string]bool{}
	for _, m := range imageMatches(md) {
		t := imageTarget(md, m)
		if strings.HasPrefix(t, "data:") {
			continue
		}
		data, ok := load(t)
		uri := ""
		if ok {
			uri = ImageDataURI(data, t)
		}
		if uri == "" {
			if !done[t] {
				failed = append(failed, t)
			}
			done[t] = true
			continue
		}
		if !done[t] {
			embedded = append(embedded, t)
		}
		done[t] = true
		title := ""
		if m[6] >= 0 {
			title = strings.Trim(strings.TrimSpace(md[m[6]:m[7]]), `"`)
		} else {
			title = ImageFileName(t)
		}
		b.WriteString(md[last:m[0]])
		b.WriteString(FormatImage(md[m[2]:m[3]], uri, title))
		last = m[1]
	}
	b.WriteString(md[last:])
	return b.String(), embedded, failed
}

// ImageDataURI is data as a data: URL, "" when it cannot be embedded. name
// decides only for SVG, which is text and has no signature of its own. WebP
// stays out: mlcdoc reads it from a file, but skips it as a data: URL.
func ImageDataURI(data []byte, name string) string {
	mime := http.DetectContentType(data)
	switch mime {
	case "image/png", "image/jpeg", "image/gif":
	default:
		if !strings.HasSuffix(strings.ToLower(ImageFileName(name)), ".svg") || !strings.Contains(string(data), "<svg") {
			return ""
		}
		mime = "image/svg+xml"
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}

// IsWebP reports a WebP image: usable as a reference, not embeddable.
func IsWebP(data []byte) bool { return http.DetectContentType(data) == "image/webp" }

// ImageFileName is the file name of an image target: never the folders, they
// would travel with the document.
func ImageFileName(target string) string {
	t := strings.ReplaceAll(target, `\`, "/")
	if u, err := url.Parse(t); err == nil && u.Scheme != "" && u.Host != "" {
		t = u.Path
	} else if k := strings.IndexAny(t, "?#"); k >= 0 {
		t = t[:k]
	}
	return path.Base(t)
}

// FormatImage writes an image. A target with spaces goes in angle brackets.
func FormatImage(alt, target, title string) string {
	if strings.ContainsAny(target, " \t") {
		target = "<" + target + ">"
	}
	out := "![" + strings.NewReplacer("[", `\[`, "]", `\]`, "\n", " ").Replace(alt) + "](" + target
	if title = strings.ReplaceAll(strings.ReplaceAll(title, `"`, ""), "\n", " "); title != "" {
		out += ` "` + title + `"`
	}
	return out + ")"
}
