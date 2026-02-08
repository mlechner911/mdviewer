<script lang="ts">
  /**
   * Preview component handles the complex rendering of Markdown-derived HTML.
   * It integrates Mermaid.js for diagrams and KaTeX for mathematical expressions.
   */
  import { onMount, tick } from 'svelte';
  import mermaid from 'mermaid';
  import katex from 'katex';
  import 'katex/dist/katex.min.css';
  import renderMathInElement from 'katex/dist/contrib/auto-render';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js';
  import type { Theme } from '../themes';

  // Component Props
  export let html: string;
  export let css: string = "";
  export let theme: Theme;
  export let fontSize: number = 100;

  let previewContainer: HTMLElement;

  /**
   * renderContent performs sequential rendering of advanced Markdown features.
   */
  async function renderContent() {
    await tick();
    if (!previewContainer) return;

    // 1. Handle External Links
    const links = previewContainer.querySelectorAll('a');
    links.forEach(link => {
      const href = link.getAttribute('href');
      if (href && (href.startsWith('http://') || href.startsWith('https://'))) {
        link.classList.add('external-link');
        link.target = "_blank";
        link.onclick = (e) => {
          e.preventDefault();
          BrowserOpenURL(href);
        };
      }
    });

    // 2. Render Mermaid Diagrams
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
  //    parseError: () => {}
    });

    try {
      const nodes = previewContainer.querySelectorAll('.mermaid');
      if (nodes.length > 0) {
        await mermaid.run({ nodes });
      }
    } catch (err) {
      console.error("Mermaid render failed:", err);
    }

    // 3. Render Mathematical Expressions (KaTeX)
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
  class="flex-1 overflow-y-auto p-8 transition-colors duration-300 {theme.containerClass}"
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

  :global(.prose-invert .markdown-alert) {
    background: rgba(255, 255, 255, 0.05);
  }

  :global(.markdown-alert::before) {
    display: block;
    font-weight: 600;
    margin-bottom: 0.25rem;
    text-transform: capitalize;
    font-size: 0.875rem;
  }

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
  }
  :global(.mermaid .marker) { fill: currentColor !important; }
  :global(.mermaid .edgePath .path) { stroke: currentColor !important; }

  /* Improve Mermaid Edge Label readability */
  :global(.mermaid .edgeLabel), :global(.mermaid .edgeLabel span) {
    background-color: transparent !important;
    color: currentColor !important;
  }

  :global(.mermaid .edgeLabel rect) {
    opacity: 0.8;
  }

  :global(.mermaid svg[id^="mermaid-error"]) {
    border: 3px solid #ef4444 !important;
    border-radius: 0.5rem;
    padding: 1rem;
    background: rgba(239, 68, 68, 0.1) !important;
  }

  /* Dynamic Theme Overrides for Mermaid */
  :global(.bg-white .mermaid) { background: #f9fafb; }
  :global(.bg-slate-900 .mermaid) { background: #1e293b; }
  :global(.bg-\[\#f4ecd8\] .mermaid) { background: #e4dcc7; }
  :global(.monochrome .mermaid) {
    background: #ffffff;
    border: 1px solid #000;
  }

  :global(.bg-slate-900 .mermaid .edgeLabel rect) { fill: #1e293b !important; }
  :global(.bg-white .mermaid .edgeLabel rect) { fill: #f9fafb !important; }
  :global(.bg-\[\#f4ecd8\] .mermaid .edgeLabel rect) { fill: #f4ecd8 !important; }
  :global(.monochrome .mermaid .edgeLabel rect) { fill: #ffffff !important; }
</style>