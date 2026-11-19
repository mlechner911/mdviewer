# Changelog

## [0.2.0] - 2026-11-19

P-MARKS-20261117-01, phase 4.

### Added

- `<MlcFieldDialog>`: insert or edit a form field; the host writes it
  through `format` (mlcext.FormatField).
- `<MlcImageDialog>`: insert or edit an image, chosen (`choose`, default the
  browser's file input), dropped (`chosen`) or pasted (`dataUri`), as a
  reference or embedded; `toMarkdown` of the host or the built-in one.
- Texts German and English built in, `labels` for others; colours as CSS
  variables `--mlc-dialog-*`. Helpers `chooseImageFile`, `imageMarkdown`,
  `dialogTexts`, `formatBytes`.

## [0.1.0] - 2026-11-19

First version, included as source (P-MARKS-20261117-01, phase 2).

### Added

- `<MlcEditor>` (Svelte 5): CodeMirror 6 with Markdown and `mlcext-editor`;
  `bind:value`, light and dark theme, `source`, host `completions` and
  `diagnostics`, `extraDiagnostics`, field box (`onEditField`), host
  `extensions`, placeholder, read-only, line wrapping.
- Scroll sync with `<MlcMarkdown>`: `onScrollLine`, `scrollToLine`,
  `getTopLine`.
- Editing helpers: `insertText`, `replaceRange`, `replaceOffsets`,
  `wrapSelection`, `prefixSelection`, `insertSnippet`, `refresh`.
- Colours as CSS variables `--mlc-editor-*`, falling back to MarkSafe's
  palette (`PALETTE`).
