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
  import renderMathInElement from 'katex/dist/contrib/auto-render';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js';
  import * as backend from '../lib/backend';
  import type { Theme } from '../themes';

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

    try {
      mermaid.initialize({
        startOnLoad: false,
        theme: theme.mermaidTheme,
        themeVariables: theme.mermaidVars || {},
        fontFamily: 'inherit',
      });

      const nodes = previewContainer.querySelectorAll('.mermaid');
      if (nodes.length > 0) {
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
    class="prose max-w-none {theme.proseClass}"
    style="font-size: {fontSize}%;"
  >
    {@html html}
  </article>
</div>

<style>
  /* Compact & Readable Typography for Markdown Preview */
  :global(.prose) {
    line-height: 1.6;
  }
  :global(.prose p) {
    margin-top: 0.75em;
    margin-bottom: 0.75em;
    line-height: 1.6;
  }
  :global(.prose h1) {
    font-size: 1.875rem;
    line-height: 1.25;
    margin-top: 1.4em;
    margin-bottom: 0.5em;
    font-weight: 700;
  }
  :global(.prose h1:first-child) {
    margin-top: 0;
  }
  :global(.prose h2) {
    font-size: 1.5rem;
    line-height: 1.3;
    margin-top: 1.3em;
    margin-bottom: 0.45em;
    font-weight: 600;
  }
  :global(.prose h3) {
    font-size: 1.25rem;
    line-height: 1.35;
    margin-top: 1.1em;
    margin-bottom: 0.4em;
    font-weight: 600;
  }
  :global(.prose h4, .prose h5, .prose h6) {
    line-height: 1.4;
    margin-top: 1em;
    margin-bottom: 0.3em;
    font-weight: 600;
  }
  :global(.prose ul, .prose ol) {
    margin-top: 0.5em;
    margin-bottom: 0.5em;
    padding-left: 1.5em;
  }
  :global(.prose li) {
    margin-top: 0.25em;
    margin-bottom: 0.25em;
    line-height: 1.6;
  }
  :global(.prose li > p) {
    margin-top: 0.25em;
    margin-bottom: 0.25em;
  }
  :global(.prose blockquote) {
    margin-top: 0.85em;
    margin-bottom: 0.85em;
  }
  :global(.prose hr) {
    margin-top: 1.5em;
    margin-bottom: 1.5em;
  }
  :global(.prose table) {
    margin-top: 0.85em;
    margin-bottom: 0.85em;
  }

  /* Base Markdown Styling */
  :global(.prose pre) {
    border-radius: 0.5rem;
    padding: 1rem;
    overflow-x: auto;
    margin-top: 0.85em;
    margin-bottom: 0.85em;
    line-height: 1.45;
  }

  /* Light Mode Code Styling */
  :global(.bg-white .prose pre) {
    background-color: #f8fafc !important;
    border: 1px solid #e2e8f0;
  }
  :global(.bg-white .prose pre code) { color: #1e293b; }

  /* Dark Mode Code Styling */
  :global(.bg-slate-900 .prose pre) {
    background-color: #0f172a !important;
    border: 1px solid #334155;
  }
  :global(.bg-slate-900 .prose pre code) { color: #f1f5f9; }

  /* External Link Indicator */
  :global(.external-link::after) {
    content: " ↗";
    font-size: 0.8em;
    opacity: 0.6;
  }

  /* Task Lists (Checkboxes) */
  :global(.prose ul > li:has(input[type="checkbox"])) {
    list-style-type: none;
    padding-left: 0;
  }
  :global(.prose ul > li > input[type="checkbox"]) {
    margin-right: 0.5rem;
    margin-bottom: 0.125rem;
    vertical-align: middle;
    pointer-events: none;
    width: 1.1rem;
    height: 1.1rem;
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

  /* Theme Overrides for Mermaid */
    :global(.bg-white .mermaid) { background: transparent !important; border: 1px solid #e2e8f0; }
    :global(.bg-slate-900 .mermaid) { background: transparent !important; border: 1px solid #334155; }
    :global(.bg-white .mermaid .edgeLabel rect) { fill: transparent !important; }
    :global(.bg-slate-900 .mermaid .edgeLabel rect) { fill: transparent !important; }

  /* Front Matter Metadata Styling */
  :global(.frontmatter-container) {
    border: 1px solid #e2e8f0;
    border-radius: 0.5rem;
    background-color: #f8fafc;
    margin-bottom: 1.5rem;
  }
  :global(.frontmatter-container summary) {
    border-bottom: 1px solid #e2e8f0;
    background-color: #f1f5f9;
    padding: 0.5rem 1rem;
    cursor: pointer;
  }
  :global(.bg-slate-900 .frontmatter-container) {
    border-color: #334155;
    background-color: #0f172a;
  }
  :global(.bg-slate-900 .frontmatter-container summary) {
    border-color: #334155;
    background-color: #1e293b;
    color: #f1f5f9;
  }
  :global(.frontmatter-tag) {
    display: inline-block;
    padding: 0.1rem 0.45rem;
    margin: 0.1rem 0.2rem 0.1rem 0;
    font-size: 0.75rem;
    font-family: ui-monospace, monospace;
    border-radius: 9999px;
    background-color: #e2e8f0;
    color: #1e293b;
  }
  :global(.bg-slate-900 .frontmatter-tag) {
    background-color: #334155;
    color: #f1f5f9;
  }

  /* =========================================================
   * PRINT STYLESHEET: Pristine, High-Contrast Document Output
   * ========================================================= */
  @media print {
    :global(html), :global(body), :global(#app), :global(main), .preview-container {
      background: #ffffff !important;
      background-color: #ffffff !important;
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

    .preview-container {
      padding: 0 !important;
    }

    article.prose {
      font-size: 11pt !important;
      line-height: 1.55 !important;
      max-width: 100% !important;
      color: #111827 !important;
      padding: 0 !important;
      margin: 0 !important;
    }

    /* Headings */
    :global(.prose h1), :global(.prose h2), :global(.prose h3),
    :global(.prose h4), :global(.prose h5), :global(.prose h6) {
      color: #000000 !important;
      page-break-after: avoid !important;
      break-after: avoid !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
    }

    :global(.prose h1) {
      font-size: 18pt !important;
      margin-top: 0 !important;
      margin-bottom: 0.8rem !important;
      border-bottom: 1.5pt solid #e5e7eb !important;
      padding-bottom: 0.3rem !important;
    }

    :global(.prose h2) {
      font-size: 14pt !important;
      margin-top: 1.2rem !important;
      margin-bottom: 0.5rem !important;
      border-bottom: 1pt solid #f3f4f6 !important;
      padding-bottom: 0.2rem !important;
    }

    :global(.prose h3) {
      font-size: 12pt !important;
      margin-top: 1rem !important;
      margin-bottom: 0.4rem !important;
    }

    :global(.prose p), :global(.prose li) {
      color: #1f2937 !important;
      orphans: 3 !important;
      widows: 3 !important;
    }

    :global(.prose li) {
      page-break-inside: avoid !important;
      break-inside: avoid !important;
    }

    /* Code Blocks & Inlines in Print */
    :global(.prose pre) {
      background-color: #f8fafc !important;
      border: 1px solid #cbd5e1 !important;
      border-radius: 4px !important;
      padding: 8pt 10pt !important;
      margin: 10pt 0 !important;
      white-space: pre-wrap !important;
      word-break: break-word !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      box-shadow: none !important;
    }

    :global(.prose pre code) {
      background: transparent !important;
      color: #0f172a !important;
      font-size: 9.5pt !important;
      line-height: 1.4 !important;
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace !important;
    }

    :global(.prose :not(pre) > code) {
      background-color: #f1f5f9 !important;
      color: #0f172a !important;
      border: 1px solid #e2e8f0 !important;
      border-radius: 3px !important;
      padding: 1px 4px !important;
      font-size: 9.5pt !important;
    }

    /* Chroma Syntax Highlighting for Print */
    :global(.chroma) { color: #0f172a !important; background-color: transparent !important; }
    :global(.chroma .k), :global(.chroma .kd), :global(.chroma .kn), :global(.chroma .kp), :global(.chroma .kr) { color: #cf222e !important; font-weight: bold !important; }
    :global(.chroma .s), :global(.chroma .sa), :global(.chroma .sb), :global(.chroma .sc), :global(.chroma .s1), :global(.chroma .s2) { color: #0a3069 !important; }
    :global(.chroma .c), :global(.chroma .ch), :global(.chroma .cm), :global(.chroma .c1), :global(.chroma .cs) { color: #57606a !important; font-style: italic !important; }
    :global(.chroma .nf), :global(.chroma .na), :global(.chroma .nb) { color: #8250df !important; }
    :global(.chroma .nc), :global(.chroma .nn) { color: #953800 !important; font-weight: bold !important; }
    :global(.chroma .m), :global(.chroma .mb), :global(.chroma .mf), :global(.chroma .mh), :global(.chroma .mi), :global(.chroma .mo) { color: #0550ae !important; }
    :global(.chroma .o), :global(.chroma .ow) { color: #cf222e !important; }

    /* Tables */
    :global(.prose table) {
      width: 100% !important;
      border-collapse: collapse !important;
      margin: 12pt 0 !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      font-size: 10pt !important;
    }

    :global(.prose th), :global(.prose td) {
      border: 1px solid #cbd5e1 !important;
      padding: 5pt 8pt !important;
      color: #111827 !important;
    }

    :global(.prose th) {
      background-color: #f1f5f9 !important;
      font-weight: 600 !important;
    }

    :global(.prose tr:nth-child(even) td) {
      background-color: #f8fafc !important;
    }

    :global(.prose tr) {
      page-break-inside: avoid !important;
      break-inside: avoid !important;
    }

    /* Blockquotes & Alerts */
    :global(.prose blockquote) {
      border-left: 3pt solid #94a3b8 !important;
      background-color: #f8fafc !important;
      color: #334155 !important;
      padding: 6pt 10pt !important;
      margin: 10pt 0 !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      border-radius: 0 4px 4px 0 !important;
    }

    :global(.markdown-alert) {
      padding: 8pt 10pt !important;
      margin: 10pt 0 !important;
      border-left: 3.5pt solid !important;
      border-radius: 0 4px 4px 0 !important;
      background-color: #f8fafc !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      color: #1f2937 !important;
    }

    :global(.markdown-alert-note) { border-color: #0969da !important; background-color: #f0f6fc !important; }
    :global(.markdown-alert-tip) { border-color: #1a7f37 !important; background-color: #f0fdf4 !important; }
    :global(.markdown-alert-important) { border-color: #8250df !important; background-color: #faf5ff !important; }
    :global(.markdown-alert-warning) { border-color: #9a6700 !important; background-color: #fffbeb !important; }
    :global(.markdown-alert-caution) { border-color: #cf222e !important; background-color: #fef2f2 !important; }

    /* Mermaid Diagrams in Print */
    :global(.mermaid) {
      background-color: #ffffff !important;
      border: 1px solid #e2e8f0 !important;
      border-radius: 4px !important;
      padding: 10pt !important;
      margin: 12pt auto !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      max-width: 100% !important;
    }

    :global(.mermaid svg) {
      max-width: 100% !important;
      height: auto !important;
    }

    /* KaTeX & Media in Print */
    :global(.katex-display) {
      margin: 10pt 0 !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
    }

    :global(.katex) {
      color: #000000 !important;
    }

    :global(.prose img) {
      max-width: 100% !important;
      height: auto !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      margin: 10pt auto !important;
    }

    :global(.prose a) {
      color: #0969da !important;
      text-decoration: underline !important;
    }

    :global(.external-link::after) {
      content: "" !important;
    }

    /* Front Matter Metadata in Print */
    :global(.frontmatter-container) {
      border: 1px solid #cbd5e1 !important;
      background-color: #f8fafc !important;
      margin-bottom: 12pt !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      border-radius: 4px !important;
    }
    :global(.frontmatter-container summary) {
      border-bottom: 1px solid #cbd5e1 !important;
      background-color: #f1f5f9 !important;
      color: #111827 !important;
      padding: 4pt 8pt !important;
    }
    :global(.frontmatter-tag) {
      border: 1px solid #cbd5e1 !important;
      background-color: #ffffff !important;
      color: #0f172a !important;
    }
  }
</style>