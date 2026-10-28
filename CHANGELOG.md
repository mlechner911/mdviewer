# Changelog

All notable changes to MarkSafe. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed
- **The connection to the MLC Sensor Monitor is an optional module.** Public builds no longer contain it; `task … SENSOR=1` builds it in (Go tag `sensormonitor`, frontend variant). Its code lives in `sensormonitor/` and `frontend/src/sensormonitor/`, its address in its own `sensormonitor.json` (taken over once from `config.json`). File › Settings… only appears when a module has settings. See `docs/SENSOR-MONITOR.md`.

### Added
- **Fill in a form** ("Ausfüllen" in the preview bar): input mask next to the editor, values straight into the fields, "Werte laden…" from JSON, saving as a filled-in form with the `meta:` block (who, when, which form version) under `<doc_id>_<date>.md`; the form is never overwritten. Date and time fields (`TT.MM.JJJJ`, `HH:MM`) use the system's pickers with "Heute"/"Jetzt"; an invalid earlier value is flagged. The mask follows the app theme (dark mode). A filled-in form's PDF is named after the day it was completed (`FB-LAG-001_2026-10-28.pdf`), not the form's version. `mlcext.SetFormRecord`, `FormRecordOf`, `FormRefOf`.
- **Signature field** `[@Unterschrift: Rolle]` / `[@#name Unterschrift: Rolle]` (English `Signature:`, as mlcmd2pdf 3.28.0): names who signs; checker, preview (line with the role below) and the editor's building blocks know it. Filling leaves it alone.
- **Ticked check box** `[@x]` / `[@#name x]` (as mlcmd2pdf 3.27.0): the checker and the preview know it, the preview draws a cross.
- **Filling forms** in `mlcext` (plan P-SENS-20261028-01): `FillFields` puts values into the fields (also unnamed ones, check boxes ticked by `x`), `FieldValues` reads them back, `FormSchema` describes the fields for an input mask with a label taken from the text around them; moved here from the Sensor Monitor.
- Front matter block **`meta:`** for the record data of a filled-in form (`schema: mlc-form-record/v1`), not printed; `IsFormRecord`.

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
