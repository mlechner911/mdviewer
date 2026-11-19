# Changelog

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
