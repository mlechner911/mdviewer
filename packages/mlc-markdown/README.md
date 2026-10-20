# mlc-markdown

Display of Markdown with the **MLC Markdown Extension** — the same look in
MarkSafe, the MLC Sensor Monitor and the MLC Terminal.

> [!IMPORTANT]
> The MLC Markdown Extension is **not standard Markdown**: form fields
> `[@…]`, field tooltips `<!--* … -->`, fill tips `<!--! … -->`, page breaks
> and anchor targets for documents of the MLC Sensor Monitor. See
> [docs/MLC-EXTENSIONS.md](https://github.com/mlechner911/mdviewer/blob/main/docs/MLC-EXTENSIONS.md).

**Display only.** Whether a document is valid is decided by the Go module
[`mlcext`](https://github.com/mlechner911/mdviewer/tree/main/mlcext); the
syntax is defined by mlcmd2pdf. Malformed fields stay text here. Both
recognise fields with the same test vectors (`mlcext/testdata/fields.json`).

```ts
import { createRenderer, enhance } from 'mlc-markdown';
import 'mlc-markdown/styles.css';

const r = createRenderer({
  locale: 'de',
  highlight: (code, lang) => hljs.getLanguage(lang) ? hljs.highlight(code, { language: lang }).value : undefined,
  resolveUrl: (url, kind) => (url.startsWith('/api/') ? base + url : url),
  linkify: (text) => splitEventRefs(text),
});
const { html, frontMatter, meta, signature } = r.render(markdown);
el.innerHTML = html;                       // sanitized by DOMPurify
await enhance(el, { katex, mermaid });     // formulas and diagrams, lazily loaded by the host
```

`npm install mlc-markdown` — dependencies: `marked`, `marked-footnote`,
`dompurify`. KaTeX, Mermaid and a highlighter are up to the host.

| Feature | Output |
| --- | --- |
| Form fields `[@…]` | `.mlc-md-field-check`, `-line`, `-multiline` (like the fillable PDF), `data-field` = field name |
| Tooltip `<!--* … -->` | `title` on the field + `.mlc-md-field-hint` (screen only) |
| Fill tip `<!--! … -->` | `<details class="mlc-md-tip">`, Markdown inside, not printed |
| Alerts `> [!WARNING]` | `.mlc-md-alert-warning`, title in the document language |
| Headings | `id` like mlcmd2pdf `heading_slug`, `{#id}` explicit, duplicates `-1` |
| Page break, `<a id>` | `.mlc-md-page-break` (page break in print), anchor target |
| Tables | full width with column shares when the separator row is longer than 72 characters (like mlcdoc) |
| Every block | `data-line` = line in the source file (scroll sync) |
| Math, Mermaid | `.mlc-md-math[data-tex]`, `pre.mlc-md-mermaid` → `enhance()` |
| Front matter | not shown; `frontMatter` (raw), `frontMatterEntries()` for a key/value display |
| `doc_id`, `version`, `title` | `meta` (front matter first, the legacy header line `**Dokumenten-ID:** …` as a fallback) |
| Variables `{{doc_id}}`, `{{version}}`, `{{title}}` | replaced with the front matter values (not inside code); unknown names stay visible |
| Signature `<!--mlc-signature … -->` (end of file; old form: `signature:` block in front) | not shown; `signature` = `{ form, fields }` (`signed_by`, `signed_at` …), not verified |
| Other raw HTML | dropped (`rawHtml: 'sanitize'` keeps it, cleaned by DOMPurify; its `href`/`src` go through `resolveUrl`, `srcset` is removed) — MarkSafe uses `sanitize` |

Styling: put `class="mlc-md"` on the container and set the `--mlc-md-*`
variables to your design; `data-theme="dark"` switches the defaults. No
Tailwind needed.

Also exported: `documentMeta`, `expandVariables`, `splitSignature`,
`kindFromRef`, `scanFields`, `headingSlug`, `labelsFor`. The Go module
`mlcext` and this package run the same test cases
(`mlcext/testdata/fields.json`, `mlcext/testdata/meta.json`).

License: MIT.
