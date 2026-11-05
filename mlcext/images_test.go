package mlcext

import (
	"slices"
	"strings"
	"testing"
)

var onePixelPNG = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func TestImageTargets(t *testing.T) {
	md := "![a](a.png) ![a2](a.png)\n\n![b](<b c.png> \"T\")\n\n`![k](k.png)`\n\n```\n![f](f.png)\n```\n\n![d](data:image/png;base64,AAAA)\n\n[Link](x.png)\n"
	if got, want := ImageTargets(md), []string{"a.png", "b c.png"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmbedImages(t *testing.T) {
	load := func(target string) ([]byte, bool) {
		switch target {
		case "ordner/a.png", "b c.png", "https://x.eu/p/c.png?v=1":
			return onePixelPNG, true
		case "text.png":
			return []byte("kein Bild"), true
		}
		return nil, false
	}
	uri := ImageDataURI(onePixelPNG, "a.png")
	if !strings.HasPrefix(uri, "data:image/png;base64,") {
		t.Fatalf("uri %q", uri)
	}
	for _, c := range []struct {
		name, in, want string
		failed         []string
	}{
		{"name as title, no folders", "![A](ordner/a.png)", `![A](` + uri + ` "a.png")`, nil},
		{"own title stays, angle brackets go", `![A](<b c.png> "Mein Titel")`, `![A](` + uri + ` "Mein Titel")`, nil},
		{"URL: file name without query", "![A](https://x.eu/p/c.png?v=1)", `![A](` + uri + ` "c.png")`, nil},
		{"not found stays", "![A](fehlt.png)", "![A](fehlt.png)", []string{"fehlt.png"}},
		{"no image stays", "![A](text.png)", "![A](text.png)", []string{"text.png"}},
		{"embedded stays", "![A](data:image/png;base64,AAAA)", "![A](data:image/png;base64,AAAA)", nil},
		{"code stays", "```\n![A](ordner/a.png)\n```", "```\n![A](ordner/a.png)\n```", nil},
		{"text around", "vor ![A](ordner/a.png) nach\n", `vor ![A](` + uri + ` "a.png") nach` + "\n", nil},
	} {
		got, _, failed := EmbedImages(c.in, load)
		if got != c.want || !slices.Equal(failed, c.failed) {
			t.Errorf("%s:\n got %q %q\nwant %q %q", c.name, got, failed, c.want, c.failed)
		}
	}
}

func TestImageDataURI(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`)
	if got := ImageDataURI(svg, "a.SVG"); !strings.HasPrefix(got, "data:image/svg+xml;base64,") {
		t.Errorf("svg: %q", got)
	}
	if got := ImageDataURI(svg, "a.txt"); got != "" {
		t.Errorf("svg text under another name: %q", got)
	}
	webp := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
	if !IsWebP(webp) {
		t.Error("WebP not recognized")
	}
	if got := ImageDataURI(webp, "a.webp"); got != "" {
		t.Errorf("WebP must stay a reference, the PDF would skip it: %q", got)
	}
}

func TestImageFileName(t *testing.T) {
	for in, want := range map[string]string{
		`C:\Users\Max\Bilder\logo.png`:  "logo.png",
		"../a/b.png":                    "b.png",
		"https://x.eu/a/c.jpg?size=1#t": "c.jpg",
	} {
		if got := ImageFileName(in); got != want {
			t.Errorf("%s: got %q", in, got)
		}
	}
}

func TestFormatImage(t *testing.T) {
	for _, c := range [][4]string{
		{"A", "a.png", "", "![A](a.png)"},
		{"A", "a b.png", "", "![A](<a b.png>)"},
		{"A", "a.png", `T"x`, `![A](a.png "Tx")`},
		{"a]b", "a.png", "", `![a\]b](a.png)`},
	} {
		if got := FormatImage(c[0], c[1], c[2]); got != c[3] {
			t.Errorf("got %q, want %q", got, c[3])
		}
	}
}
