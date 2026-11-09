package mlcext

import (
	"encoding/json"
	"os"
	"testing"
)

// imageAltCases: alt text → text, size, and whether the size is unreadable.
// testdata/imagesize.json holds the same for the TypeScript port
// (packages/mlc-markdown); regenerate with: go test -run TestImageSizeVectors -update
var imageAltCases = []struct {
	Alt  string    `json:"alt"`
	Text string    `json:"text"`
	Size ImageSize `json:"size"`
	Err  bool      `json:"error,omitempty"`
}{
	{"Foto", "Foto", ImageSize{}, false},
	{"Foto|200", "Foto", ImageSize{Width: 200}, false},
	{"Foto|200x100", "Foto", ImageSize{Width: 200, Height: 100}, false},
	{"Foto|50%", "Foto", ImageSize{Percent: 50}, false},
	{"Foto | 100% ", "Foto", ImageSize{Percent: 100}, false},
	{"|300", "", ImageSize{Width: 300}, false},
	{"A|B|120", "A|B", ImageSize{Width: 120}, false},
	{"Ja|Nein", "Ja|Nein", ImageSize{}, false},
	{"Foto|", "Foto|", ImageSize{}, false},
	{"Foto|200px", "Foto", ImageSize{}, true},
	{"Foto|200x", "Foto", ImageSize{}, true},
	{"Foto|150%", "Foto", ImageSize{}, true},
	{"Foto|0%", "Foto", ImageSize{}, true},
	{"Foto|0", "Foto", ImageSize{}, true},
	{"Foto|200x0", "Foto", ImageSize{}, true},
}

func TestSplitImageAlt(t *testing.T) {
	for _, c := range imageAltCases {
		text, size, err := SplitImageAlt(c.Alt)
		if text != c.Text || size != c.Size || (err != nil) != c.Err {
			t.Errorf("SplitImageAlt(%q) = %q %+v %v; want %q %+v error=%v", c.Alt, text, size, err, c.Text, c.Size, c.Err)
		}
		if c.Err || size.IsZero() {
			continue
		}
		if back, s2, _ := SplitImageAlt(ImageAlt(text, size)); back != text || s2 != size {
			t.Errorf("round trip %q: %q %+v", c.Alt, back, s2)
		}
	}
	if ImageAlt("Foto", ImageSize{}) != "Foto" || ImageAlt("Foto", ImageSize{Width: 2, Height: 3}) != "Foto|2x3" {
		t.Error("ImageAlt")
	}
}

func TestCheckImageSize(t *testing.T) {
	md := "# T\n\n![Foto|200px](a.png) ![Logo|50%](b.png)\n\n```\n![x|abc1](c.png) ![x|9x](c.png)\n```\n"
	ds := Check(md, Options{})
	var got []Diagnostic
	for _, d := range ds {
		if d.Code == "image-size" {
			got = append(got, d)
		}
	}
	if len(got) != 1 || got[0].Line != 3 || got[0].Col != 1 || got[0].Severity != Warning {
		t.Fatalf("%+v", got)
	}
}

func TestImageSizeVectors(t *testing.T) {
	b, _ := json.MarshalIndent(imageAltCases, "", "  ")
	b = append(b, '\n')
	if *update {
		if err := os.WriteFile("testdata/imagesize.json", b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old, err := os.ReadFile("testdata/imagesize.json")
	if err != nil || string(old) != string(b) {
		t.Errorf("testdata/imagesize.json is out of date; regenerate with -update and check the diff")
	}
}
