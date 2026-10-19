<script lang="ts">
  /**
   * Preview component handles the rendering of Markdown-derived HTML,
   * syntax highlighting, Mermaid diagrams, and KaTeX math.
   * Refactored for Svelte 5 Runes and clean Dark/Light mode support.
   */
  import { tick, untrack } from 'svelte';
  import mermaid from 'mermaid';
  import katex from 'katex';
  import 'katex/dist/katex.min.css';
  import '../shared/markdown/document.css';
  import renderMathInElement from 'katex/dist/contrib/auto-render';
  import { BrowserOpenURL } from '../lib/wails';
  import * as backend from '../lib/backend';
  import type { Theme } from '../themes';
  import { locale } from '../i18n';

  // --- Svelte 5 Runes: Props ---
  let { 
    html, 
    css = "", 
    theme, 
    fontSize = 100, 
    currentFilePath = null,
    onsecurity_request,
    onopen_file,
    onscroll
  } = $props<{
    html: string;
    css?: string;
    theme: Theme;
    fontSize?: number;
    currentFilePath?: string | null;
    onsecurity_request?: (detail: { type: 'path' | 'url', resource: string }) => void;
    onopen_file?: (detail: { path: string }) => void;
    onscroll?: (e: Event) => void;
  }>();

  let previewContainer = $state<HTMLElement>();

  // Theme-switch blending: dim the article while freshly themed markup
  // is being re-rendered, so the swap reads as a soft fade instead of
  // a flash. Layout is untouched (opacity only), so nothing reflows.
  let themeFading = $state(false);
  let fadeTimer: ReturnType<typeof setTimeout> | null = null;

  $effect(() => {
    theme; // subscribe to theme switches
    if (!html) return; // nothing displayed, nothing to blend
    themeFading = true;
    if (fadeTimer) clearTimeout(fadeTimer);
    // Safety: never stay dimmed if the re-render fails silently.
    fadeTimer = setTimeout(() => {
      themeFading = false;
    }, 2500);
    return () => {
      if (fadeTimer) clearTimeout(fadeTimer);
    };
  });

  $effect(() => {
    html; // fresh markup arrived -> restore full opacity
    if (fadeTimer) {
      clearTimeout(fadeTimer);
      fadeTimer = null;
    }
    themeFading = false;
  });

  /**
   * getScrollPercentage returns the current scroll position as a percentage (0-1).
   */
  export function getScrollPercentage(): number {
    if (!previewContainer) return 0;
    const { scrollTop, scrollHeight, clientHeight } = previewContainer;
    if (scrollHeight <= clientHeight) return 0;
    return scrollTop / (scrollHeight - clientHeight);
  }

  /**
   * setScrollPercentage sets the scroll position based on a percentage (0-1).
   */
  /**
   * getExportHTML returns the rendered document (Mermaid SVG, KaTeX, images)
   * as a self-contained <article>: local images become data: URIs, each
   * diagram keeps its Mermaid source in an inert <template>. Formulas keep
   * their TeX source in KaTeX's MathML annotation.
   */
  export async function getExportHTML(): Promise<string> {
    const article = previewContainer?.querySelector('article');
    if (!article) return '';
    const clone = article.cloneNode(true) as HTMLElement;
    clone.classList.remove('theme-fading');
    clone.removeAttribute('style');

    for (const div of Array.from(clone.querySelectorAll<HTMLElement>('.mermaid[data-source]'))) {
      const tpl = document.createElement('template');
      tpl.className = 'mermaid-source';
      tpl.content.append(div.dataset.source ?? '');
      div.removeAttribute('data-source');
      div.before(tpl);
    }

    for (const img of Array.from(clone.querySelectorAll('img'))) {
      const src = img.getAttribute('src') || '';
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

  export function setScrollPercentage(percentage: number) {
    if (!previewContainer) return;
    const { scrollHeight, clientHeight } = previewContainer;
    previewContainer.scrollTop = percentage * (scrollHeight - clientHeight);
  }

  /**
   * checkAndHandleResource validates if a path or URL is whitelisted.
   */
  async function checkAndHandleResource(target: string, type: 'path' | 'url'): Promise<boolean> {
    if (type === 'url') {
      const isAllowed = await backend.isURLAllowed(target);
      if (!isAllowed) {
        onsecurity_request?.({ type: 'url', resource: target });
        return false;
      }
    } else {
      const isAllowed = await backend.isPathAllowed(target);
      if (!isAllowed) {
        const parentDir = await backend.getParentDir(target);
        onsecurity_request?.({ type: 'path', resource: parentDir });
        return false;
      }
    }
    return true;
  }

  /**
   * renderContent performs sequential rendering of advanced Markdown features.
   */
  async function renderContent() {
    await tick();
    if (!previewContainer) return;

    // 1. Handle External Links & Markdown Internal Links
    const links = previewContainer.querySelectorAll('a');
    for (const link of Array.from(links)) {
      const href = link.getAttribute('href');
      if (!href) continue;

      const isExternal = href.startsWith('http://') || href.startsWith('https://');
      const isMarkdown = href.endsWith('.md') || href.endsWith('.markdown');

      if (isExternal) {
        link.classList.add('external-link');
        link.onclick = async (e) => {
          e.preventDefault(); e.stopPropagation();
          try {
            const domain = new URL(href).hostname;
            if (await checkAndHandleResource(domain, 'url')) {
              BrowserOpenURL(href);
            }
          } catch (err) { console.error("Invalid URL clicked:", href); }
        };
      } else if (isMarkdown && !href.startsWith('#')) {
        link.onclick = async (e) => {
          e.preventDefault(); e.stopPropagation();
          const baseDir = currentFilePath ? await backend.getParentDir(currentFilePath) : "";
          const absPath = await backend.resolveRelativePath(baseDir, href);
          if (await checkAndHandleResource(absPath, 'path')) {
            onopen_file?.({ path: absPath });
          }
        };
      } else if (href.startsWith('#')) {
        // Internal anchor links: let webview handle scroll to ID
      } else {
        link.onclick = (e) => {
          e.preventDefault();
          console.warn("Direct file links are blocked for security. Use Markdown files or explicit whitelist.");
        };
      }
    }

    // 2. Handle Images Security
    const images = previewContainer.querySelectorAll('img');
    for (const img of Array.from(images)) {
      const src = img.getAttribute('src');
      if (!src) continue;

      const isExternal = src.startsWith('http://') || src.startsWith('https://');
      if (isExternal) {
        try {
          const domain = new URL(src).hostname;
          const isAllowed = await backend.isURLAllowed(domain);
          if (!isAllowed) {
            img.style.display = 'none';
            onsecurity_request?.({ type: 'url', resource: domain });
          }
        } catch (e) { img.style.display = 'none'; }
      } else if (!src.startsWith('data:')) {
        const baseDir = currentFilePath ? await backend.getParentDir(currentFilePath) : "";
        const absPath = await backend.resolveRelativePath(baseDir, src);
        const isAllowed = await backend.isPathAllowed(absPath);
        if (!isAllowed) {
          img.style.display = 'none';
          const parentDir = await backend.getParentDir(absPath);
          onsecurity_request?.({ type: 'path', resource: parentDir });
        } else {
          // Served by the Go asset handler (localresource.go), which re-checks the whitelist
          img.src = "/local-resource?path=" + encodeURIComponent(absPath);
        }
      }
    }

    // 3. Render Mermaid Diagrams
    // Only freshly converted fences may run: re-running on already
    // rendered SVG (e.g. theme toggle before the debounced re-render
    // delivers fresh markup) yields red syntax-error boxes instead.
    const mermaidDivs = previewContainer.querySelectorAll('pre code.language-mermaid');
    let freshDiagrams = 0;
    mermaidDivs.forEach((el) => {
      const parent = el.parentElement;
      if (parent) {
        const content = el.textContent || "";
        const div = document.createElement('div');
        div.className = 'mermaid';
        div.dataset.source = content; // kept for the HTML export
        div.textContent = content;
        parent.replaceWith(div);
        freshDiagrams++;
      }
    });

    try {
      mermaid.initialize({
        startOnLoad: false,
        theme: theme.mermaidTheme,
        themeVariables: theme.mermaidVars || {},
        fontFamily: 'inherit',
      });

      const nodes = previewContainer.querySelectorAll('.mermaid');
      if (freshDiagrams > 0 && nodes.length > 0) {
        await mermaid.run({ querySelector: '.mermaid', suppressErrors: true });
      }
    } catch (err) {
      console.error("Mermaid render failed:", err);
    }

    // 4. Render Mathematical Expressions (KaTeX)
    try {
      renderMathInElement(previewContainer, {
        delimiters: [
          {left: '$$', right: '$$', display: true},
          {left: '$', right: '$', inline: true},
          {left: '\\(', right: '\\)', inline: true},
          {left: '\\[', right: '\\]', display: true}
        ],
        throwOnError: false
      });
    } catch (err) {
      console.error("KaTeX render failed:", err);
    }
  }

  // --- Svelte 5 Runes: Effect ---
  $effect(() => {
    if (html !== undefined || theme !== undefined) {
      untrack(() => renderContent());
    }
  });
</script>

<!-- Inject dynamic Chroma Syntax Highlighting CSS -->
{@html '<' + 'style' + '>' + css + '</' + 'style' + '>'}

<div
  bind:this={previewContainer}
  onscroll={(e) => {
    onscroll?.(e);
  }}
  class="preview-container flex-1 overflow-auto p-8 transition-colors duration-200 {theme.containerClass}"
>
  <article
    class="markdown-body md-doc"
    data-theme={theme.id}
    lang={$locale}
    class:theme-fading={themeFading}
    style="font-size: {fontSize}%;"
  >
    {@html html}
  </article>
</div>

<style>
  /* Content styles live in shared/markdown/document.css (shared with
     the browse view and the HTML export). Only preview chrome here. */
  :global(.markdown-body) { transition: opacity 0.22s ease; }
  :global(.markdown-body.theme-fading) { opacity: 0.3; }

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
    article.markdown-body {
      opacity: 1 !important;
      transition: none !important;
      padding: 0 !important;
      margin: 0 !important;
    }
  }
</style>
