# Changelog

## [0.1.1] - 2026-10-21

### Fixed

- Header lines made of variables (`**Dokumenten-ID:** {{doc_id}}`) no longer
  count as id, version or title; they show the front matter.
- Tables keep the alignment of their separator row (`:--`, `:-:`, `--:`) in
  `styles/mlc-markdown.css`; before, every cell was left-aligned.

## [0.1.0] - 2026-10-20

First release.

### Added

- Renderer (`createRenderer`, `renderMarkdown`) for Markdown with the MLC
  Markdown Extension: form fields, field tooltips, fill tips, alerts,
  heading anchors like mlcmd2pdf, page breaks, anchor targets, table column
  widths, footnotes, `data-line` per block, DOMPurify.
- Front matter as the only source of `doc_id`, `version`, `title`
  (`documentMeta`), variables `{{doc_id}}`, `{{version}}`, `{{title}}`
  (`expandVariables`), the Sensor Monitor's signature in both forms
  (`splitSignature`).
- `enhance()` for KaTeX and Mermaid, loaded by the host.
- Stylesheet `mlc-markdown/styles.css` with `--mlc-md-*` variables, light
  and dark, print rules, document header and watermark.
