# Changelog

All notable changes to MarkSafe. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- **Field names with a dot** (`[@#probe.calibration.last]`) for properties a **field schema** describes. The schema comes from the front matter (`schema:` with an `https://schemas.mlcgo.eu/…/v1` address or inline) and is looked up in the folder **File › Open Field Schemas Folder**. The editor completes property names, shows their description on hover, warns about unknown properties and about a date pattern on a text field; the preview can show the schema's example values. In the PDF a dotted name becomes real form levels (`probe` › `calibration` › `last`).
- **Filling in a form** shows the document's text, images and links between the fields, so the instructions stay in view. Field names under the inputs can be hidden.
- **HTML in the PDF:** `<b>`, `<strong>`, `<i>`, `<em>`, `<s>`, `<del>` and `<code>` print as bold, italic, struck through and code; other inline tags print their text. HTML blocks such as `<div>` or `<details>` do not appear in the PDF; the editor marks them when a PDF can be made, and the export says so. The HTML preview and export stay as on GitHub.
- **Installer:** the last page offers to run MarkSafe.
- **Start page** shows what MarkSafe can do, with an embedded logo.
- **Images:** 🖼 in the toolbar inserts a file, a picture from the clipboard or a dropped file, as a reference or embedded in the document. Saving a filled-in form and Save As offer to embed the document's images, also **File › Embed Images…**; a signed document is never changed. The editor shows an embedded image as `🖼 logo.png ‹44,9 KB›`; a click shows the data, ▲ or Esc shortens it again.

### Changed
- The toolbar menu **MLC ▾** is replaced by a button for the form field dialog, shown for every document. With a schema, the old prefill names are no longer suggested; existing ones stay valid.
- The bundled PDF toolkit is mlcdoc 0.21.0 (mlcmd2pdf 3.29.0). From this version mlcdoc is under the **Elastic License 2.0** (`licenses/mlcdoc/LICENSE-mlcdoc`); MarkSafe stays MIT. Its packages now also carry the licenses of the fonts Inter, Dancing Script and JetBrains Mono and of the PRC reader.
- The toolbar groups the headings H1 to H3 in a menu **H ▾**.

### Fixed
- The check now warns when a form field holds characters the PDF's field font lacks (`pdf-field-glyph-missing`, `mlcext.HasPDFFieldGlyph`): ✔, ✅ or ⇒ print in the text but as an empty box in a field.
- Cancelling a file dialog on Windows no longer reports an error.
- Filling in: an edit in the Markdown now replaces what a field of the mask still showed, and images resolve from the form's folder in the mask, the preview and the PDF.
- `Kontrolle: [@#k] Ort: [@#o]`: "Ort" labels only the following field, not the check box.

### Known limitations
- WebP images stay references: the PDF skips them when embedded. A reference works only inside the document's folder.

## [1.10.0] – 2026-11-04

### Security
- The bundled PDF toolkit is mlcdoc 0.19.1: mlcmd2pdf 3.28.3 with upstream md2pdf 3.4.2 (updated `zip`, `rustls-webpki`, `webpki-roots`, `pulldown-cmark`) and PdfCraft 0.18.1 with its fixes.

### Added
- MarkSafePlus: **File › Sensor Monitor › Values from Event…** fills the form being filled in with the values of a chosen event, if the form belongs to it. MarkSafe itself gains a general hook for this, the same path as "Load values…" (T-MARKS-20261028-04).

### Changed
- The module build calls itself **MarkSafePlus** in the window title, the menus and the About window, keeps its own WebView2 profile on Windows and is a separate app on macOS (`MarkSafePlus.app`, `eu.mlcgo.marksafeplus`, signed and notarized). Its packages are named `marksafeplus-<version>-…`.
- The Sensor Monitor module moved out of this repository into a private submodule (`frontend/src/sensormonitor`). Public clones and builds are unaffected; `SENSOR=1` needs the submodule. `sensormonitor_on.go`, the only part left here, only registers it.

### Fixed
- PDF: a numbered list counted `1. 1. 1.` as soon as an item had a second paragraph, a code block, a table or a quote; the block now stays under its item. An item's text stays before a block that follows it directly, and a list keeps its start number (`5.`).
- Sensor Monitor module: saving unchanged text no longer says it matches the approved version (it may match the waiting draft). Dates in Postgres format from the server also show correctly where the web view is stricter (macOS).

