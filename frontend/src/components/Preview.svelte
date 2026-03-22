<script lang="ts">
  /**
   * Preview component handles the complex rendering of Markdown-derived HTML.
   * It integrates Mermaid.js for diagrams and KaTeX for mathematical expressions.
   * Now with security whitelist enforcement for local and external resources.
   */
  import { onMount, tick, createEventDispatcher } from 'svelte';
  import mermaid from 'mermaid';
  import katex from 'katex';
  import 'katex/dist/katex.min.css';
  import renderMathInElement from 'katex/dist/contrib/auto-render';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js';
  import * as backend from '../lib/backend';
  import type { Theme } from '../themes';

  const dispatch = createEventDispatcher();

  // Component Props
  export let html: string;
  export let css: string = "";
  export let theme: Theme;
  export let fontSize: number = 100;
  export let currentFilePath: string | null = null;

  let previewContainer: HTMLElement;

  /**
   * checkAndHandleResource validates if a path or URL is whitelisted.
   * Returns true if allowed, otherwise triggers a security request and returns false.
   */
  async function checkAndHandleResource(target: string, type: 'path' | 'url'): Promise<boolean> {
    if (type === 'url') {
      const isAllowed = await backend.isURLAllowed(target);
      if (!isAllowed) {
        dispatch('security-request', { type: 'url', resource: target });
        return false;
      }
    } else {
      const isAllowed = await backend.isPathAllowed(target);
      if (!isAllowed) {
        const parentDir = await backend.getParentDir(target);
        dispatch('security-request', { type: 'path', resource: parentDir });
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
        // Prevent default browser opening. Only open via Wails after whitelist check.
        link.onclick = async (e) => {
          e.preventDefault();
          e.stopPropagation();
          try {
            const domain = new URL(href).hostname;
            if (await checkAndHandleResource(domain, 'url')) {
              BrowserOpenURL(href);
            }
          } catch (err) {
            console.error("Invalid URL clicked:", href);
          }
        };
      } else if (isMarkdown && !href.startsWith('#')) {
        link.onclick = async (e) => {
          e.preventDefault();
          e.stopPropagation();
          const baseDir = currentFilePath ? await backend.getParentDir(currentFilePath) : "";
          const absPath = await backend.resolveRelativePath(baseDir, href);
          if (await checkAndHandleResource(absPath, 'path')) {
            dispatch('open-file', { path: absPath });
          }
        };
      } else if (href.startsWith('#')) {
        // Internal anchor links: let the browser/webview handle scroll to ID
      } else {
        // Other local file links
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
              img.style.display = 'none'; // Hide until allowed
              dispatch('security-request', { type: 'url', resource: domain });
            }
        } catch (e) { img.style.display = 'none'; }
      } else if (!src.startsWith('data:')) {
        const baseDir = currentFilePath ? await backend.getParentDir(currentFilePath) : "";
        const absPath = await backend.resolveRelativePath(baseDir, src);
        const isAllowed = await backend.isPathAllowed(absPath);
        if (!isAllowed) {
          img.style.display = 'none';
          const parentDir = await backend.getParentDir(absPath);
          dispatch('security-request', { type: 'path', resource: parentDir });
        } else {
          img.src = "wails:///" + absPath.replace(/\\/g, '/');
        }
      }
    }

    // 3. Render Mermaid Diagrams
    const mermaidDivs = previewContainer.querySelectorAll('pre code.language-mermaid');
    mermaidDivs.forEach((el) => {
      const parent = el.parentElement;
      if (parent) {
        const content = el.textContent || "";
        const div = document.createElement('div');
        div.className = 'mermaid';
        div.textContent = content;
        parent.replaceWith(div);
      }
    });

    mermaid.initialize({
      startOnLoad: false,
      theme: theme.mermaidTheme,
      themeVariables: theme.mermaidVars || {},
      fontFamily: 'inherit',
    });

    try {
      const nodes = previewContainer.querySelectorAll('.mermaid');
      if (nodes.length > 0) {
          await mermaid.run({ 
            querySelector: '.mermaid',
            suppressErrors: true
          });
      }
    } catch (err) {
      console.error("Mermaid render failed:", err);
    }

    // 4. Render Mathematical Expressions (KaTeX)
    renderMathInElement(previewContainer, {
      delimiters: [
        {left: '$$', right: '$$', display: true},
        {left: '$', right: '$', inline: true},
        {left: '\\(', right: '\\)', inline: true},
        {left: '\\[', right: '\\]', display: true}
      ],
      throwOnError: false
    });
  }

  // Reactive updates
  $: if (html !== undefined || theme !== undefined) {
    renderContent();
  }

  onMount(() => {
    renderContent();
  });
</script>

<!-- Inject dynamic Chroma Syntax Highlighting CSS -->
{@html '<' + 'style' + '>' + css + '</' + 'style' + '>'}

<div
  bind:this={previewContainer}
  class="flex-1 overflow-auto p-8 transition-colors duration-300 {theme.containerClass}"
>
  <article
    class="prose lg:prose-xl max-w-none {theme.proseClass}"
    style="font-size: {fontSize}%;"
  >
    {@html html}
  </article>
