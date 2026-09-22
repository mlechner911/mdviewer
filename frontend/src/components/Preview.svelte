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
  import { BrowserOpenURL } from '../lib/wails';
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
          // Served by the Go asset handler (localresource.go), which re-checks the whitelist
          img.src = "/local-resource?path=" + encodeURIComponent(absPath);
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
    class="markdown-body"
    style="font-size: {fontSize}%;"
  >
    {@html html}
  </article>
</div>

<style>
  /* Browse design system (ported from marksafe_browse): monochrome
     IDE look. Tokens keyed on the theme container classes, same pattern
     as the rules below. */
  :global(.bg-white .markdown-body) {
    --font-sans: 'Inter', system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    --font-mono: 'JetBrains Mono', 'Fira Code', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    --radius-sm: 3px;
    --radius-md: 4px;
    --lh-body: 1.42;
    --lh-code: 1.3;
    --bg: #ffffff;
    --surface: #f8fafc;
    --surface-hover: rgba(99, 102, 241, 0.08);
    --text: #0f172a;
    --heading: #334155;
    --muted: #64748b;
    --text-muted: #64748b;
    --border: #e2e8f0;
    --accent: #4f46e5;
    --accent-strong: #4338ca;
    --accent-soft: rgba(99, 102, 241, 0.1);
    --code-bg: #f1f5f9;
    --code-border: #e2e8f0;
    --quote-bg: rgba(99, 102, 241, 0.06);
    --table-head-bg: #f8fafc;
  }
  :global(.bg-slate-900 .markdown-body) {
    --font-sans: 'Inter', system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    --font-mono: 'JetBrains Mono', 'Fira Code', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    --radius-sm: 3px;
    --radius-md: 4px;
    --lh-body: 1.42;
    --lh-code: 1.3;
    --bg: #0d0e11;
    --surface: #14161a;
    --surface-hover: rgba(129, 140, 248, 0.12);
    --text: #adbac7;
    --heading: #f3f4f6;
    --muted: #94a3b8;
    --text-muted: #94a3b8;
    --border: #21262d;
    --accent: #818cf8;
    --accent-strong: #a5b4fc;
    --accent-soft: rgba(129, 140, 248, 0.12);
    --code-bg: #151b23;
    --code-border: #21262d;
    --quote-bg: rgba(129, 140, 248, 0.08);
    --table-head-bg: #151b23;
  }
  :global(.markdown-body) {
    font-family: var(--font-sans);
    font-size: 1rem;
    line-height: var(--lh-body);
    max-width: 960px;
    overflow-wrap: break-word;
    color: var(--text);
  }
  :global(.markdown-body h1), :global(.markdown-body h2), :global(.markdown-body h3) {
    font-family: var(--font-mono);
    font-weight: 700;
    line-height: 1.3;
    letter-spacing: -0.01em;
    color: var(--heading);
  }
  :global(.markdown-body h1) {
    font-size: 1.75rem;
    margin-bottom: 1rem;
    border-bottom: 2px solid var(--border);
    padding-bottom: 0.5rem;
  }
  :global(.markdown-body h1:first-child) { margin-top: 0; }
  :global(.markdown-body h2) { font-size: 1.35rem; margin: 1.2rem 0 0.4rem; }
  :global(.markdown-body h3) { font-size: 1.1rem; margin: 0.9rem 0 0.35rem; }
  :global(.markdown-body h4), :global(.markdown-body h5), :global(.markdown-body h6) {
    line-height: 1.4;
    margin-top: 1em;
    margin-bottom: 0.3em;
    font-weight: 600;
  }
  :global(.markdown-body p) { margin: 0.7rem 0; line-height: var(--lh-body); }
  :global(.markdown-body ul), :global(.markdown-body ol) { margin: 0.7rem 0; padding-left: 1.75rem; }
  :global(.markdown-body li) { margin: 0.2rem 0; line-height: var(--lh-body); }
  :global(.markdown-body li > p) { margin-top: 0.25em; margin-bottom: 0.25em; }
  :global(.markdown-body a) { color: var(--accent); text-decoration: none; }
  :global(.markdown-body a:hover) { text-decoration: underline; }
  :global(.markdown-body img) { max-width: 100%; border-radius: var(--radius-md); }
  :global(.markdown-body hr) { border: none; border-top: 1px solid var(--border); margin: 2rem 0; }
  :global(.markdown-body code) {
    font-family: var(--font-mono);
    font-size: 0.85em;
    background: var(--code-bg);
    border: 1px solid var(--code-border);
    padding: 1px 6px;
    border-radius: var(--radius-sm);
  }
  :global(.markdown-body pre) {
    position: relative;
    background: var(--code-bg);
    border: 1px solid var(--code-border);
    border-radius: var(--radius-md);
    padding: 1rem 1.25rem;
    overflow-x: auto;
    margin: 1.25rem 0;
  }
  :global(.markdown-body pre code) {
    background: none;
    border: none;
    padding: 0;
    font-size: 0.875rem;
    line-height: var(--lh-code);
  }
  :global(.markdown-body table) { border-collapse: collapse; width: 100%; margin: 1.25rem 0; font-size: 0.925rem; }
  :global(.markdown-body th), :global(.markdown-body td) {
    border: 1px solid var(--border);
    padding: 6px 10px;
    text-align: left;
    line-height: 1.3;
  }
  :global(.markdown-body th) {
    background: var(--table-head-bg);
    font-family: var(--font-mono);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  :global(.markdown-body tr:nth-child(even) td) { background: var(--accent-soft); }
  :global(.markdown-body blockquote) {
    background: var(--quote-bg);
    border-left: 3px solid var(--accent);
    border-radius: 0 var(--radius-md) var(--radius-md) 0;
    margin: 1.25rem 0;
    padding: 0.75rem 1.25rem;
  }
  :global(.markdown-body blockquote p) { margin: 0.5rem 0; }

  /* External Link Indicator */
  :global(.markdown-body a.external-link::after) {
    content: "";
    display: inline-block;
    width: 0.72em;
    height: 0.72em;
    margin-left: 0.28em;
    vertical-align: -0.08em;
    background-color: currentColor;
    opacity: 0.75;
    -webkit-mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='black' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6'/%3E%3Cpolyline points='15 3 21 3 21 9'/%3E%3Cline x1='10' y1='14' x2='21' y2='3'/%3E%3C/svg%3E") no-repeat center / contain;
    mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='black' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6'/%3E%3Cpolyline points='15 3 21 3 21 9'/%3E%3Cline x1='10' y1='14' x2='21' y2='3'/%3E%3C/svg%3E") no-repeat center / contain;
  }
  :global(.markdown-body a.external-link:hover::after) { opacity: 1; }

  /* Task Lists (Checkboxes) */
  :global(.markdown-body ul > li:has(input[type="checkbox"])) {
    list-style-type: none;
    padding-left: 0;
  }
  :global(.markdown-body ul > li > input[type="checkbox"]) {
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
  :global(.bg-slate-900 .markdown-alert) { background: rgba(255, 255, 255, 0.05); }
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

    article.markdown-body {
      font-size: 11pt !important;
      line-height: 1.55 !important;
      max-width: 100% !important;
      color: #111827 !important;
      padding: 0 !important;
      margin: 0 !important;
    }

    /* Headings */
    :global(.markdown-body h1), :global(.markdown-body h2), :global(.markdown-body h3),
    :global(.markdown-body h4), :global(.markdown-body h5), :global(.markdown-body h6) {
      color: #000000 !important;
      page-break-after: avoid !important;
      break-after: avoid !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
    }

    :global(.markdown-body h1) {
      font-size: 18pt !important;
      margin-top: 0 !important;
      margin-bottom: 0.8rem !important;
      border-bottom: 1.5pt solid #e5e7eb !important;
      padding-bottom: 0.3rem !important;
    }

    :global(.markdown-body h2) {
      font-size: 14pt !important;
      margin-top: 1.2rem !important;
      margin-bottom: 0.5rem !important;
      border-bottom: 1pt solid #f3f4f6 !important;
      padding-bottom: 0.2rem !important;
    }

    :global(.markdown-body h3) {
      font-size: 12pt !important;
      margin-top: 1rem !important;
      margin-bottom: 0.4rem !important;
    }

    :global(.markdown-body p), :global(.markdown-body li) {
      color: #1f2937 !important;
      orphans: 3 !important;
      widows: 3 !important;
    }

    :global(.markdown-body li) {
      page-break-inside: avoid !important;
      break-inside: avoid !important;
    }

    /* Code Blocks & Inlines in Print */
    :global(.markdown-body pre) {
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

    :global(.markdown-body pre code) {
      background: transparent !important;
      color: #0f172a !important;
      font-size: 9.5pt !important;
      line-height: 1.4 !important;
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace !important;
    }

    :global(.markdown-body :not(pre) > code) {
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
    :global(.markdown-body table) {
      width: 100% !important;
      border-collapse: collapse !important;
      margin: 12pt 0 !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      font-size: 10pt !important;
    }

    :global(.markdown-body th), :global(.markdown-body td) {
      border: 1px solid #cbd5e1 !important;
      padding: 5pt 8pt !important;
      color: #111827 !important;
    }

    :global(.markdown-body th) {
      background-color: #f1f5f9 !important;
      font-weight: 600 !important;
    }

    :global(.markdown-body tr:nth-child(even) td) {
      background-color: #f8fafc !important;
    }

    :global(.markdown-body tr) {
      page-break-inside: avoid !important;
      break-inside: avoid !important;
    }

    /* Blockquotes & Alerts */
    :global(.markdown-body blockquote) {
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

    :global(.markdown-body img) {
      max-width: 100% !important;
      height: auto !important;
      page-break-inside: avoid !important;
      break-inside: avoid !important;
      margin: 10pt auto !important;
    }

    :global(.markdown-body a) {
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