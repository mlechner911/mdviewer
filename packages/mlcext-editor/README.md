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

## Findings and completions of the host

A host with rules of its own (the template workshop of the Sensor Monitor:
profile placeholders, zones, event kinds of the house) adds them without
touching the extension's:

```ts
import { mlcExtension, setMlcDiagnostics, type MlcCompletionContext } from 'mlcext-editor';

const view = new EditorView({
  extensions: [
    basicSetup,
    markdown(),
    mlcExtension(source, {
      completions: async (ctx: MlcCompletionContext) => {
        if (ctx.kind === 'placeholder') return profile.map((p) => ({ label: `{{${p.name}}}`, detail: p.label }));
        if (ctx.kind === 'front-matter-value' && ctx.key === 'zone') return zones.map((z) => ({ label: z.name, apply: z.id }));
        return null;
      },
    }),
  ],
});

// After the server checked the text:
setMlcDiagnostics(view, findings); // [{ line, col, severity, code, message, source: 'Werkstatt' }]
```

**`setMlcDiagnostics(view, list)`** shows the host's findings next to the
extension's. Same form as `mlcext` diagnostics: 1-based line, column in
code points, `severity` `error` / `warning` / `info`, `source` names the
checker in the tooltip. They move with the text while the author types, and
each call replaces the previous set (an empty list removes them). They show
without any analysis too, and setting them does not run `analyze` again.
`extraDiagnostics` (pulled on every analysis) stays for findings that
follow the analysis, like MarkSafe's folder check.

**`completions(context)`** adds the host's items to the list. `context.kind`
says where the cursor is; an item replaces the token being typed:

| `kind` | Where | Token (`prefix`) | Built-in items next to the host's |
| --- | --- | --- | --- |
| `placeholder` | `{{…` in the text | what follows `{{`; labels are written in full, `{{name}}` | `{{doc_id}}`, `{{version}}`, `{{title}}` |
| `field-name` | after `[@#` | the name typed so far | schema properties, field snippets |
| `event-kind` | `event_kinds: [` in the front matter | the word | `Analyze().lists.eventKinds` |
| `front-matter-key` | start of a front matter line | the key; without `apply` an item inserts `label: ` | `Analyze().lists.frontMatterKeys` |
| `front-matter-value` | after `key: ` in the front matter, `context.key` names it | the word; an empty one only on Ctrl+Space | none |

An item is `{ label, detail?, info?, apply?, type? }`; `apply` is the
inserted text (default the label). An item whose label the built-in list
already has is left out. `completions` may answer with a promise; when it
fails, the built-in list still shows. The context also carries `line`, `col`
(1-based, code points), `explicit` (Ctrl+Space) and the latest `analysis`.

**Install:** the package ships as TypeScript source. CodeMirror is a peer
dependency; the host must resolve exactly one copy (Vite: `resolve.dedupe`).
