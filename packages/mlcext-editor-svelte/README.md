# mlcext-editor-svelte

Svelte 5 components `<MlcEditor>`, `<MlcFieldDialog>` and `<MlcImageDialog>`.
`<MlcEditor>` is a CodeMirror 6 Markdown editor for the
**MLC Markdown Extension** (non-standard; form fields `[@…]`, fill tips,
front matter of the MLC Sensor Monitor), built on
[`mlcext-editor`](../mlcext-editor). It implements **no rules**: the
analysis comes from the Go checker `mlcext` through the host, and the
host's own findings and completions come in through props. MarkSafe's
editor is built on it.

```svelte
<script>
  import { MlcEditor } from 'mlcext-editor-svelte';
  import { MlcMarkdown } from 'mlc-markdown-svelte';

  let text = $state(initial);
  let editor, preview;
  const source = { analyze: (md) => fetch('/api/mlcext/analyze', { method: 'POST', body: md }).then((r) => r.json()) };
</script>

<MlcEditor
  bind:this={editor}
  bind:value={text}
  theme="dark"
  locale="de"
  {source}
  completions={(ctx) => (ctx.kind === 'placeholder' ? profilePlaceholders : null)}
  diagnostics={serverFindings}
  onEditField={(line, col) => openFieldDialog(line, col)}
  onScrollLine={(line) => preview.scrollToLine(line)}
/>
<MlcMarkdown bind:this={preview} source={text} onScrollLine={(line) => editor.scrollToLine(line)} />
```

| Prop | Meaning |
| --- | --- |
| `value` | the Markdown text; `bind:value` keeps it in sync both ways |
| `theme` | `light` (default) / `dark` |
| `locale` | language of schema descriptions and of the field box |
| `source` | `{ analyze(md) }`: `mlcext.Analyze` of the host. Without it, completion and `diagnostics` still work |
| `completions(context)` | items of the host (placeholders, field names, event kinds, front matter keys and values), see `mlcext-editor` |
| `diagnostics` | findings of the host (`{ line, col, severity, code, message, source }`), shown next to the extension's; replaced when the array changes |
| `extraDiagnostics(md)` | findings pulled on every analysis; `refresh()` shows changes at once |
| `onAnalysis(a)` | every fresh analysis (status bar, menus) |
| `onEditField(line, col)` | the host's field dialog: box at the field under the cursor, double click; `fieldHintLabel`, `fieldHintTitle` replace its texts |
| `placeholder`, `readOnly`, `lineWrapping` (default on) | |
| `extensions` | CodeMirror extensions of the host (menus, folding, phrases); reconfigured when the array changes |
| `onchange(value)` | after every edit |
| `onScrollLine(line)` | top visible source line while the user scrolls, for `<MlcMarkdown>` |
| `onReady(view)` | the `EditorView`, once |

| Method | |
| --- | --- |
| `scrollToLine(line)`, `getTopLine()` | scroll sync; `scrollToLine` does not echo `onScrollLine` |
| `insertText(text)`, `replaceRange(line, col, endLine, endCol, text)`, `replaceOffsets(from, to, text)` | positions as in the analysis (1-based, code points, end exclusive) |
| `wrapSelection(prefix, suffix?)`, `prefixSelection(prefix)` | toolbar helpers (`**`, `## `) |
| `insertSnippet(snippet)` | a building block from `mlcSnippets()` |
| `refresh()` | analyse again now |
| `getView()`, `focus()` | |

Search (Ctrl+F), folding, bracket closing and the rest of CodeMirror's
`basicSetup` are on.

## Dialogs

The insert dialogs of MarkSafe, without a desktop backend: everything that
needs the checker or the file system is a callback of the host.

