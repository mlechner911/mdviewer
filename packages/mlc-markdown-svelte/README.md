# mlc-markdown-svelte

Svelte 5 component `<MlcMarkdown>` for Markdown with the **MLC Markdown
Extension** (non-standard; form fields `[@…]`, fill tips, alerts, anchors,
page breaks), built on [`mlc-markdown`](../mlc-markdown). Display only —
validity comes from the Go module `mlcext`.

```svelte
<script>
  import { MlcMarkdown } from 'mlc-markdown-svelte';
  import 'mlc-markdown/styles.css';
  let view;
</script>

<MlcMarkdown
  bind:this={view}
  source={markdown}
  locale="de"
  theme="dark"
  debounce={250}
  frontMatter="hide"
  resolveUrl={(url, kind) => url.startsWith('/api/') ? base + url : url}
  linkify={splitEventRefs}
  onNavigate={({ href, kind }) => kind === 'ref' ? (openEntity(href), true) : false}
  onScrollLine={(line) => editor.scrollToLine(line)}
  loadKatex={() => import('katex').then((m) => m.default)}
  loadMermaid={() => import('mermaid').then((m) => (m.default.initialize({ startOnLoad: false }), m.default))}
>
  {#snippet header({ frontMatter, entries })}
    <MyDocHeader {entries} />
  {/snippet}
</MlcMarkdown>
```

| Prop | Meaning |
| --- | --- |
| `source` | Markdown (with or without front matter) |
| `locale` | language of the document: alert titles, "Ausfülltipp" |
| `theme` | `light` / `dark`; unset inherits `data-theme` from an ancestor |
| `frontMatter` | `hide` (default) or `box` (key/value table, not printed) |
| `header` | snippet above the content, gets `{ frontMatter, entries }` |
| `debounce` | ms to wait after a change of `source` (live preview) |
| `highlight`, `math`, `resolveUrl`, `linkify` | hooks of `mlc-markdown` |
| `loadKatex`, `loadMermaid` | lazy loaders, called only when the document needs them |
| `onNavigate({ href, kind, event })` | `kind`: anchor, link, external, ref; return `true` when handled |
| `onScrollLine(line)` / `scrollToLine(line)` | scroll sync over `data-line` (source line), interpolated between blocks |
| `onRendered(el)` | after every render and enhancement |
| `scrollContainer` | the scrolling element (default: nearest scrollable ancestor) |
