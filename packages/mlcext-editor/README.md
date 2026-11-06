# mlcext-editor — CodeMirror 6 support for the MLC Markdown Extension

> [!IMPORTANT]
> The MLC Markdown Extension is **not standard Markdown**. See
> [docs/MLC-EXTENSIONS.md](../../docs/MLC-EXTENSIONS.md).

Editor support for the form extension, shared by every editor that writes
such documents (MarkSafe, the admin of the MLC Sensor Monitor). It
implements **no rules**: everything it shows comes from the Go checker
[`mlcext`](../../mlcext) (`mlcext.Analyze`), which the host hands in. MarkSafe
calls it over Wails, a web frontend through its server. There is one
implementation of the rules, and this package only displays its answer.

```ts
import { mlcExtension, mlcSnippets, insertMlcSnippet } from 'mlcext-editor';

const source = { analyze: (md: string) => fetch('/api/mlcext/analyze', { method: 'POST', body: md }).then((r) => r.json()) };

new EditorView({ extensions: [basicSetup, markdown(), mlcExtension(source, { onAnalysis: (a) => updateStatus(a) })] });
```

| Feature | Source |
| --- | --- |
| Highlighting of fields `[@…]`, tooltips `<!--* -->`, fill hints `<!--! -->` | field and hint positions from `Analyze` |
| Diagnostics with line, gutter and message | `Analyze().diagnostics` |
| Completion after `[@` and `<!--`: fields, date/time, multi-line, prefill fields | `Analyze().lists.prefill` |
| Completion in the front matter: keys and `event_kinds` | `Analyze().lists` |
| `mlcSnippets()` and `insertMlcSnippet()` for host menus; with a schema without the old prefill names | the same building blocks |
| Completion of a schema's properties after `[@#`, description on hover (`locale` option picks the language) | `Analyze().schema` |
| `fieldHint({ label, title, onEdit })`: a box after the field under the cursor, double click on a field; `onEdit(line, col)` opens the host's field dialog. `fieldAt()`, `fieldAtCursor()` | field positions from `Analyze` |

Plain Markdown is left alone: when `Analyze` does not detect an MLC
document, there is no highlighting and no diagnostic.

**Install:** the package ships as TypeScript source. CodeMirror is a peer
dependency; the host must resolve exactly one copy (Vite: `resolve.dedupe`).
