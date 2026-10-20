<script lang="ts">
  // <MlcMarkdown> — Markdown with the MLC Markdown Extension (non-standard),
  // rendered by mlc-markdown. Display only: validity comes from mlcext.
  //
  // Hooks connect what only the app knows (base paths, local resources,
  // event numbers, navigation); styling goes through --mlc-md-* variables.
  import { onDestroy, tick, untrack, type Snippet } from 'svelte';
  import { createRenderer, enhance, frontMatterEntries, kindFromRef, labelsFor, type DocMeta, type RenderOptions, type Signature } from 'mlc-markdown';

  type NavTarget = { href: string; kind: 'anchor' | 'external' | 'link' | 'ref'; event: MouseEvent };

  let {
    source = '',
    locale = 'de',
    theme,
    frontMatter = 'hide',
    docHeader = true,
    watermark = '',
    debounce = 0,
    breaks = false,
    rawHtml = 'drop',
    highlight,
    math,
    resolveUrl,
    linkify,
    loadKatex,
    loadMermaid,
    onNavigate,
    onScrollLine,
    onRendered,
    scrollContainer,
    header,
    class: className = '',
  }: {
    source?: string;
    /** Language of the document (alert titles, "Ausfülltipp"). */
    locale?: string;
    /** 'light' | 'dark'; unset inherits data-theme from an ancestor. */
    theme?: 'light' | 'dark';
    /** Front matter: hidden, or a small key/value box. */
    frontMatter?: 'hide' | 'box';
    /** Header from the front matter (title, doc_id, version, signature); only for documents with a doc_id. */
    docHeader?: boolean;
    /** Watermark text across the document (set by the host, e.g. by the version's status); empty for none. */
    watermark?: string;
    /** Wait this long (ms) after the last change of source before rendering. */
    debounce?: number;
    breaks?: boolean;
    rawHtml?: 'drop' | 'sanitize';
    highlight?: RenderOptions['highlight'];
    math?: RenderOptions['math'];
    resolveUrl?: RenderOptions['resolveUrl'];
    linkify?: RenderOptions['linkify'];
    /** Lazy KaTeX (formulas) and Mermaid (diagrams), loaded on first use. */
    loadKatex?: () => Promise<{ renderToString(tex: string, o?: Record<string, unknown>): string }>;
    loadMermaid?: () => Promise<{ render(id: string, src: string): Promise<{ svg: string }> }>;
    /** A click on a link or anchor; return true when the app handled it. */
    onNavigate?: (t: NavTarget) => boolean | void;
    /** The top visible source line while scrolling (scroll sync with an editor). */
    onScrollLine?: (line: number) => void;
    /** After every render (and enhancement) with the content element. */
    onRendered?: (el: HTMLElement) => void;
    /** The scrolling element; default: the nearest scrollable ancestor. */
    scrollContainer?: HTMLElement;
    /** Own header above the content, e.g. "how the Sensor Monitor reads it". */
    header?: Snippet<[{ frontMatter: string; entries: Array<[string, string]>; meta: DocMeta; signature: Signature | null }]>;
    class?: string;
  } = $props();

  let root = $state<HTMLElement>();
  let content = $state<HTMLElement>();
  let html = $state('');
  let fm = $state('');
  let meta = $state<DocMeta>({ docId: '', version: '', title: '' });
  let signature = $state<Signature | null>(null);
  const entries = $derived(frontMatterEntries(fm));
  const labels = $derived(labelsFor(locale));
  const kind = $derived(kindFromRef(meta.docId));
  // The title shows in the header unless the document has its own H1 (as
  // the templates do, for the PDF): never twice.
  const headerTitle = $derived(meta.title && !/<h1[\s>]/.test(html) ? meta.title : '');

  function fmtDate(iso: string): string {
    const d = new Date(iso);
    if (!iso || isNaN(d.getTime())) return iso;
    return d.toLocaleString(locale, { dateStyle: 'medium', timeStyle: 'short' });
  }
  const signedText = $derived(
    !signature
      ? ''
      : signature.fields.signed_by
        ? labels.signedBy.replace('{by}', signature.fields.signed_by).replace('{at}', fmtDate(signature.fields.signed_at ?? ''))
        : labels.signed.replace('{at}', fmtDate(signature.fields.signed_at ?? '')),
  );

  // Diagonal text as a repeated SVG background: covers every screen and print page.
  const watermarkUrl = $derived.by(() => {
    const t = watermark.trim().slice(0, 60);
    if (!t) return '';
    const esc = t.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="600"><text x="300" y="300" text-anchor="middle" dominant-baseline="middle" transform="rotate(-35 300 300)" font-family="Helvetica, Arial, sans-serif" font-size="${Math.max(28, Math.min(110, 900 / t.length))}" font-weight="700" fill="#808080">${esc}</text></svg>`;
    return `url("data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}")`;
  });

  const renderer = $derived(createRenderer({ locale, breaks, rawHtml, highlight, math, resolveUrl, linkify }));

  let timer: ReturnType<typeof setTimeout> | undefined;
  let first = true;
  $effect(() => {
    const r = renderer;
    const src = source;
    clearTimeout(timer);
    const run = () => untrack(() => void update(r.render(src)));
    if (first || !debounce) {
      first = false;
      run();
    } else timer = setTimeout(run, debounce);
  });

  async function update(res: { html: string; frontMatter: string; meta: DocMeta; signature: Signature | null }) {
    html = res.html;
    fm = res.frontMatter;
    meta = res.meta;
    signature = res.signature;
    await tick();
    if (!content) return;
    const needsMath = !math && content.querySelector('.mlc-md-math[data-tex]');
    const needsMermaid = content.querySelector('pre.mlc-md-mermaid');
    if ((needsMath && loadKatex) || (needsMermaid && loadMermaid)) {
      // A failing loader (or library) must not stop the rest: formulas,
      // diagrams and onRendered are independent of each other.
      async function load<T>(f?: () => Promise<T>): Promise<T | undefined> {
        try {
          return f ? await f() : undefined;
        } catch (err) {
          console.error('mlc-markdown: loading failed', err);
          return undefined;
        }
      }
      const [katex, mermaid] = await Promise.all([needsMath ? load(loadKatex) : undefined, needsMermaid ? load(loadMermaid) : undefined]);
      try {
        if (content) await enhance(content, { katex, mermaid });
      } catch (err) {
        console.error('mlc-markdown: enhance failed', err);
      }
    }
    if (content) onRendered?.(content);
  }

  function onClick(e: MouseEvent) {
    const a = (e.target as HTMLElement)?.closest?.('a');
    if (!a || !content?.contains(a)) return;
    const href = a.getAttribute('href') ?? '';
    const kind: NavTarget['kind'] = a.classList.contains('mlc-md-ref')
      ? 'ref'
      : href.startsWith('#')
        ? 'anchor'
        : /^https?:\/\//i.test(href)
          ? 'external'
          : 'link';
    if (onNavigate?.({ href, kind, event: e })) {
      e.preventDefault();
      return;
    }
    if (kind === 'anchor' && href.length > 1) {
      // in-document jump without changing the page URL
      const target = content.querySelector(`[id="${CSS.escape(decodeURIComponent(href.slice(1)))}"]`);
      if (target) {
        e.preventDefault();
        target.scrollIntoView({ block: 'start' });
      }
    }
  }

  // ---------- Scroll sync over data-line ----------
  function scroller(): HTMLElement | undefined {
    if (scrollContainer) return scrollContainer;
    for (let el = root?.parentElement; el; el = el.parentElement) {
      const o = getComputedStyle(el).overflowY;
      if (o === 'auto' || o === 'scroll') return el;
    }
    return document.scrollingElement as HTMLElement | undefined;
  }

  function blocks(): HTMLElement[] {
    return content ? Array.from(content.querySelectorAll<HTMLElement>('[data-line]')) : [];
  }

  const top = (el: HTMLElement, sc: HTMLElement) => el.getBoundingClientRect().top - sc.getBoundingClientRect().top + sc.scrollTop;

  /** Scroll so that source line `line` is at the top (interpolated between blocks). */
  export function scrollToLine(line: number) {
    const sc = scroller();
    const bs = blocks();
    if (!sc || !bs.length) return;
    let i = 0;
    while (i + 1 < bs.length && Number(bs[i + 1].dataset.line) <= line) i++;
    const a = bs[i];
    const b = bs[i + 1];
    let y = top(a, sc);
    if (b) {
      const la = Number(a.dataset.line);
      const lb = Number(b.dataset.line);
      if (lb > la && line > la) y += ((line - la) / (lb - la)) * (top(b, sc) - y);
    }
    suppress = true;
    sc.scrollTop = y;
    requestAnimationFrame(() => (suppress = false));
  }

  let suppress = false;
  let raf = 0;
  let attached: HTMLElement | undefined;
  function onScroll() {
    if (suppress || !onScrollLine || raf) return;
    raf = requestAnimationFrame(() => {
      raf = 0;
      const sc = attached;
      const bs = blocks();
      if (!sc || !bs.length) return;
      const y = sc.scrollTop;
      let i = 0;
      while (i + 1 < bs.length && top(bs[i + 1], sc) <= y) i++;
      const a = bs[i];
      const b = bs[i + 1];
      let line = Number(a.dataset.line);
      if (b) {
        const ya = top(a, sc);
        const yb = top(b, sc);
        if (yb > ya && y > ya) line += Math.round(((y - ya) / (yb - ya)) * (Number(b.dataset.line) - line));
      }
      onScrollLine(line);
    });
  }

  // (Re)attach whenever the scrolling element is known or changes.
  $effect(() => {
    void scrollContainer;
    if (!root) return;
    attached = scroller();
    const target = attached === document.scrollingElement ? window : attached;
    target?.addEventListener('scroll', onScroll, { passive: true });
    return () => target?.removeEventListener('scroll', onScroll);
  });
  onDestroy(() => clearTimeout(timer));
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div bind:this={root} class="mlc-md {className}" data-theme={theme} lang={locale} onclick={onClick}>
  {#if watermarkUrl}
    <div class="mlc-md-watermark" aria-hidden="true" data-text={watermark.trim().slice(0, 60)} style="background-image: {watermarkUrl}"></div>
  {/if}
  {#if docHeader && meta.docId}
    <header class="mlc-md-docheader">
      {#if kind}<div class="mlc-md-docheader-kind">{labels.kinds[kind]}</div>{/if}
      {#if headerTitle}<div class="mlc-md-docheader-title">{headerTitle}</div>{/if}
      <div class="mlc-md-docheader-meta">
        <span class="mlc-md-docheader-id">{meta.docId}</span>{#if meta.version}{` · ${labels.version} ${meta.version}`}{/if}
      </div>
      {#if signature}<div class="mlc-md-docheader-signed">{signedText}</div>{/if}
    </header>
  {/if}
  {#if header}
    {@render header({ frontMatter: fm, entries, meta, signature })}
  {/if}
  {#if frontMatter === 'box' && entries.length}
    <details class="mlc-md-frontmatter" open>
      <summary>Front Matter</summary>
      <table>
        <tbody>
          {#each entries as [k, v]}
            <tr><td>{k}</td><td>{v}</td></tr>
          {/each}
        </tbody>
      </table>
    </details>
  {/if}
  <div bind:this={content} class="mlc-md-content">{@html html}</div>
</div>
