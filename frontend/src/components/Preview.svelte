<script lang="ts">
  /**
   * Preview renders the active document with the shared component
   * <MlcMarkdown> (mlc-markdown), so a document looks the same in MarkSafe,
   * the Sensor Monitor and the MLC Terminal. MarkSafe adds its security
   * model through the component's hooks: local images only from
   * whitelisted folders (served by /local-resource, which checks again),
   * external images and links only for whitelisted domains — an image is
   * not requested before that check.
   */
  import 'katex/dist/katex.min.css';
  import 'mlc-markdown/styles.css';
  import hljs from 'highlight.js/lib/common';
  import hljsLight from 'highlight.js/styles/github.css?raw';
  import hljsDark from 'highlight.js/styles/github-dark.css?raw';
  import { MlcMarkdown } from 'mlc-markdown-svelte';
  import DocInfoCard from './DocInfoCard.svelte';
  import { BrowserOpenURL } from '../lib/wails';
  import * as backend from '../lib/backend';
  import type { Theme } from '../themes';
  import { locale } from '../i18n';

  let {
    source = '',
    theme,
    fontSize = 100,
    currentFilePath = null,
    analysis = null,
    onsecurity_request,
    onopen_file,
    onscrollline,
    onrendered,
  } = $props<{
    source: string;
    theme: Theme;
    fontSize?: number;
    currentFilePath?: string | null;
    /** mlcext.Analyze of the document: the "how the Sensor Monitor reads it" card. */
    analysis?: backend.MLCAnalysis | null;
    onsecurity_request?: (detail: { type: 'path' | 'url'; resource: string }) => void;
    onopen_file?: (detail: { path: string }) => void;
    onscrollline?: (line: number) => void;
    onrendered?: (el: HTMLElement) => void;
  }>();

  let container = $state<HTMLElement>();
  let view = $state<{ scrollToLine(line: number): void }>();
  const mlcDoc = $derived(!!analysis?.detected);

  // Images are not requested before the whitelist check: the renderer gets a
  // transparent placeholder carrying the real URL, checked after rendering.
  const PIXEL = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7';
  const MARK = '#mlc-img=';
  function resolveUrl(url: string, kind: 'link' | 'image'): string {
    if (kind === 'image' && url && !url.startsWith('data:')) return PIXEL + MARK + encodeURIComponent(url);
    return url;
  }

  function highlight(code: string, lang: string): string | undefined {
    if (!hljs.getLanguage(lang)) return undefined;
    return hljs.highlight(code, { language: lang, ignoreIllegals: true }).value;
  }

  async function baseDir(): Promise<string> {
    return currentFilePath ? await backend.getParentDir(currentFilePath) : '';
  }

  async function checkImages(root: HTMLElement) {
    for (const img of Array.from(root.querySelectorAll('img'))) {
      const src = img.getAttribute('src') ?? '';
      const at = src.indexOf(MARK);
      if (!src.startsWith(PIXEL) || at < 0) continue;
      const url = decodeURIComponent(src.slice(at + MARK.length));
      img.dataset.src = url;
      if (/^https?:\/\//i.test(url)) {
        try {
          const domain = new URL(url).hostname;
          if (await backend.isURLAllowed(domain)) img.src = url;
          else {
            img.style.display = 'none';
            onsecurity_request?.({ type: 'url', resource: domain });
          }
        } catch {
          img.style.display = 'none';
        }
        continue;
      }
      const abs = await backend.resolveRelativePath(await baseDir(), url);
      if (await backend.isPathAllowed(abs)) {
        img.src = '/local-resource?path=' + encodeURIComponent(abs); // the handler checks again
      } else {
        img.style.display = 'none';
        onsecurity_request?.({ type: 'path', resource: await backend.getParentDir(abs) });
      }
    }
  }

  function onNavigate({ href, kind }: { href: string; kind: string }): boolean {
    if (kind === 'anchor') return false; // the component jumps inside the document
    if (kind === 'external') {
      void (async () => {
        try {
          const domain = new URL(href).hostname;
          if (await backend.isURLAllowed(domain)) BrowserOpenURL(href);
          else onsecurity_request?.({ type: 'url', resource: domain });
        } catch {
          console.error('Invalid URL clicked:', href);
        }
      })();
      return true;
    }
    if (/\.(md|markdown)(#.*)?$/i.test(href)) {
      void (async () => {
        const abs = await backend.resolveRelativePath(await baseDir(), href.replace(/#.*$/, ''));
        if (await backend.isPathAllowed(abs)) onopen_file?.({ path: abs });
        else onsecurity_request?.({ type: 'path', resource: await backend.getParentDir(abs) });
      })();
      return true;
    }
    console.warn('Direct file links are blocked for security. Use Markdown files or explicit whitelist.');
    return true;
  }

  async function loadMermaid() {
    const { default: mermaid } = await import('mermaid');
    mermaid.initialize({ startOnLoad: false, theme: theme.mermaidTheme, themeVariables: theme.mermaidVars || {}, fontFamily: 'inherit' });
    return mermaid;
  }
  const loadKatex = async () => (await import('katex')).default;

  async function rendered(el: HTMLElement) {
    await checkImages(el);
    onrendered?.(el);
  }

  /** Scroll so that source line `line` is at the top. */
  export function scrollToLine(line: number) {
    view?.scrollToLine(line);
  }

  /**
   * getExportHTML returns the rendered document (diagrams, formulas, images)
   * as a self-contained element: local images become data: URIs, each
   * diagram keeps its Mermaid source in an inert <template>, formulas keep
   * their TeX in KaTeX's MathML annotation.
   */
  export async function getExportHTML(): Promise<string> {
    const root = container?.querySelector('.mlc-md');
    if (!root) return '';
    const clone = root.cloneNode(true) as HTMLElement;
    clone.removeAttribute('style');
    // Editing aids, not content.
    clone.querySelectorAll('.mlc-docinfo, .mlc-md-frontmatter').forEach((el) => el.remove());
    for (const img of Array.from(clone.querySelectorAll('img'))) {
      const src = img.getAttribute('src') ?? '';
      if (src.startsWith(PIXEL)) {
        img.remove(); // not allowed, never loaded
        continue;
      }
      if (!src.startsWith('/local-resource')) continue;
      try {
        const blob = await (await fetch(src)).blob();
        img.src = await new Promise<string>((resolve, reject) => {
          const reader = new FileReader();
          reader.onload = () => resolve(reader.result as string);
          reader.onerror = () => reject(reader.error);
          reader.readAsDataURL(blob);
        });
      } catch (err) {
        console.error('Export: image not embedded:', src, err);
      }
    }
    return clone.outerHTML;
  }

  /** The highlight.js stylesheet of the active theme (also for the export). */
  export function codeCSS(): string {
    return theme.id === 'dark' ? hljsDark : hljsLight;
  }
</script>

{@html '<' + 'style' + '>' + (theme.id === 'dark' ? hljsDark : hljsLight) + '</' + 'style' + '>'}

<div bind:this={container} class="preview-container flex-1 overflow-auto p-8 {theme.containerClass}" style="--preview-scale: {fontSize / 100};">
  {#key theme.id}
    <MlcMarkdown
      bind:this={view}
      {source}
      locale={$locale}
      theme={theme.id}
      frontMatter={mlcDoc ? 'hide' : 'box'}
      debounce={150}
      rawHtml="sanitize"
      {highlight}
      {resolveUrl}
      {onNavigate}
      {loadKatex}
      {loadMermaid}
      scrollContainer={container}
      onScrollLine={(line: number) => onscrollline?.(line)}
      onRendered={rendered}
    >
      {#snippet header()}
        {#if mlcDoc && analysis}<DocInfoCard info={analysis.document} />{/if}
      {/snippet}
    </MlcMarkdown>
  {/key}
</div>

<style>
  .preview-container :global(.mlc-md) { font-size: calc(1rem * var(--preview-scale, 1)); }

  @media print {
    :global(html), :global(body), :global(#app), :global(main), .preview-container {
      background: #ffffff !important;
      color: #111827 !important;
      height: auto !important;
      min-height: 0 !important;
      overflow: visible !important;
      display: block !important;
      width: 100% !important;
      margin: 0 !important;
      padding: 0 !important;
      border: none !important;
      box-shadow: none !important;
    }
  }
</style>
