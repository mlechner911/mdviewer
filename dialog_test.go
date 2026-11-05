package main

import (
	"errors"
	"testing"
)

type fakeDialog struct {
	path string
	err  error
}

func (d fakeDialog) PromptForSingleSelection() (string, error) { return d.path, d.err }

func TestPromptPath(t *testing.T) {
	failed := errors.New("no dialog")
	for _, c := range []struct {
		name    string
		dlg     fakeDialog
		want    string
		wantErr error
	}{
		{"chosen", fakeDialog{path: "a.md"}, "a.md", nil},
		{"cancelled on macOS and Linux", fakeDialog{}, "", nil},
		{"cancelled on Windows", fakeDialog{err: errors.New("cancelled by user")}, "", nil},
		{"failed", fakeDialog{err: failed}, "", failed},
	} {
		got, err := promptPath(c.dlg)
		if got != c.want || !errors.Is(err, c.wantErr) {
			t.Errorf("%s: got %q, %v; want %q, %v", c.name, got, err, c.want, c.wantErr)
		}
	}
}
