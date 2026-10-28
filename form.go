package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"marksafe/internal/filesystem"

	"github.com/mlechner911/mdviewer/mlcext"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Filling in a form (plan P-SENS-20261028-01 Phase 2): the mask comes from
// mlcext.FormSchema, every value goes through mlcext.FillFields into the
// Markdown, and saving writes a new file with the record's "meta:" block.
// The form itself is never overwritten.

// MLCFormSchema describes the fields of md for the input mask.
func (a *App) MLCFormSchema(md string) []mlcext.FormField {
	return mlcext.FormSchema(md)
}

// MLCFillForm sets values (field name → value) into md.
func (a *App) MLCFillForm(md string, values map[string]string) string {
	out, _ := mlcext.FillFields(md, values)
	return out
}

// MLCFormRefOf is the form md is filled from (for a record: the one it pins).
func (a *App) MLCFormRefOf(md string) mlcext.FormRef {
	return mlcext.FormRefOf(md)
}

// MLCIsFormRecord reports a filled-in form (meta.schema mlc-form-record/v1).
func (a *App) MLCIsFormRecord(md string) bool {
	return mlcext.IsFormRecord(md)
}

// ErrNotARecord: saving a filled-in form must not replace a file that is
// not itself a filled-in form, above all not the form it came from.
var ErrNotARecord = errors.New("Die Datei ist kein ausgefülltes Formblatt und wird nicht überschrieben (die Vorlage bleibt unverändert). Bitte einen anderen Namen wählen.")

// SaveFilledForm writes md as a record: the "meta:" block with form, filler
// and time is set, then the file is saved at path, or after a save dialog
// suggesting "<doc_id>_<date>.md" when path is empty. An existing file is
// only replaced when it is a record itself.
func (a *App) SaveFilledForm(md string, form mlcext.FormRef, filledBy, path string) (*FileResult, error) {
	now := time.Now()
	rec, err := mlcext.SetFormRecord(md, mlcext.FormRecord{
		Form: form, FilledBy: filledBy, CompletedAt: now.Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		name := form.DocID
		if name == "" {
			name = "Formblatt"
		}
		dlg := application.Get().Dialog.SaveFile().
			SetMessage("Ausgefülltes Formblatt speichern").
			SetFilename(name + "_" + now.Format("2006-01-02") + ".md").
			AddFilter("Markdown Files (*.md)", "*.md")
		if dir := a.documentRoot(); dir != "" {
			dlg.SetDirectory(dir)
		}
		path, err = dlg.PromptForSingleSelection()
		if err != nil || path == "" {
			return nil, err
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if old, err := os.ReadFile(abs); err == nil && !mlcext.IsFormRecord(string(old)) {
		return nil, ErrNotARecord
	}
	if err := filesystem.WriteFile(abs, rec); err != nil {
		return nil, err
	}
	a.config.AddRecentFile(abs)
	return &FileResult{Path: abs, Content: rec}, nil
}

// LoadFormValues reads field values from a JSON file chosen in a dialog: an
// object of field name → value (text, number or true/false for a check box).
// For testing a form and for prefilling it from another system.
func (a *App) LoadFormValues() (map[string]string, error) {
	path, err := application.Get().Dialog.OpenFile().
		SetTitle("Werte für das Formblatt laden").
		AddFilter("JSON (*.json)", "*.json").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseFormValues(raw)
}

func parseFormValues(raw []byte) (map[string]string, error) {
	var in map[string]any
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("keine gültige JSON-Datei (erwartet: {\"feldname\": \"Wert\"}): %w", err)
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		switch v := v.(type) {
		case string:
			out[k] = v
		case bool:
			if v {
				out[k] = "x"
			} else {
				out[k] = ""
			}
		case float64:
			out[k] = strconv.FormatFloat(v, 'f', -1, 64)
		case nil:
			out[k] = ""
		default:
			return nil, fmt.Errorf("Feld %q: nur Text, Zahl oder true/false", k)
		}
	}
	return out, nil
}
