// Package mlcext checks Markdown documents written with the MLC Markdown
// Extension.
//
// The MLC Markdown Extension is NOT standard Markdown. It is an MLC-specific
// extension for documents of the MLC Sensor Monitor (SOPs, forms, reports)
// that the MLC PDF toolkit (mlcdoc / mlcmd2pdf) turns into PDFs: form fields
// "[@…]", field tooltips "<!--* … -->", fill hints "<!--! … -->", document
// header lines ("**Dokumenten-ID:** SOP-QM-001") and the upload front
// matter. See docs/MLC-EXTENSIONS.md in the MarkSafe repository.
//
// The package is the single source of these rules for every MLC tool: the
// MarkSafe editor shows its diagnostics, the Sensor Monitor uses the same
// checks on upload. Form fields follow mlcmd2pdf (src/form.rs) exactly, so a
// document without errors here is one mlcdoc accepts.
//
// Messages are German (the documents and their authors are); every
// diagnostic also carries a stable Code for other languages.
package mlcext
