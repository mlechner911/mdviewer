<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import '../markdown/document.css';
  // KaTeX's stylesheet comes from the host (browse server: /assets/katex.min.css,
  // MarkSafe: the preview bundle), so its fonts resolve where they live.
  import 'mlc-markdown/styles.css';
  import hljs from 'highlight.js/lib/common';
  import hljsLight from 'highlight.js/styles/github.css?raw';
  import hljsDark from 'highlight.js/styles/github-dark.css?raw';
  import { MlcMarkdown } from 'mlc-markdown-svelte';
  import { addCopyButtons, MERMAID_VARS } from './enhance';
  import type { BrowseClient, BrowseTheme, LinkTarget } from './browseClient';

  // Two modes: a document's Markdown source is rendered here with the
  // shared component <MlcMarkdown> (same output as the MarkSafe preview,
  // the Sensor Monitor and the MLC Terminal); directory listings come as
  // HTML from the host (renderDoc).
  let {
    htmlContent = '',
    source = null,
    docPath = '',
    client,
    theme = 'dark',
  }: {
    htmlContent?: string;
    source?: string | null;
    /** Root-relative path of the file source was read from (link base). */
    docPath?: string;
    client: BrowseClient;
    theme?: BrowseTheme;
  } = $props();

  const dispatch = createEventDispatcher();
  let docEl = $state<HTMLElement | null>(null);

  // Visible-text check: an "empty" render (frontmatter-only file,
  // tag-only fragment) shows the friendly placeholder instead of a
  // blank section that looks like a broken renderer.
  function hasVisibleText(html: string): boolean {
    if (!html) return false;
    return (
      html
        .replace(/<[^>]*>/g, '')
        .replace(/\s+/g, ' ')
        .trim().length > 0
    );
  }
  let showDoc = $derived(hasVisibleText(htmlContent));
  // Alert headings follow lang; until documents carry their own lang
  // (front matter), use the viewer's language.
  const lang = typeof navigator !== 'undefined' ? navigator.language.slice(0, 2) : 'en';

  // ---------- Source mode: hooks for <MlcMarkdown> ----------

  // Images are not requested before the host checked them: the renderer
  // gets a transparent placeholder carrying the written URL.
  const PIXEL = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7';
  const MARK = '#mlc-img=';
  function resolveUrl(url: string, kind: 'link' | 'image'): string {
    if (kind === 'image' && url && !url.startsWith('data:')) return PIXEL + MARK + encodeURIComponent(url);
    return url;
  }

  function highlight(code: string, language: string): string | undefined {
    if (!hljs.getLanguage(language)) return undefined;
    return hljs.highlight(code, { language, ignoreIllegals: true }).value;
  }

  const loadKatex = async () => (await import('katex')).default;
  async function loadMermaid() {
    const { default: mermaid } = await import('mermaid');
    const dark = theme === 'dark';
    mermaid.initialize({
      startOnLoad: false,
      theme: dark ? 'dark' : 'default',
      themeVariables: dark ? MERMAID_VARS.dark : MERMAID_VARS.light,
      fontFamily: 'inherit',
    });
    return mermaid;
  }

  const isWeb = (href: string) => /^(https?:)?\/\//i.test(href);

  // Tolerant anchor key, as mdbrowse.slugifyAnchor: lowercase, blanks,
  // underscores and hyphens to "-", letters and digits kept.
  function slugifyAnchor(s: string): string {
    let out = '';
    for (const ch of s.toLowerCase()) {
      if (/[\s_-]/.test(ch)) out += '-';
      else if (/[\p{L}\p{N}]/u.test(ch)) out += ch;
    }
    return out;
  }

  function markBroken(a: HTMLAnchorElement, href: string, reason: string) {
    a.removeAttribute('href');
    a.classList.add('broken-link');
    a.setAttribute('aria-disabled', 'true');
    a.title = reason;
    a.dataset.broken = href;
  }

  function brokenImage(img: HTMLImageElement, src: string, reason: string) {
    const span = document.createElement('span');
    span.className = 'broken-img';
    span.title = reason;
    span.textContent = 'Bild nicht gefunden: ' + (img.alt || src);
    img.replaceWith(span);
  }

  // Links and images get the same treatment as the host's renderDoc:
  // in-app navigation for documents and directories, files via the
  // host's raw URL, missing targets as non-clickable placeholders.
  async function rendered(el: HTMLElement) {
    const path = docPath;
    const links = Array.from(el.querySelectorAll<HTMLAnchorElement>('a[href]'));
    const imgs = Array.from(el.querySelectorAll<HTMLImageElement>('img'))
      .map((img) => {
        const src = img.getAttribute('src') ?? '';
        const at = src.indexOf(MARK);
        return src.startsWith(PIXEL) && at >= 0 ? { img, src: decodeURIComponent(src.slice(at + MARK.length)) } : null;
      })
      .filter((x): x is { img: HTMLImageElement; src: string } => !!x);

    // In-page anchors: exact id, else the slug-equivalent id, else broken.
    const ids = new Set(Array.from(el.querySelectorAll('[id]')).map((n) => n.id));
    const slugs = new Map<string, string>();
    for (const id of ids) {
      const s = slugifyAnchor(id);
      if (s && !slugs.has(s)) slugs.set(s, id);
    }

    const hrefs = new Set<string>();
    for (const a of links) {
      const href = a.getAttribute('href') ?? '';
      if (href.startsWith('#')) {
        const frag = href.slice(1);
        let id = frag;
        try {
          id = decodeURIComponent(frag);
        } catch {
          // keep as written
        }
        if (!frag || ids.has(frag) || ids.has(id)) continue;
        const canon = slugs.get(slugifyAnchor(frag));
        if (canon) a.setAttribute('href', '#' + canon);
        else markBroken(a, href, 'Anker nicht gefunden: ' + href);
      } else if (isWeb(href)) {
        // External links: recognizable, a new window, confirmed first.
        a.classList.add('external-link');
        a.target = '_blank';
        a.rel = 'nofollow noopener noreferrer';
      } else if (href) hrefs.add(href);
    }
    for (const { img, src } of imgs) {
      if (isWeb(src)) img.src = src;
      else hrefs.add(src);
    }

    let targets: Record<string, LinkTarget> = {};
    if (hrefs.size) {
      try {
        targets = await client.resolveLinks(path, [...hrefs]);
      } catch (err) {
        console.error('resolveLinks failed:', err);
      }
    }
    if (path !== docPath || !el.isConnected) return; // superseded

    for (const a of links) {
      const href = a.getAttribute('href') ?? '';
      const t = targets[href];
      if (!t || t.kind === 'keep') continue;
      if (t.kind === 'broken') markBroken(a, href, t.reason ?? 'Ziel nicht gefunden');
      else if (t.kind === 'md' && t.path) a.dataset.md = t.path;
      else if (t.kind === 'file' && t.path) {
        a.href = client.rawUrl(t.path);
        // Downloads keep the app alive and the real filename; images open.
        if (!t.image) a.download = t.path.split('/').pop() ?? '';
      }
    }
    for (const { img, src } of imgs) {
      if (isWeb(src)) continue;
      const t = targets[src];
      if (t?.kind === 'file' && t.path) img.src = client.rawUrl(t.path);
      else if (t?.kind === 'keep') img.src = src;
      else brokenImage(img, src, t?.reason ?? 'Bild nicht gefunden: ' + src);
    }
    addCopyButtons(el);
    dispatch('rendered', { text: el.textContent ?? '' });
  }

  // ---------- Clicks (both modes) ----------

  // In-app navigation: links the host validated (a[data-md]) open via
  // the client instead of a full page load (which would 404). Broken
  // links carry no href at all, external links confirm first.
  function onClick(e: MouseEvent) {
    const target = e.target as HTMLElement | null;
    const ext = target?.closest?.(
      'a.external-link[href],a[href^="http://"],a[href^="https://"]',
    ) as HTMLAnchorElement | null;
    if (ext && ext.href) {
      e.preventDefault();
      dispatch('external', { url: ext.href });
      return;
    }
    // Empty hrefs would reload the whole app (landing on the index
    // fallback); swallow them, there is nothing to navigate to.
    if (target?.closest?.('a[href=""]')) {
      e.preventDefault();
      return;
    }
    const anchor = target?.closest?.('a[data-md]') as HTMLAnchorElement | null;
    if (!anchor) return;
    e.preventDefault();
    const p = anchor.getAttribute('data-md');
    if (p) dispatch('open', { path: p });
  }

  // Attached imperatively (not via on:click): the section itself is NOT an
  // interactive element — the handler only delegates to dynamic links.
  // This keeps a11y checkers (correctly) quiet.
  onMount(() => {
    docEl?.addEventListener('click', onClick);
    return () => docEl?.removeEventListener('click', onClick);
  });
