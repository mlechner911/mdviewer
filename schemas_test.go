package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mlechner911/mdviewer/mlcext"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const testSchemaID = "https://schemas.example/forms/x/v1"

type plainExtension struct{}

func (plainExtension) Service() application.Service { return application.Service{} }
func (plainExtension) FileMenu(*application.Menu, func(key, fallback string) string) {
}
func (plainExtension) HasSettings() bool { return false }

type providingExtension struct {
	plainExtension
	title string
}

func (p providingExtension) FormSchema(id string) (mlcext.FieldSchema, bool) {
	if id != testSchemaID {
		return mlcext.FieldSchema{}, false
	}
	return mlcext.FieldSchema{ID: id, Title: map[string]string{"de": p.title}}, true
}

func TestLookupSchema(t *testing.T) {
	dir := t.TempDir()
	extensions := []extension{plainExtension{}, providingExtension{title: "Server"}}

	if s, ok := lookupSchema(dir, extensions, testSchemaID); !ok || s.Title["de"] != "Server" {
		t.Errorf("from the extension: %+v %v", s, ok)
	}
	if _, ok := lookupSchema(dir, []extension{plainExtension{}}, testSchemaID); ok {
		t.Error("found a schema nobody has")
	}

	file := "id: " + testSchemaID + "\ntitle: { de: Lokal }\nfields:\n  a.b: { type: text }\n"
	if err := os.WriteFile(filepath.Join(dir, "x.yaml"), []byte(file), 0o644); err != nil {
		t.Fatal(err)
	}
	if s, ok := lookupSchema(dir, extensions, testSchemaID); !ok || s.Title["de"] != "Lokal" {
		t.Errorf("the local file wins: %+v %v", s, ok)
	}
}
