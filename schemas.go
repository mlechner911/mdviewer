package main

// Schemas for form fields (mlcext.FieldSchema) a document names by its ID.
// They are never fetched from the address: an unknown document must not cause
// a request, and own installations often have no internet.

import (
	"path/filepath"

	"github.com/mlechner911/mdviewer/mlcext"
)

// schemaProvider is optional for an extension: a module build that knows
// schemas from elsewhere, such as its signed-in server. FormSchema runs during
// the analysis while the user types, so it must answer from memory.
type schemaProvider interface {
	FormSchema(id string) (mlcext.FieldSchema, bool)
}

// schemasDir is the user's schema folder (<user config>/marksafe/schemas).
func (a *App) schemasDir() string {
	if a.config == nil {
		return ""
	}
	return filepath.Join(a.config.Dir(), "schemas")
}

// findSchema looks up id for the analysis. It runs only for a document that
// names a schema.
func (a *App) findSchema(id string) (mlcext.FieldSchema, bool) {
	return lookupSchema(a.schemasDir(), a.extensions, id)
}

// lookupSchema: the local folder first, so a user can pin a schema, then the
// extensions.
func lookupSchema(dir string, extensions []extension, id string) (mlcext.FieldSchema, bool) {
	if dir != "" {
		local, errs := mlcext.LoadSchemas(dir)
		for _, err := range errs {
			logger().Warn("Schema file skipped", "dir", dir, "error", err)
		}
		if s, ok := local[id]; ok {
			return s, true
		}
	}
	for _, e := range extensions {
		if p, ok := e.(schemaProvider); ok {
			if s, ok := p.FormSchema(id); ok {
				return s, true
			}
		}
	}
	return mlcext.FieldSchema{}, false
}
