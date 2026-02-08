<script lang="ts">
  import { onMount } from 'svelte';
  import { RenderMarkdown, OpenFile, SaveFile } from '../wailsjs/go/main/App.js'

  let markdown: string = "# Welcome to MD Viewer\n\nStart typing here...";
  let htmlContent: string = "";

  async function updatePreview() {
    try {
      htmlContent = await RenderMarkdown(markdown);
    } catch (err) {
      console.error("Failed to render markdown:", err);
    }
  }

  async function handleOpen() {
    try {
      const content = await OpenFile();
      if (content !== undefined && content !== null) {
        markdown = content;
      }
    } catch (err) {
      console.error("Failed to open file:", err);
    }
  }

  async function handleSave() {
    try {
      await SaveFile(markdown);
    } catch (err) {
      console.error("Failed to save file:", err);
    }
  }

  // Initial render
  onMount(() => {
    updatePreview();
  });

  $: if (markdown !== undefined) {
    updatePreview();
  }
</script>

<main class="flex h-screen w-full overflow-hidden bg-slate-900 text-slate-100 flex-col">
  <!-- Toolbar -->
  <div class="h-12 bg-slate-800 border-b border-slate-700 flex items-center px-4 gap-4">
    <button 
      on:click={handleOpen}
      class="px-3 py-1 bg-slate-700 hover:bg-slate-600 rounded text-sm transition-colors"
    >
      Open
    </button>
    <button 
      on:click={handleSave}
      class="px-3 py-1 bg-slate-700 hover:bg-slate-600 rounded text-sm transition-colors"
    >
      Save
    </button>
    <div class="flex-1"></div>
    <span class="text-xs text-slate-500 font-mono">MD Viewer v0.1</span>
  </div>

  <div class="flex flex-1 overflow-hidden">
    <!-- Editor Column -->
    <div class="flex-1 flex flex-col border-r border-slate-700">
      <div class="p-2 bg-slate-800 text-xs font-bold uppercase tracking-wider text-slate-400 border-b border-slate-700">
        Editor
      </div>
      <textarea
        bind:value={markdown}
        class="flex-1 p-4 bg-slate-900 text-slate-100 focus:outline-none resize-none font-mono text-sm"
        placeholder="Type markdown here..."
      ></textarea>
    </div>

    <!-- Preview Column -->
    <div class="flex-1 flex flex-col">
      <div class="p-2 bg-slate-800 text-xs font-bold uppercase tracking-wider text-slate-400 border-b border-slate-700">
        Preview
      </div>
      <div class="flex-1 overflow-y-auto p-8 bg-white text-slate-900">
        <article class="prose lg:prose-xl max-w-none">
          {@html htmlContent}
        </article>
      </div>
    </div>
  </div>
</main>

<style>
  /* Base styles for the prose class are handled by @tailwindcss/typography */
  :global(.prose pre) {
    background-color: #1e1e1e !important;
  }
</style>
