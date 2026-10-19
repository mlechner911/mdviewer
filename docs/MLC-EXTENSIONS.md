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
| Document header | `**Dokumenten-ID:** SOP-QM-001`, `**Version:** 1.0` | document id and version for the Sensor Monitor | ✅ checked; ⏳ readable header in the preview |
| Layout front matter | `header`, `footer`, `signatures`, `toc`, `page-size` … | page header/footer, signature block, layout | ⏳ generic metadata table |

✅ supported · ⏳ planned (see roadmap below)

### Rules

- Raw HTML: only the two forms above (page break, anchor target). MarkSafe
  drops any other raw HTML.
- Escape a literal `[@` as `\[@`.
- Fields cannot be placed in headings or image captions, and a field cannot
  span a blank line.
- Validity dates and approval status are **not** written in the document. The
  Sensor Monitor sets them on approval.

## Tooling

- **Checker** ([`mlcext`](../mlcext), Go module
  `github.com/mlechner911/mdviewer/mlcext`): validates the extension and
  reports problems with line and column. MarkSafe and the Sensor Monitor use
  the same module. MarkSafe shows the result in the status bar.
- **Editor support** ([`packages/mlcext-editor`](../packages/mlcext-editor),
  CodeMirror 6): highlights fields and hints, shows the checker's messages at
  their lines, and offers the building blocks after `[@` and in the editor's
  "MLC" menu. It appears only for documents that use the extension.

- **PDF preview** (`mlcext/mlcdoc`): the "PDF" button in the preview bar
  shows the pages as `mlcdoc` renders them, so fields look exactly as
  printed. MarkSafe does not re-implement the PDF layout; without `mlcdoc`
  installed the button is hidden. Header, footer and logo of the final PDF
  come from the Sensor Monitor and are not part of the local preview.

## Roadmap

1. A readable document header instead of the generic front matter table.