## [1.9.2] – 2026-11-04

### Changed
- The bundled PDF toolkit is mlcdoc 0.18.0 (PdfCraft form fixes; mlcmd2pdf stays 3.28.1).
- Sensor Monitor module: saving a document whose draft is waiting for approval revises that draft (Sensor Monitor since 2026-11-03). The dialog says so before saving, the message afterwards says "revised" and that someone else must approve or reject it. The status bar adds "draft 1.2 waiting", and its tooltip shows the approved version with approval and review dates (T-MARKS-20261103-01, T-MARKS-20261021-03).

### Fixed
- Field dialog: the pattern `x` and patterns starting with `Unterschrift:` or `Signature:` are refused; they would have become a ticked check box or a signature field.

### Added
- Front matter block **`profile`** ("Steckbrief", `mlcext` 0.15.0): purpose, appears, adapt (placeholders the organisation fills in), keep, orientation. A placeholder declared in `profile.adapt` may stand in the text as `{{name}}` and is a warning (`placeholder-open`), not an error; the Sensor Monitor does not approve while one is open. `ParseProfile`, `OpenPlaceholders`, `IsPlaceholderName`.
- Front matter keys **`attachment_kind`** and **`attachment_once`** (`mlcext` 0.14.0): a template as a kind of attachment in the Sensor Monitor (short code `CAPA`, `SB`; one per event or device). The checker reports an invalid or reserved code (`ANH`), a code without `doc_id` and a non-boolean `attachment_once`. `IsAttachmentKind`.

## [1.9.1] – 2026-10-29

### Fixed
- Ticked check boxes (`[@x]`) and signature fields printed wrong in the PDF: the bundled PDF toolkit was mlcdoc 0.16.0 (mlcmd2pdf 3.26), which predates both. Now mlcdoc 0.17.0 (mlcmd2pdf 3.28.1). Filled-in forms saved with 1.9.0 print correctly once opened in this version.

## [1.9.0] – 2026-10-29

### Added
- **Field dialog** for form fields of the MLC Markdown Extension: kind (text, several lines, check box, date, time, own pattern, signature), name (with the names the Sensor Monitor fills in, duplicate warning), width, lines, preset, ticked, role and tooltip, with the written syntax and a preview. Insert: "Form field…" in the MLC menu or `Ctrl+Shift+F`. Edit: the box "✎ Field" after the field under the cursor, a double click or right click on the field. `mlcext.FormatField`, `FieldAt` (vectors in `mlcext/testdata/fieldspec.json`).
- **Fill in a form** ("Ausfüllen" in the preview bar): input mask next to the editor, values straight into the fields, "Werte laden…" from JSON, saving as a filled-in form with the `meta:` block (who, when, which form version) under `<doc_id>_<date>.md`; the form is never overwritten. Date and time fields (`TT.MM.JJJJ`, `HH:MM`) use the system's pickers with "Heute"/"Jetzt"; an invalid earlier value is flagged. The mask follows the app theme (dark mode). A filled-in form's PDF is named after the day it was completed (`FB-LAG-001_2026-10-28.pdf`), not the form's version. `mlcext.SetFormRecord`, `FormRecordOf`, `FormRefOf`.
- **Signature field** `[@Unterschrift: Rolle]` / `[@#name Unterschrift: Rolle]` (English `Signature:`, as mlcmd2pdf 3.28.0): names who signs; checker, preview (line with the role below) and the editor's building blocks know it. Filling leaves it alone.
- **Ticked check box** `[@x]` / `[@#name x]` (as mlcmd2pdf 3.27.0): the checker and the preview know it, the preview draws a cross.
- **Filling forms** in `mlcext` (plan P-SENS-20261028-01): `FillFields` puts values into the fields (also unnamed ones, check boxes ticked by `x`), `FieldValues` reads them back, `FormSchema` describes the fields for an input mask with a label taken from the text around them; moved here from the Sensor Monitor.
- Front matter block **`meta:`** for the record data of a filled-in form (`schema: mlc-form-record/v1`), not printed; `IsFormRecord`.

