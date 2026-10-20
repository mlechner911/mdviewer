<script lang="ts">
  // <MlcMarkdown> — Markdown with the MLC Markdown Extension (non-standard),
  // rendered by mlc-markdown. Display only: validity comes from mlcext.
  //
  // Hooks connect what only the app knows (base paths, local resources,
  // event numbers, navigation); styling goes through --mlc-md-* variables.
  import { onDestroy, onMount, tick, untrack, type Snippet } from 'svelte';
  import { createRenderer, enhance, frontMatterEntries, type RenderOptions } from 'mlc-markdown';

  type NavTarget = { href: string; kind: 'anchor' | 'external' | 'link' | 'ref'; event: MouseEvent };

  let {
    source = '',
    locale = 'de',
    theme,
    frontMatter = 'hide',
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
    header?: Snippet<[{ frontMatter: string; entries: Array<[string, string]> }]>;
    class?: string;
  } = $props();

  let root = $state<HTMLElement>();
  let content = $state<HTMLElement>();
  let html = $state('');
  let fm = $state('');
  const entries = $derived(frontMatterEntries(fm));

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

  async function update(res: { html: string; frontMatter: string }) {
    html = res.html;
    fm = res.frontMatter;
    await tick();
    if (!content) return;
    const needsMath = !math && content.querySelector('.mlc-md-math[data-tex]');
    const needsMermaid = content.querySelector('pre.mlc-md-mermaid');
    if ((needsMath && loadKatex) || (needsMermaid && loadMermaid)) {
      const [katex, mermaid] = await Promise.all([
        needsMath && loadKatex ? loadKatex() : undefined,
        needsMermaid && loadMermaid ? loadMermaid() : undefined,
      ]);
      if (content) await enhance(content, { katex, mermaid });
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
      if ((o === 'auto' || o === 'scroll') && el.scrollHeight > el.clientHeight) return el;
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

  onMount(() => {
    attached = scroller();
    const target = attached === document.scrollingElement ? window : attached;
    target?.addEventListener('scroll', onScroll, { passive: true });
    return () => target?.removeEventListener('scroll', onScroll);
  });
  onDestroy(() => clearTimeout(timer));
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div bind:this={root} class="mlc-md {className}" data-theme={theme} lang={locale} onclick={onClick}>
  {#if header}
    {@render header({ frontMatter: fm, entries })}
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
