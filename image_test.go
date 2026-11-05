package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"marksafe/internal/config"
)

func TestReferenceTo(t *testing.T) {
	doc := filepath.Join("base", "doc.md")
	for path, want := range map[string]string{
		filepath.Join("base", "a.png"):         "a.png",
		filepath.Join("base", "bild", "a.png"): "bild/a.png",
		filepath.Join("other", "a.png"):        "",
	} {
		if got := referenceTo(path, doc); got != want {
			t.Errorf("%s: got %q, want %q", path, got, want)
		}
	}
	if got := referenceTo("a.png", ""); got != "" {
		t.Errorf("unsaved document: got %q", got)
	}
}

func TestImageMarkdown(t *testing.T) {
	dir := t.TempDir()
	png := filepath.Join(dir, "a b.png")
	if err := os.WriteFile(png, []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"), 0o600); err != nil {
		t.Fatal(err)
	}
	doc := filepath.Join(dir, "doc.md")
	a := &App{}

	got, err := a.ImageMarkdown(ImageSpec{Path: png, Alt: "A"}, doc)
	if err != nil || got != "![A](<a b.png>)" {
		t.Errorf("reference: %q, %v", got, err)
	}
	got, err = a.ImageMarkdown(ImageSpec{Path: png, Alt: "A", Embed: true}, doc)
	if err != nil || !strings.HasPrefix(got, "![A](data:image/png;base64,") || !strings.HasSuffix(got, ` "a b.png")`) {
		t.Errorf("embed: %q, %v", got, err)
	}
	if _, err = a.ImageMarkdown(ImageSpec{Path: png, Alt: "A"}, filepath.Join(t.TempDir(), "doc.md")); err == nil {
		t.Error("reference outside the folder must fail")
	}
	got, err = a.ImageMarkdown(ImageSpec{DataURI: "data:image/png;base64,AAAA", Alt: "A"}, "")
	if err != nil || got != "![A](data:image/png;base64,AAAA)" {
		t.Errorf("pasted: %q, %v", got, err)
	}
	if _, err = a.ImageMarkdown(ImageSpec{DataURI: "data:text/html;base64,AAAA"}, ""); err == nil {
		t.Error("pasted non-image must fail")
	}
}

func TestInspectWebP(t *testing.T) {
	dir := t.TempDir()
	webp := filepath.Join(dir, "a.webp")
	if err := os.WriteFile(webp, []byte("RIFF\x00\x00\x00\x00WEBPVP8 "), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := (&App{}).InspectImage(webp, filepath.Join(dir, "doc.md"))
	if err != nil || c.Embeddable || c.Reference != "a.webp" {
		t.Errorf("inside the folder: %+v, %v", c, err)
	}
	if _, err := (&App{}).InspectImage(webp, filepath.Join(t.TempDir(), "doc.md")); err == nil {
		t.Error("outside the folder it can be neither referenced nor embedded")
	}
}

func TestInspectImageRejectsOtherFiles(t *testing.T) {
	txt := filepath.Join(t.TempDir(), "a.png")
	if err := os.WriteFile(txt, []byte("kein Bild"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (&App{}).InspectImage(txt, ""); err == nil {
		t.Error("text under an image name must fail")
	}
}

func TestEmbedFollowsTheWhitelist(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("AppData", t.TempDir())
	cfg, err := config.NewConfigManager()
	if err != nil {
		t.Fatal(err)
	}
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	allowed, blocked := t.TempDir(), t.TempDir()
	for _, p := range []string{filepath.Join(allowed, "a.png"), filepath.Join(allowed, "b c.png"), filepath.Join(blocked, "x.png")} {
		if err := os.WriteFile(p, png, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := cfg.AddPath(allowed); err != nil {
		t.Fatal(err)
	}
	a := &App{config: cfg}
	doc := filepath.Join(allowed, "doc.md")
	md := "![A](a.png)\n\n![B](<b c.png>)\n\n![X](" + filepath.ToSlash(filepath.Join(blocked, "x.png")) + ")\n\n![F](fehlt.png)\n\n![U](https://nicht-frei.example/u.png)\n"

	plan := a.PlanEmbedImages(md, doc)
	if len(plan.Images) != 2 || len(plan.Missing) != 3 || plan.Bytes != int64(2*len(png)) {
		t.Errorf("plan: %+v", plan)
	}
	res := a.EmbedImages(md, doc)
	if len(res.Embedded) != 2 || len(res.Missing) != 3 {
		t.Errorf("result: embedded %q missing %q", res.Embedded, res.Missing)
	}
	if !strings.Contains(res.Markdown, `"a.png")`) || !strings.Contains(res.Markdown, `"b c.png")`) || !strings.Contains(res.Markdown, "fehlt.png") {
		t.Errorf("markdown: %s", res.Markdown)
	}
	if strings.Contains(res.Markdown, filepath.ToSlash(allowed)) {
		t.Error("the folder must not travel with the document")
	}
	if got := a.PlanEmbedImages("![A](a.png)", ""); len(got.Missing) != 1 {
		t.Errorf("relative image without a document: %+v", got)
	}
}
