<script lang="ts">
  import { onMount, tick } from 'svelte';
  import mermaid from 'mermaid';
  import katex from 'katex';
  import 'katex/dist/katex.min.css';
  import renderMathInElement from 'katex/dist/contrib/auto-render';

  export let html: string;
  export let theme: 'light' | 'dark' | 'sepia' | 'monochrome' = 'dark';
  export let fontSize: number = 100;

  let previewContainer: HTMLElement;

  async function renderContent() {
    await tick();
    if (!previewContainer) return;

    // 1. Render Mermaid
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
      theme: (theme === 'light' || theme === 'monochrome') ? 'default' : theme === 'dark' ? 'dark' : 'neutral',
      themeVariables: theme === 'sepia' ? {
        primaryColor: '#d4c5a9',
        primaryTextColor: '#5b4636',
        primaryBorderColor: '#a89d85',
        lineColor: '#5b4636',
        secondaryColor: '#e4dcc7',
        tertiaryColor: '#f4ecd8'
      } : (theme === 'monochrome' ? {
        primaryColor: '#ffffff',
        primaryTextColor: '#000000',
        primaryBorderColor: '#000000',
        lineColor: '#000000',
        secondaryColor: '#eeeeee',
        tertiaryColor: '#ffffff'
      } : {})
    });

    try {
      if (previewContainer.querySelectorAll('.mermaid').length > 0) {
        await mermaid.run({
          nodes: previewContainer.querySelectorAll('.mermaid'),
        });
      }
    } catch (err) {
      console.error("Mermaid render failed:", err);
    }

    // 2. Render Math (KaTeX)
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

  // Re-run whenever html or theme changes
  $: if (html !== undefined || theme !== undefined) {
    renderContent();
  }

  onMount(() => {
    renderContent();
  });
</script>

<div 
  bind:this={previewContainer}
  class="flex-1 overflow-y-auto p-8 transition-colors duration-300 {theme === 'light' ? 'bg-white text-slate-900' : theme === 'dark' ? 'bg-slate-900 text-slate-100' : theme === 'sepia' ? 'bg-[#f4ecd8] text-[#5b4636]' : 'bg-white text-black monochrome'}"
>
  <article 
    class="prose lg:prose-xl max-w-none {theme === 'dark' ? 'prose-invert' : ''} {theme === 'sepia' ? 'prose-stone' : ''} {theme === 'monochrome' ? 'prose-slate' : ''}"
    style="font-size: {fontSize}%;"
  >
    {@html html}
  </article>
</div>

<style>
  :global(.prose pre) {
    background-color: #1e1e1e !important;
    border-radius: 0.5rem;
  }
  
  /* Reset code blocks for monochrome theme */
  :global(.monochrome .prose pre) {
    background-color: #f3f4f6 !important;
    border: 1px solid #000;
  }
  :global(.monochrome .prose code) {
    color: #000 !important;
  }

  :global(.mermaid) {
    background: transparent;
    padding: 1rem;
    border-radius: 0.5rem;
    margin: 1.5rem 0;
    display: flex;
    justify-content: center;
  }
  
  :global(.bg-white .mermaid) {
    background: #f9fafb;
  }
  
  :global(.bg-slate-900 .mermaid) {
    background: #1e293b;
  }

  :global(.bg-\[\#f4ecd8\] .mermaid) {
    background: #e4dcc7;
  }

  :global(.monochrome .mermaid) {
    background: #ffffff;
    border: 1px solid #000;
  }
</style>
