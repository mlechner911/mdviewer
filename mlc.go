package main

// Bindings for the MLC Markdown Extension (non-standard, see
// docs/MLC-EXTENSIONS.md): analysis, quick fixes, form fields and
// signatures, all done by mlcext.

import "github.com/mlechner911/mdviewer/mlcext"

// AnalyzeMLC returns the MLC Markdown Extension analysis of md (a
// non-standard extension, docs/MLC-EXTENSIONS.md): diagnostics, field and
// hint positions and the shared lists, for the editor and the status bar.
// Plain Markdown gets an empty analysis (Detected false).
func (a *App) AnalyzeMLC(md string) mlcext.Analysis {
	return mlcext.Analyze(md)
}

// MLCMoveHeader is the quick fix "Kopf ins Front Matter übernehmen": the
// legacy header lines become doc_id, version and title. "" when there is
// nothing to move.
func (a *App) MLCMoveHeader(md string) string {
	out, ok := mlcext.MoveHeaderToFrontMatter(md)
	if !ok {
		return ""
	}
	return out
}

// MLCFitTable sets the column widths of the table at line (1-based):
// mode "auto" (mlcdoc sizes by content) or "content" (full width, shares
// after the content). "" when line is in no table.
func (a *App) MLCFitTable(md string, line int, mode string) string {
	out, ok := mlcext.FitTableColumns(md, line, mode)
	if !ok {
		return ""
	}
	return out
}

// FormatFieldResult answers MLCFormatField: the written field, or why the
// dialog's input does not make one.
type FormatFieldResult struct {
	Text  string `json:"text"`
	Error string `json:"error"`
}

// MLCFormatField writes a form field from the field dialog
// (mlcext.FormatField), with its tooltip comment.
func (a *App) MLCFormatField(spec mlcext.FieldSpec) FormatFieldResult {
	text, err := mlcext.FormatField(spec)
	if err != nil {
		return FormatFieldResult{Error: err.Error()}
	}
	return FormatFieldResult{Text: text}
}

// MLCFieldAt reads the form field at line and col (1-based, characters)
// for the field dialog: its spec and the range to replace.
func (a *App) MLCFieldAt(md string, line, col int) mlcext.FieldEdit {
	return mlcext.FieldAt(md, line, col)
}

// MLCSignatureState tells whether md is signed and still matches its
// checksum (an edit voids the signature).
func (a *App) MLCSignatureState(md string) mlcext.SignatureState {
	return mlcext.CheckSignature(md)
}

// MLCRemoveSignature returns md without its signature block.
func (a *App) MLCRemoveSignature(md string) string {
	return mlcext.RemoveSignature(md)
}
