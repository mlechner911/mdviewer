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
| Text field | `[@#name ______]` (width = number of `_`), multi-line over several lines | fillable text field | ✅ checked, highlighted; preview shows text |
| Default value | `[@#name ____value____]` | text field with a preset value | ✅ checked, highlighted; preview shows text |
| Pattern field | `[@#name TT.MM.JJJJ]`, `[@HH:MM]`, with default `=16\.10\.2026` | date/time field | ✅ checked, highlighted; preview shows text |
| Field tooltip | `<!--* hint -->` right after a field | tooltip in the PDF viewer, not printed | ✅ highlighted in the editor; hidden in the preview |
| Fill hint | `<!--! hint -->` | hint for the whole form, not printed | ✅ highlighted in the editor; hidden in the preview |
| Document id and version | front matter `doc_id: SOP-QM-001`, `version: "1.0"`, `title: …` | page header (via the layout file, see below) | ✅ checked; ⏳ document header in the preview |
| Legacy header line | `**Dokumenten-ID:** SOP-QM-001 \| **Version:** 1.0` in the text | read as a fallback only | ✅ warning `header-in-body`, error on a contradiction |
| Variables | `{{doc_id}}`, `{{version}}`, `{{title}}` in the text | replaced with the front matter values when rendering | ✅ checked; ⏳ replaced in the preview |
| Signature | `<!--mlc-signature … -->` at the end of the file (old form: a `signature:` block before the front matter) | seal of the Sensor Monitor; not part of the content | ✅ recognised, left out of the content; ⏳ "signed by … on …" in the preview |
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
  `version`, `description`, `event_kinds`. The kind (SOP, VA, FB, INFO)
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

- **PDF preview and export** (`mlcext/mlcdoc`, File → "Export as PDF…",
  `Ctrl+Shift+E`): for every document, rendered by the MLC PDF toolkit
  `mlcdoc`, which MarkSafe ships. Fields of the MLC extension look exactly as
  printed; MarkSafe does not re-implement the PDF layout. Front matter keys
  `mlcdoc` does not know (tags, event_kinds …) are left out and listed.
  Header, footer and logo of the Sensor Monitor's PDFs come from the server
  and are not part of the local PDF. Formulas are not typeset by `mlcdoc`
  yet; they appear as text, and MarkSafe says so.
- **Templates** (File → "New from Template…", `Ctrl+Shift+N`): a template
  is data, not code: a schema `<id>.json` (name, input fields, output file
  name) plus a Markdown body with `{{placeholders}}`. The dialog is built
  from the schema, so new kinds, fields or choice lists need a new file, not
  a new program. Built-in templates (SOP, VA, form, information sheet) are a
  starting set; a template with the same id in `_vorlagen/` of the document
  folder replaces it. Field types: `text`, `multiline`, `number`, `date`,
  `time` (also `hour`), `select`, `multiselect`, `refno` (document ID with
  prefix and a suggested next number), `docref` (other documents of the
  document folder, optional prefix filter, `multiple`). Options can be a
  list or a shared list (`event_kinds`). Placeholders: `{{id}}`,
  `{{id|yaml}}`, `{{id|list}}`, `{{id|ref}}` (docref: "ID (Titel)"). The schema of the schema file is
  [`mlcext/mlc-template.schema.json`](../mlcext/mlc-template.schema.json);
  editors validate templates with it. A created document must pass the
  checker; whether its ID is free among the approved documents only the
  Sensor Monitor can tell, on upload.
- **Document folder** (File → "Choose Document Folder…", "Open Document
  Folder", `Ctrl+Shift+D`): one fixed base folder for the document
  collection, like an Obsidian vault. It is a workspace only: status,
  versions and approval stay in the Sensor Monitor. The browse window shows
  a "Dokumente" overview (ID, kind, title, version, fields, check result).
  Own templates (schema + body, see above) go into `_vorlagen/` in that
  folder.

## Roadmap

1. A readable document header instead of the generic front matter table.
