# MLC Markdown Extension

> [!IMPORTANT]
> The MLC Markdown Extension is **not standard Markdown**. It is an
> MLC-specific extension for documents of the MLC Sensor Monitor (SOPs,
> forms, reports) that are turned into PDFs by the MLC PDF toolkit
> (`mlcdoc`). Files that use it are still valid Markdown. Other Markdown tools
> show the extension's syntax as plain text and hide its comments.

MarkSafe supports three levels of Markdown. This page explains which features
belong to which level, so it is always clear what is portable and what is not.

## 1. Standard Markdown

These have a published specification and work in every compliant tool.

| Feature | Specification |
| --- | --- |
| Paragraphs, headings, lists, links, images, code, quotes | [CommonMark](https://commonmark.org/) |
| Tables, task lists `- [ ]`, strikethrough, autolinks | [GitHub Flavored Markdown (GFM)](https://github.github.com/gfm/) |

## 2. Common extensions (widely supported, not part of a specification)

Many tools support these, but none of them is part of CommonMark or GFM. Most
of them come from GitHub, Pandoc or the KaTeX and Mermaid projects.

| Feature | Syntax | Origin |
| --- | --- | --- |
| Footnotes | `Text[^1]` … `[^1]: Note` | GitHub, Pandoc, PHP Markdown Extra |
| Alerts | `> [!NOTE]`, `[!TIP]`, `[!IMPORTANT]`, `[!WARNING]`, `[!CAUTION]` | GitHub |
| Math | `$…$`, `$$…$$` | KaTeX / MathJax convention |
| Diagrams | `` ```mermaid `` fenced block | Mermaid |
| Front matter | leading `---` YAML block `---` | Jekyll, Hugo, Obsidian … |
| Heading ids | `## Title {#anchor}` | Pandoc, kramdown, PHP Markdown Extra |
| Emoji | `:smile:` | GitHub |

Automatic heading anchors follow GitHub: lower case, space becomes `-`, letters
(umlauts included), digits, `-` and `_` are kept, and duplicates get `-1`,
`-2` and so on. `## Größe & Übersicht` gives `#größe--übersicht`.

## 3. MLC Markdown Extension (MLC-specific)

These features exist only for MLC documents. The PDF toolkit `mlcdoc` defines
them; MarkSafe displays them. Outside MLC tools they have no meaning.

| Feature | Syntax | Meaning in the PDF | MarkSafe |
| --- | --- | --- | --- |
| Page break | `<div class="page-break"></div>` | starts a new page | ✅ marker on screen, page break in print |
| Anchor target | `<a id="name"></a>` | link target inside the document | ✅ |
| Checkbox field | `[@#name]` | fillable checkbox | ✅ checked, highlighted in the editor; preview shows text |
| Ticked checkbox | `[@x]`, `[@#name x]` (also `X`) | checkbox that opens ticked (mlcmd2pdf 3.27.0) | ✅ checked, highlighted; preview shows a cross |
| Text field | `[@#name ______]` (width = number of `_`), multi-line over several lines | fillable text field | ✅ checked, highlighted; preview shows text |
| Default value | `[@#name ____value____]` | text field with a preset value | ✅ checked, highlighted; preview shows text |
| Pattern field | `[@#name TT.MM.JJJJ]`, `[@HH:MM]`, with default `=16\.10\.2026` | date/time field | ✅ checked, highlighted; preview shows text |
| Field tooltip | `<!--* hint -->` right after a field | tooltip in the PDF viewer, not printed | ✅ highlighted in the editor; hidden in the preview |
| Fill hint | `<!--! hint -->` | hint for the whole form, not printed | ✅ highlighted in the editor; hidden in the preview |
| Document id and version | front matter `doc_id: SOP-QM-001`, `version: "1.0"`, `title: …` | page header (via the layout file, see below) | ✅ checked; ⏳ document header in the preview |
| Legacy header line | `**Dokumenten-ID:** SOP-QM-001 \| **Version:** 1.0` in the text | read as a fallback only | ✅ warning `header-in-body`, error on a contradiction |
| Variables | `{{doc_id}}`, `{{version}}`, `{{title}}` in the text | replaced with the front matter values when rendering | ✅ checked; ⏳ replaced in the preview |
| Signature | `<!--mlc-signature … -->` at the end of the file (old form: a `signature:` block before the front matter) | seal of the Sensor Monitor; not part of the content | ✅ recognised, left out of the content; ⏳ "signed by … on …" in the preview |
| Signature field | `[@Unterschrift: Rolle]`, `[@#name Unterschrift: Rolle]` (English `Signature:`, either case) | room to sign above a line, the role below it; an empty PDF signature field (mlcmd2pdf 3.28.0) | ✅ checked, highlighted; preview shows the line and the role |
| Filled-in form | front matter `meta:` with `schema: mlc-form-record/v1`, values in the fields | not printed (mlcdoc does not read `meta`) | ✅ checked (`mlcext.FillFields`, `FieldValues`, `FormSchema`, `SetFormRecord`); fill mode with input mask |
| Layout front matter | `header`, `footer`, `signatures`, `toc`, `page-size` … | page header/footer, signature block, layout | ⏳ generic metadata table |

✅ supported · ⏳ planned (see roadmap below)

### Rules

- Raw HTML: the extension itself only defines the two forms above (page
  break, anchor target). MarkSafe shows other embedded HTML (e.g.
  `<details>`, `<kbd>`) after sanitizing it: scripts, event handlers, styles
  and form elements are removed. The PDF conversion (mlcdoc) may render
  embedded HTML differently or not at all; use it sparingly in documents
  meant for PDF or the Sensor Monitor.
- Escape a literal `[@` as `\[@`.
- Fields cannot be placed in headings or image captions, and a field cannot
  span a blank line.
- Validity dates and approval status are **not** written in the document. The
  Sensor Monitor sets them on approval.
- **The front matter is the only place for id, version and title** (`doc_id`,
  `version`, `title`, like Hugo and Jekyll). Allowed keys: `doc_id`, `title`,
  `version`, `description`, `event_kinds`, `meta` (see below). The kind (SOP, VA, FB, INFO)
  follows from the `doc_id` prefix. The header line in the text is the legacy
  form: still read, but it gives a warning; when both say different things,
  that is an error.
- **Variables:** only `{{doc_id}}`, `{{version}}` and `{{title}}`, no logic.
  They stay in the file and are replaced when the document is rendered, so
  they stay right after a new version. Any other `{{…}}` is an error
  (`unknown-variable`, or `template-placeholder-left` for a template
  placeholder that was not filled in). Not replaced inside code.
- **Signature:** the Sensor Monitor seals a released document with a comment
  at the very end:

  ```markdown
  <!--mlc-signature
  doc_no: SOP-2026-0013
  meaning: approved
  signed_by: "Dr. Anna Berger"
  signed_at: 2026-10-19T14:02:11Z
  sha256: …
  alg: ed25519
  key_id: …
  sig: …
  -->
  ```

  The checksum covers the file before `<!--mlc-signature`, byte for byte; the
  comment starts on a line of its own. Editing a signed file voids the
  signature (it becomes a new version). Older documents carry the signature
  as a `signature:` front matter block before the document's own front
  matter; both forms are read (`mlcext.SplitSignature`). Not to be confused
  with `signatures:` in the layout front matter of mlcdoc: those are the
  printed signature lines.
- **Filling a form** (`mlcext.FillFields`, plan P-SENS-20261028-01): the
  values go **into the fields**, so a filled-in form is still the same form
  and renders as PDF/A with the values in its fields:
  `[@#datum TT.MM.JJJJ=28\.10\.2026]`, `[@#ort __Kühlraum 2____]`,
  `[@#geprueft x]`. Every field can be filled, also one without a name (by
  its generated name `<section>_<n>`). A check box is ticked by `x`
  (also `true`, `1`, `ja`, `yes`), anything else unticks it. A text field
  keeps its width; more lines than the field has make it longer. Every
  ASCII punctuation character of a value is escaped, so no value can close a
  field or open one. `mlcext.FieldValues` reads the values back (`x` or
  empty for a check box), `mlcext.FormSchema` describes the fields for an
  input mask: kind, width, pattern, tooltip, section (heading), and a label
  taken from the text around the field (before it on the line; after a
  check box; table header plus row, e.g. "Charge (1)"; for a field alone on
  its line the line above). The label is a guess from the text, so a mask
  shows the field name as well.
- **Signature fields** name **who** signs, not how: on paper today, later
  with the signing PIN in the Sensor Monitor or on a signature pad. Write
  one per person (`[@#sig_vernichtend Unterschrift: Vernichtende Person]`,
  `[@#sig_zeuge1 Unterschrift: 1. Zeuge]`), each with a name, so a record
  can say which signature is missing. Filling never touches them
  (`FillFields` skips them, `FieldValues` leaves them out); `FormSchema`
  reports them as kind `signature` with their role. Without the colon,
  `[@Unterschrift]` stays a pattern field.
  - **Several people, later (concept):** the signatures are collected in the
    Sensor Monitor per role (person, time, method; for a signature pad the
    biometric data encrypted and bound to the content's checksum) and
    listed in `meta.signatures` of the record. The PDF gets **one** seal
    when the record is completed; each signature field shows who signed
    and when. One PDF signature per person would append a revision per
    signer, which some viewers show as invalid and which the archive rule
    "a sealed PDF never changes" forbids; that stays an option for a
    customer who requires it.
- **Record data of a filled-in form** go into the front matter block
  `meta:`, which mlcdoc reserves for data it does not render (the PDF is the
  same with and without it):

  ```yaml
  meta:
    schema: mlc-form-record/v1
    form: { doc_id: FB-LAG-001, version: "2.0", doc_no: SOP-2026-0031, sha256: … }
    context: { event_no: EV-2026-2069 }   # or { zone: …, sop: … } for a standalone record
    filled_by: Dr. Anna Berger            # display name, never an e-mail address
    completed_at: 2026-10-28T09:12:00+01:00
  ```

  `form` pins the exact version filled in. `meta` must be a block; an
  unknown `meta.schema` gives a warning (`frontmatter-meta-schema`).
  `mlcext.IsFormRecord` tells a filled-in form from a blank one.
- **Watermark** is not part of the document (`watermark` in the front matter
  is an error). It belongs to printing: the Sensor Monitor sets it by the
  version's status, MarkSafe by its setting (off by default). Both pass it to
  mlcdoc in the layout file (`mlcext.LayoutYAML`), next to the page header
  with id and version.

## Tooling

- **Checker** ([`mlcext`](../mlcext), Go module
  `github.com/mlechner911/mdviewer/mlcext`): validates the extension and
  reports problems with line and column. MarkSafe and the Sensor Monitor use
  the same module. MarkSafe shows the result in the status bar.
- **Editor support** ([`packages/mlcext-editor`](../packages/mlcext-editor),
  CodeMirror 6): highlights fields and hints, shows the checker's messages at
  their lines, and offers the building blocks after `[@` and in the editor's
  "MLC" menu. It appears only for documents that use the extension.

- **PDF preview, export and printing** (`mlcext/mlcdoc`; File → Export →
  "As PDF…", `Ctrl+Shift+E`; File → "Print…", `Ctrl+P`): for every
  document, rendered by the MLC PDF toolkit `mlcdoc`, which MarkSafe ships.
  Fields of the MLC extension look exactly as printed; MarkSafe does not
  re-implement the PDF layout.
  - Documents of the MLC extension get a page header with kind, id and
    version (`mlcext.LayoutYAML`, passed to mlcdoc as a layout file; the
    document stays unchanged). Variables are replaced. A signature that
    still matches the content is printed as mlcdoc's note "Elektronische
    Signatur"; an edited document gets none.
  - **Watermark:** a text of the user's setting (empty by default), set in
    the bar of the PDF preview; it applies to preview, export and printing.
  - **Images:** local files and base64 (`data:`) images are embedded. mlcdoc
    loads nothing from the network, so MarkSafe fetches `http(s)` images
    itself, only from whitelisted hosts (as the HTML preview); others become
    mlcdoc's placeholder "Abbildung nicht verfügbar". The PDF preview can
    show placeholders for all images (switch "Bilder" in its bar); export
    and print always embed them. Raw HTML `<img>` is not set by mlcdoc; use
    `![Text](bild.png)`.
  - **Table column widths** follow mlcmd2pdf: a separator row up to 72
    characters sizes the columns by content; a longer one makes the table
    full width with shares in proportion to the dashes. The check warns
    (`table-column-narrow`) when a column gets too narrow for its text; a
    right click in a table offers "automatic" (`---` per column) and "full
    width, shares by content" (`mlcext.FitTableColumns`).
  - **Symbols and emoji:** the PDF fonts are Liberation Sans and DejaVu Sans.
    ✅ ❌ ❎ and the colored dots 🔴🟠🟡🟢🔵🟣🟤 become colored symbols; many
    symbols print as they are (☐ ☑ ⚠ ℹ ★ ⚙ ⚡ ☎ ✉, arrows, shapes, some
    faces); pictographs such as 🧠 📌 🔒 print as an empty box. The check
    reports them (`pdf-glyph-missing`); the editor's symbol menu offers only
    characters that print (`mlcext.PDFSymbols`).
  - **Printing** renders the same PDF and opens it in the system's PDF
    viewer, which prints it, so the paper looks like the export.
  - Front matter keys `mlcdoc` does not know (tags, event_kinds …) are left
    out and listed. Organization and logo in the Sensor Monitor's PDFs come
    from the server. Formulas are not typeset by `mlcdoc` yet; they appear
    as text, and MarkSafe says so.
- **Templates** (File → "New from Template…", `Ctrl+Shift+N`): a template
  is data, not code: a schema `<id>.json` (name, input fields, output file
  name) plus a Markdown body with `{{placeholders}}`. The dialog is built
  from the schema, so new kinds, fields or choice lists need a new file, not
  a new program.
  - **Where they live:** MarkSafe copies the built-in templates once into
    the user's template folder (`%APPDATA%\marksafe\vorlagen`, on Linux
    `~/.config/marksafe/vorlagen`; the dialog has "Open template folder").
    There they can be changed, deleted or added to without a new release.
    An update never overwrites them; a template that is new in the program
    is added, a deleted one stays deleted. A template with the same id in
    `_vorlagen/` of the document folder replaces the one from the user
    folder (shared in a team).
  - **Built-in:** documents of the MLC extension (SOP, VA, form,
    information sheet) and general Markdown (README, CHANGELOG after "Keep a
    Changelog", empty document, minutes). `"category": "mlc"` or
    `"general"` groups them in the dialog.
  - **Front matter:** fields with `"target": "frontmatter"` are written into
    the front matter; MLC templates put `doc_id`, `version`, `title`,
    `description` and `event_kinds` there, and their body uses
    `{{doc_id}}`, `{{version}}`, `{{title}}`, which stay in the document. A
    template in the old form (header line in the text) still works and gets
    a warning.
  - Field types: `text`, `multiline`, `number`, `date`,
  `time` (also `hour`), `select`, `multiselect`, `refno` (document ID with
  prefix and a suggested next number), `docref` (other documents of the
  document folder, optional prefix filter, `multiple`). Options can be a
  list or a shared list (`event_kinds`). Placeholders: `{{id}}`,
  `{{id|yaml}}`, `{{id|list}}`, `{{id|ref}}` (docref: "ID (Titel)"). The schema of the schema file is
  [`mlcext/mlc-template.schema.json`](../mlcext/mlc-template.schema.json);
  editors validate templates with it. A created document must pass the
  checker; whether its ID is free among the approved documents only the
  Sensor Monitor can tell, on upload.
- **Filling in a form** ("Ausfüllen" in the preview bar, for a document
  with form fields): the form opens in a new tab without a file, and an
  input mask takes the place of the preview. Every entry goes straight into
  the fields of the Markdown (`mlcext.FillFields`), so editor, preview and
  PDF show the record as it will be saved. The labels come from the text
  around each field (`mlcext.FormSchema`), the field name stands next to
  them. "Werte laden…" sets values from a JSON file (`{"feld": "Wert"}`,
  `true`/`false` for a check box). Signature fields are shown, not filled.
  - **Saving** asks who filled it in (a name, never an e-mail address;
    remembered for next time), writes the `meta:` block (form with id,
    version and SHA-256, `filled_by`, `completed_at`) and suggests
    `<doc_id>_<date>.md`. A file that is not itself a filled-in form, the
    form above all, is never overwritten. A saved record opens like any
    document and can be filled on in its own tab.
  - Prefilling from an event of the Sensor Monitor follows with the server
    side (plan P-SENS-20261028-01 Phase 3).
- **Document folder** (File → "Choose Document Folder…", "Open Document
  Folder", `Ctrl+Shift+D`): one fixed base folder for the document
  collection, like an Obsidian vault. It is a workspace only: status,
  versions and approval stay in the Sensor Monitor. The browse window shows
  a "Dokumente" overview (ID, kind, title, version, fields, check result).
  Templates for the team (schema + body, see above) go into `_vorlagen/`
  in that folder.

## Roadmap

1. A readable document header instead of the generic front matter table.