### Changed
- **License of the Sensor Monitor module: Elastic License 2.0** (`sensormonitor/LICENSE`, SPDX headers); MarkSafe without the module stays MIT. The About window of a `SENSOR=1` build names it. The portable packages now contain their licenses (`licenses/`: MarkSafe, mlcdoc and its third-party notices).
- **The connection to the MLC Sensor Monitor is an optional module.** Public builds no longer contain it; `task … SENSOR=1` builds it in (Go tag `sensormonitor`, frontend variant). Its code lives in `sensormonitor/` and `frontend/src/sensormonitor/`, its address in its own `sensormonitor.json` (taken over once from `config.json`). File › Settings… only appears when a module has settings. See `docs/SENSOR-MONITOR.md`.
- **The Windows installer of the module build is "MarkSafePlus"**: own folder, shortcuts and uninstall entry, installable next to MarkSafe; it contains the module's license (`licenses\LICENSE-sensormonitor`).

### Fixed
- A new document from a template counted as saved: quitting or closing its tab lost it without a question. Changing the app language replaced it (and a document opened from the Sensor Monitor) with the welcome text.
- Filling in: a pattern field with an escaped `=`, `|`, `[`, `]` or `\` (as the field dialog writes it) lost its pattern; fields kept growing after a long value was shortened; fast typing could lose a keystroke; closing the mask with ✕ let Ctrl+S overwrite the form; cancelled Save As made the record forget its file; two fields with the same name confused the mask.
- The Sensor Monitor address of 1.8 is kept in `config.json` by the standard build, so MarkSafePlus can still take it over.

## [1.8.0] – 2026-10-21

### Added
- **Connection to the MLC Sensor Monitor**, off by default: nothing of it is shown until its address is entered under File › Settings… › Connections. Then File › Sensor Monitor offers:
  - **Sign In…** by device login of mlcauth: confirm a code in the browser; the sign-in lasts an hour and is kept only in memory.
  - **Open…**: look up a document ID and open the approved version or the waiting draft.
  - **Save as Draft…**: a new version as a draft, with reason and, when the organisation asks for it, the signing PIN. Approving stays in the Sensor Monitor.
  - The status bar shows whether the document matches the approved version. If it does not, printing and PDF export suggest the watermark ENTWURF for that PDF; nothing is enforced.
- **File › Settings…** (`Ctrl+,`).

### Fixed
- Closing a tab or the window (the X, `⌘Q` on macOS) with unsaved changes now asks: Save, Don't save or Cancel. Before, the changes were lost without a question.
- Header lines made of variables (`**Dokumenten-ID:** {{doc_id}}`, `**Version:** {{version}}`) are no longer reported as the old header or as a conflict with the front matter (mlcext v0.9.2, B-MARKS-20261021-01).

### Changed
- Opened with a document (double click, "Open with", command line), MarkSafe starts in the preview; the editor comes on request. With the editor hidden, its toolbar button reads "Edit"; `Ctrl+E` (`⌘E`) shows and hides the editor. A new or empty file still opens in the editor, and the remembered editor setting stays as it was.

## [1.7.1] – 2026-10-21

### Fixed
- Without the PDF toolkit `mlcdoc` (e.g. the bare Linux binary), print, "Export › As PDF…" and the print button are no longer offered; `Ctrl+P` does nothing. HTML export stays available.

## [1.7.0] – 2026-10-21

MarkSafe becomes the editor for SOPs and forms of the MLC Sensor Monitor: documents written in Markdown turn into fillable PDF forms.

### Added
- **Fillable PDF forms**: form fields of the MLC Markdown Extension (`[@#name ____]`, check boxes, date and time patterns, presets, tooltips, fill hints) become a real PDF form. The extension is non-standard and documented in `docs/MLC-EXTENSIONS.md`.
- **PDF preview, export and printing** through the bundled MLC PDF toolkit `mlcdoc` (shipped with every installer):
  - page header with kind, ID and version
  - watermark as a user setting
  - the signature note of the Sensor Monitor
  - images from local files, base64 or whitelisted URLs
  - optional placeholders in the preview
  - Print (`Ctrl+P`) opens the same PDF in the system's PDF viewer.
- **Front matter as the single source** of `doc_id`, `version` and `title`. The variables `{{doc_id}}`, `{{version}}`, `{{title}}` are replaced when rendering. A document header in the preview. Quick fix for the old header line in the text.
- **Templates**: SOP, procedure, form, information sheet, a form example with all field types, README, CHANGELOG, empty document and minutes. Templates are JSON schema plus Markdown in a per-user folder, editable without an update; the document folder can override them.
- **Checks while you write** (Go module `mlcext`, shared with the Sensor Monitor):
  - form fields, ID and version, variables
  - signature
  - table columns too narrow in the PDF (with quick fix)
  - characters the PDF fonts lack
- **Editor**:
  - MLC highlighting and completion
  - alert box menu, symbol palette of characters that print in the PDF
  - table menu and column widths (toolbar, right click, a box at the table)
  - base64 images shown shortened
  - CodeMirror texts in German, Spanish and French
- **Document folder** with an overview of all documents (ID, kind, version, check result, file name hint).
- **Signed documents**: editing one asks before saving and removes the void signature.
- **File names** from the document ID: `SOP-QM-001.md`, exports `SOP-QM-001_v1.0.pdf`.
- Shared rendering packages on npm: `mlc-markdown` and `mlc-markdown-svelte`. The preview, the browse window and `marksafe_browse` use them, as will the Sensor Monitor.
- `task version:set|sync|check`: the version from `VERSION` in every file.

### Changed
- Export moved to *File → Export* (PDF, HTML); embedded HTML is shown sanitized.
- The browse button appears only for saved files; zoom only for the HTML preview.

### Fixed
- Paste, undo and cut acted twice on Windows (Wails passed the key on to WebView2 as well); Ctrl+S could open two save dialogs.
- Footnotes rendered before the text; table alignment ignored in the HTML preview; the PDF preview re-rendered every 800 ms.
- Installer and package metadata still carried version 1.5.1.
- HTML export offline: diagrams, formulas and local images are embedded.

## [1.6.1] – 2026-09-22

### Added
- **Browser mode**: open the active document's directory in a standalone window (menu *View → Browser-Modus*, `Ctrl/Cmd+Shift+B`, or the new toolbar button). Recursive directory tree, fuzzy full-text search (`Ctrl/Cmd+K`) with snippets, validated document links, and automatic directory indexes.
- Shared UI library (`frontend/src/shared/browse`) powering both the in-app browser mode and the standalone `marksafe_browse` tool: sidebar, search modal, theme toggle, info panel, external-link confirmation.
- Shared Go library (`internal/mdbrowse`): rooted file access with hard boundary, TOC scanning, title extraction, goldmark render pipeline, native fuzzy search.
- Document info panel (toolbar): title, relative path tooltip in the tree, file size, modification date, word/character counts, reading time.
- Adjustable line density (compact/comfortable/spacious) covering content and tree; resizable, collapsible sidebar persisted per profile.
- Window title of the browse window shows the current root directory.

### Fixed
- Theme toggle fully re-renders preview (Mermaid diagrams no longer flash red syntax-error boxes).
- Search no longer misses short queries: fixed slice-out-of-range panics and rune-misaligned highlighting in snippets (e.g. `roll` vs `Rollout`).
- Stale frontend bundles: the build fingerprint now covers the shared UI, and shipped assets use content hashes.
- Sidebar auto-expands and scrolls to the document opened via link or history.
- Native language menu marks the active language; editor hidden-state persists; About dialog follows the app theme (native dialog removed).

## [1.6.0] – 2026-09-18

### Added
- `marksafe_browse`: standalone CLI tool serving a Markdown directory over HTTP (`task dev`, binary in `bin/`), with interactive Svelte 5 frontend (dark/light/system themes, collapsible tree, search, validated links, directory fallback pages).
- Developer look & feel: Inter/JetBrains Mono typography, monochrome IDE palette, SVG icon set (no emoji), sticky toolbar, SEO meta tags.
- Link validation: missing/out-of-root targets render as non-clickable placeholders instead of 404s; images served via `/raw`.
- Forgiving parser: repairs glued headings/rules from pasted sources, truncates runaway titles.

### Fixed
- Svelte hydration errors (`mount` on server context), missing favicon route, lowercase TOC JSON contract.
- Intermittent blank pages: sequenced loader with stale-response dropping, abort handling, visible retry panel, per-document remount.
- `svelte-check` clean (0/0) in both frontends; Go vet clean.

## [1.5.1] and earlier

See [GitHub Releases](https://github.com/mlechner911/mdviewer/releases) for earlier versions.