</div>

<style>
  /* Base Markdown Styling from JSON Config */
  :global(.prose pre) {
    border-radius: 0.5rem;
    padding: 1rem;
    overflow-x: auto;
  }

  /* Improve Light Mode Code Readability */
  :global(.bg-white .prose pre) {
    background-color: #f8fafc !important;
    border: 1px solid #e2e8f0;
  }
  :global(.bg-white .prose pre code) { color: #1e293b !important; }
  :global(.bg-white .chroma .c, .bg-white .chroma .cm, .bg-white .chroma .c1) { color: #64748b !important; font-style: italic; }
  :global(.bg-white .chroma .m, .bg-white .chroma .mb, .bg-white .chroma .mf) { color: #0f172a !important; font-weight: 600; }
  :global(.bg-white .chroma .s, .bg-white .chroma .sa, .bg-white .chroma .sb) { color: #0f172a !important; }

  /* Ensure Sepia mode headings and text are always visible with high contrast */
  :global(.bg-\[\#f4ecd8\] article, .bg-\[\#f4ecd8\] h1, .bg-\[\#f4ecd8\] h2, .bg-\[\#f4ecd8\] h3, .bg-\[\#f4ecd8\] h4, .bg-\[\#f4ecd8\] h5, .bg-\[\#f4ecd8\] h6, .bg-\[\#f4ecd8\] p, .bg-\[\#f4ecd8\] li, .bg-\[\#f4ecd8\] strong) {
    color: #433422 !important;
  }

  /* External Link Indicator */
  :global(.external-link::after) {
    content: " ↗";
    font-size: 0.8em;
    opacity: 0.6;
  }

  /* GitHub-style Alerts (Admonitions) */
  :global(.markdown-alert) {
    padding: 0.75rem 1rem;
    margin-bottom: 1rem;
    color: inherit;
    border-left: 0.25rem solid;
    border-radius: 0 0.375rem 0.375rem 0;
    background: rgba(0, 0, 0, 0.03);
  }
  :global(.prose-invert .markdown-alert) { background: rgba(255, 255, 255, 0.05); }
  :global(.markdown-alert::before) { display: block; font-weight: 600; margin-bottom: 0.25rem; text-transform: capitalize; font-size: 0.875rem; }
  :global(.markdown-alert-note) { border-color: #0969da; }
  :global(.markdown-alert-note::before) { content: "ⓘ Note"; color: #0969da; }
  :global(.markdown-alert-tip) { border-color: #1a7f37; }
  :global(.markdown-alert-tip::before) { content: "💡 Tip"; color: #1a7f37; }
  :global(.markdown-alert-important) { border-color: #8250df; }
  :global(.markdown-alert-important::before) { content: "❗ Important"; color: #8250df; }
  :global(.markdown-alert-warning) { border-color: #9a6700; }
  :global(.markdown-alert-warning::before) { content: "⚠️ Warning"; color: #9a6700; }
  :global(.markdown-alert-caution) { border-color: #cf222e; }
  :global(.markdown-alert-caution::before) { content: "☢️ Caution"; color: #cf222e; }

  /* Mermaid / Diagrams - Base Styling */
  :global(.mermaid) {
    background: transparent;
    padding: 1rem;
    border-radius: 0.5rem;
    margin: 1.5rem 0;
    display: flex;
    justify-content: center;
    font-family: sans-serif;
  }
  :global(.mermaid .marker) { fill: currentColor !important; }
  :global(.mermaid .edgePath .path) { stroke: currentColor !important; }
  :global(.mermaid .edgeLabel), :global(.mermaid .edgeLabel span) { background-color: transparent !important; color: currentColor !important; }
  :global(.mermaid .edgeLabel rect) { opacity: 0.8; }
  :global(.mermaid svg[id^="mermaid-error"]) { border: 3px solid #ef4444 !important; border-radius: 0.5rem; padding: 1rem; background: rgba(239, 68, 68, 0.1) !important; }

  /* Dynamic Theme Overrides for Mermaid */
  :global(.bg-white .mermaid) { background: #f9fafb; }
  :global(.bg-slate-900 .mermaid) { background: #1e293b; }
  :global(.bg-\[\#f4ecd8\] .mermaid) { background: #e4dcc7; }
  :global(.monochrome .mermaid) { background: #ffffff; border: 1px solid #000; }
  :global(.bg-slate-900 .mermaid .edgeLabel rect) { fill: #1e293b !important; }
  :global(.bg-white .mermaid .edgeLabel rect) { fill: #f9fafb !important; }
  :global(.bg-\[\#f4ecd8\] .mermaid .edgeLabel rect) { fill: #f4ecd8 !important; }
  :global(.monochrome .mermaid .edgeLabel rect) { fill: #ffffff !important; }

  :global(.monochrome) { filter: grayscale(100%) contrast(110%); }

  @media print {
    div { overflow: visible !important; height: auto !important; padding: 0 !important; background: transparent !important; border: none !important; }
    article { font-size: 12pt !important; max-width: 100% !important; }
  }
</style>