</script>

{#if source !== null}
  {@html '<' + 'style' + '>' + (theme === 'dark' ? hljsDark : hljsLight) + '</' + 'style' + '>'}
  <section class="doc mlc-doc" bind:this={docEl}>
    {#key theme}
      <MlcMarkdown
        {source}
        locale={lang}
        {theme}
        rawHtml="sanitize"
        {highlight}
        {resolveUrl}
        {loadKatex}
        {loadMermaid}
        onRendered={rendered}
      />
    {/key}
  </section>
{:else}
  <section class="doc md-doc" {lang} bind:this={docEl}>
    {#if showDoc}
      {@html htmlContent}
    {:else}
      <h1>MarkSafe Browse</h1>
      <p>Wähle ein Dokument aus dem Inhaltsverzeichnis links.</p>
    {/if}
  </section>
{/if}

<style>
  /* NOTE: class intentionally NOT named "content" — that name carries a
     margin-left rule for the fixed sidebar; nesting it doubled the offset. */
  .doc {
    width: 100%;
    max-width: 960px;
    padding: 2rem;
  }
  /* <MlcMarkdown> jumps to in-page anchors itself; every other click
     reaches the delegated handler on the section. */
  /* Line density of the browse toolbar drives the shared stylesheet. */
  .mlc-doc :global(.mlc-md) {
    --mlc-md-line-height: var(--lh-body, 1.5);
  }
  /* Pre-validated broken links and images: muted, dotted, not clickable. */
  .mlc-doc :global(a.broken-link),
  .mlc-doc :global(a.broken-link:hover) {
    color: var(--mlc-md-muted);
    text-decoration: underline dotted;
    cursor: not-allowed;
  }
  .mlc-doc :global(.broken-img) {
    display: block;
    margin: 1.25rem 0;
    padding: 1rem 1.25rem;
    border: 1px dashed var(--mlc-md-border);
    border-radius: var(--mlc-md-radius);
    color: var(--mlc-md-muted);
    font-family: var(--mlc-md-font-mono);
    font-size: 0.85rem;
  }
  .mlc-doc :global(pre) { position: relative; }
  .mlc-doc :global(pre .copy-btn) {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    font-family: var(--mlc-md-font-mono);
    font-size: 0.72rem;
    line-height: 1.4;
    padding: 0.2rem 0.55rem;
    border-radius: var(--mlc-md-radius);
    border: 1px solid var(--mlc-md-border);
    background: var(--mlc-md-surface);
    color: var(--mlc-md-muted);
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.15s, color 0.15s, border-color 0.15s;
  }
  .mlc-doc :global(pre:hover .copy-btn),
  .mlc-doc :global(pre:focus-within .copy-btn),
  .mlc-doc :global(pre .copy-btn.copied) { opacity: 1; }
  .mlc-doc :global(pre .copy-btn:hover) { color: var(--mlc-md-accent); border-color: var(--mlc-md-accent); }
  @media print { .mlc-doc :global(pre .copy-btn) { display: none; } }
</style>