```svelte
<script>
  import { MlcFieldDialog, MlcImageDialog } from 'mlcext-editor-svelte';
  // Editing a field (onEditField of <MlcEditor>): the host's mlcext.FieldAt
  // gives its spec and range, e.g. fieldDialog = { spec, range: [line, col, endLine, endCol] }.
</script>

{#if fieldDialog}
  <MlcFieldDialog
    initial={fieldDialog.spec}
    format={(spec) => api.post('/mlcext/format-field', spec)}
    usedNames={analysis.fields.filter((f) => f.named).map((f) => f.name)}
    theme="dark"
    locale="de"
    onApply={(text) => (fieldDialog.range ? editor.replaceRange(...fieldDialog.range, text) : editor.insertText(text))}
    onClose={() => (fieldDialog = null)}
  />
{/if}

{#if imageDialog}
  <MlcImageDialog locale="de" onApply={(text) => editor.insertText(text)} onClose={() => (imageDialog = false)} />
{/if}
```

**`<MlcFieldDialog>`**: kind, name, width, lines, pattern, role, preset,
tooltip; shows the written field and how it renders.

| Prop | |
| --- | --- |
| `format(spec)` | **required**: `mlcext.FormatField` of the host → `{ text, error }`. The dialog knows no syntax; a late answer to an older input is dropped |
| `initial` | the field being edited (`mlcext.FieldAt().spec`); unset inserts a new one |
| `usedNames` | other field names in the document (duplicate warning) |
| `prefill`, `properties` | `{ name, label }`: prefill names and schema properties, offered for the name |
| `preview` | show the field as it renders (`mlc-markdown`, default on) |
| `onApply(text)`, `onClose()` | |

**`<MlcImageDialog>`**: a chosen, dropped or pasted image, as a reference
or embedded, with description and size (`Logo|200`, `|50%`); `existing`
edits an image of the document and keeps its file.

| Prop | |
| --- | --- |
| `choose()` | the host's file choice → `MlcImageChoice` or null. Default `chooseImageFile()`: the browser's file input, read as data: URL (embedded only) |
| `chosen` | a file the user already gave (dropped), or a promise of it |
| `dataUri` | a picture from the clipboard |
| `existing` | `{ text, alt, target, altFrom, altTo }` of an image in the document |
| `preferEmbed` | start with "embedded" even when a reference works |
| `toMarkdown(request)` | the Markdown for `{ choice, dataUri, alt, embed }`. Default: `![alt](data: URL)` or `![alt](reference)` |
| `onApply(text)`, `onClose()` | |

`MlcImageChoice` is `{ name, size, embeddable, reference?, dataUri?, path? }`:
`reference` when a relative path works (the PDF reads nothing outside the
document's folder), `path` is the host's own handle.

Both take `theme`, `locale` and `labels`. German and English are built in
(`DIALOG_LABELS`); `labels` replaces any of them, e.g. for other languages
(`%s` stands for the inserted value). Colours are CSS variables
`--mlc-dialog-*`: `bg`, `fg`, `border`, `button`, `button-hover`,
`input-bg`, `input-border`, `primary`, `primary-hover`, `overlay`, `warn`,
`error`, `font`; the fallbacks are MarkSafe's.

## Colours

Every colour is a CSS variable `--mlc-editor-<name>`, read where the editor
sits, so a host sets its palette on any ancestor without rebuilding the
editor. Unset, each theme falls back to MarkSafe's colours (`PALETTE`
exports them):

`fg`, `bg`, `caret`, `selection`, `gutter-bg`, `gutter-fg`,
`gutter-border`, `active-line`, `active-gutter-bg`, `active-gutter-fg`,
`placeholder`, `h1` … `h4`, `strong`, `emphasis`, `code`, `code-bg`,
`string`, `comment`, `link`, `url`, `quote`, `list`; plus `font` and
`font-size`.

```css
.werkstatt { --mlc-editor-bg: #fbfaf7; --mlc-editor-h1: #8a3b12; }
```

## Install

The package ships as source (Svelte and JavaScript). CodeMirror,
`mlcext-editor` and Svelte are peer dependencies; the host must resolve
exactly one copy of each (Vite: `resolve.dedupe` for `@codemirror/*`,
`codemirror`, `@lezer/highlight`, `svelte`). Before it is on npm, include
it as source like MarkSafe does (`frontend/vite.config.ts`: aliases for
`mlcext-editor` and `mlcext-editor-svelte`).
