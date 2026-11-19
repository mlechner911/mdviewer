# mlcext-editor-svelte

Svelte 5 component `<MlcEditor>`: a CodeMirror 6 Markdown editor for the
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
