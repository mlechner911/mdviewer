<script lang="ts">
  import { onMount } from 'svelte';
  import { RenderMarkdown, OpenFile, SaveFile, GetStyleCSS } from '../wailsjs/go/main/App.js'
  import Preview from './components/Preview.svelte';
  import { themes } from './themes';

  let markdown: string = `# Welcome to MD Viewer

## Mermaid Diagrams
\`\`\`mermaid
graph TD
    A[Start] --> B{Is it working?}
    B -- Yes --> C[Great!]
    B -- No --> D[Check Logs]
\`\`\`

## Feature Verification Table

| Feature | Support | Engine | Notes |
| :--- | :---: | :--- | :--- |
| **GFM** | ✅ | Goldmark | Tables, Tasklists |
| **Theming** | ✅ | Svelte | Dark, Light, Sepia |
| **Mermaid** | ✅ | Mermaid.js | Dynamic Rendering |
| **Highlight** | ✅ | Chroma | Inline CSS |

## Language Gallery

### Java
\`\`\`java
public class HelloWorld {
    public static void main(String[] args) {
        System.out.println("Hello, Java!");
    }
}
\`\`\`
`;
  let htmlContent: string = "";
  let highlightingCSS: string = "";
  let currentPreviewTheme = themes[0];
  let fontSize: number = 90;

  // App Frame Theming
  type AppTheme = 'dark' | 'light' | 'auto';
  let appTheme: AppTheme = 'dark';
  let effectiveAppTheme: 'dark' | 'light' = 'dark';

  function updateEffectiveTheme() {
    if (appTheme === 'auto') {
      effectiveAppTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    } else {
      effectiveAppTheme = appTheme;
    }
  }

  function toggleAppTheme() {
    if (appTheme === 'dark') appTheme = 'light';
    else if (appTheme === 'light') appTheme = 'auto';
    else appTheme = 'dark';
  }

  // Resizable logic
  let splitWidth: number = 50;
  let isResizing = false;

  function startResizing() { isResizing = true; }
  function stopResizing() { isResizing = false; }
  function onMouseMove(event: MouseEvent) {
    if (!isResizing) return;
    splitWidth = (event.clientX / window.innerWidth) * 100;
    if (splitWidth < 10) splitWidth = 10;
    if (splitWidth > 90) splitWidth = 90;
  }

  // Debounce logic
  let timeout: ReturnType<typeof setTimeout>;

  function debouncedUpdate(value: string, themeStyle: string) {
    clearTimeout(timeout);
    timeout = setTimeout(async () => {
      try {
        if (typeof window !== 'undefined' && (window as any).go && (window as any).go.main && (window as any).go.main.App) {
          htmlContent = await RenderMarkdown(value, themeStyle);
        } else if (typeof (window as any).marked === 'function') {
          // Use client-side marked renderer if already available in the page
          htmlContent = (window as any).marked(value);
        } else {
          htmlContent = '<pre style="white-space:pre-wrap;">Preview unavailable (no backend runtime)</pre>';
        }
      } catch (err) {
        console.error("Failed to render markdown:", err);
      }
    }, 300);
  }

  async function updateHighlightingCSS(style: string) {
    try {
      if (typeof window !== 'undefined' && (window as any).go && (window as any).go.main && (window as any).go.main.App) {
        highlightingCSS = await GetStyleCSS(style);
      } else {
        // Running in Vite/dev-only mode (no Wails runtime). Use empty CSS to avoid runtime errors.
        highlightingCSS = "";
        console.warn('Wails runtime not available; skipping GetStyleCSS call');
      }
    } catch (err) {
      console.error("Failed to get highlighting CSS:", err);
    }
  }

  async function handleOpen() {
    try {
      const content = (typeof window !== 'undefined' && (window as any).go && (window as any).go.main && (window as any).go.main.App)
        ? await OpenFile()
        : undefined;
      if (content !== undefined && content !== null) { markdown = content; }
    } catch (err) { console.error("Failed to open file:", err); }
  }

  async function handleSave() {
    try { await SaveFile(markdown); } catch (err) { console.error("Failed to save file:", err); }
  }

  function adjustFontSize(delta: number) {
    fontSize = Math.min(Math.max(fontSize + delta, 50), 200);
  }

  onMount(() => {
    updateEffectiveTheme();
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => { if (appTheme === 'auto') updateEffectiveTheme(); };
    mediaQuery.addEventListener('change', handler);
    updateHighlightingCSS(currentPreviewTheme.chromaStyle);
    debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
    return () => mediaQuery.removeEventListener('change', handler);
  });

  $: if (appTheme) updateEffectiveTheme();

  // Update CSS only when the chroma style actually changes
  $: {
    updateHighlightingCSS(currentPreviewTheme.chromaStyle);
  }

  $: if (markdown !== undefined || currentPreviewTheme !== undefined) {
    debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
  }

  // Dynamic classes for the App Frame
  $: toolbarClass = effectiveAppTheme === 'dark' ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-slate-100 border-slate-300 text-slate-900';
  $: editorClass = effectiveAppTheme === 'dark' ? 'bg-slate-900 text-slate-100 border-slate-700' : 'bg-white text-slate-900 border-slate-300';
  $: buttonClass = effectiveAppTheme === 'dark' ? 'bg-slate-700 hover:bg-slate-600' : 'bg-slate-200 hover:bg-slate-300';
  $: dividerClass = effectiveAppTheme === 'dark' ? 'bg-slate-700' : 'bg-slate-300';
</script>

<svelte:window on:mousemove={onMouseMove} on:mouseup={stopResizing} />

<main class="flex h-screen w-full overflow-hidden flex-col select-none {effectiveAppTheme === 'dark' ? 'bg-slate-900' : 'bg-white'}">
  <!-- Toolbar -->
  <div class="h-12 border-b flex items-center px-4 gap-4 shrink-0 z-20 {toolbarClass}">
    <div class="flex gap-2">
      <button on:click={handleOpen} title="Open" class="px-3 py-1 rounded text-sm transition-colors {buttonClass}">Open</button>
      <button on:click={handleSave} title="Save" class="px-3 py-1 rounded text-sm transition-colors {buttonClass}">Save</button>
    </div>

    <div class="h-6 w-px {dividerClass}"></div>

    <div class="flex gap-2 items-center">
      <span class="text-xs opacity-60">Preview Theme:</span>
      <select bind:value={currentPreviewTheme} class="text-xs rounded border-none py-1 cursor-pointer bg-transparent focus:ring-1 focus:ring-blue-500">
        {#each themes as theme}
          <option value={theme}>{theme.name}</option>
        {/each}
      </select>
    </div>

    <div class="h-6 w-px {dividerClass}"></div>

    <div class="flex gap-1 items-center">
      <span class="text-xs opacity-60 mr-1">Zoom:</span>
      <button on:click={() => adjustFontSize(-5)} class="w-6 h-6 flex items-center justify-center rounded text-sm font-bold {buttonClass}">-</button>
      <span class="text-[10px] opacity-60 w-8 text-center font-mono">{fontSize}%</span>
      <button on:click={() => adjustFontSize(5)} class="w-6 h-6 flex items-center justify-center rounded text-sm font-bold {buttonClass}">+</button>
    </div>

    <div class="flex-1"></div>

    <!-- App Theme Toggle -->
    <div class="flex items-center gap-3">
        <button
            on:click={toggleAppTheme}
            title="Toggle App Theme ({appTheme})"
            class="p-1.5 rounded-full transition-colors {buttonClass}"
        >
            {#if appTheme === 'dark'}
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
            {:else if appTheme === 'light'}
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
            {:else}
                <div class="relative">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
                    <span class="absolute -top-1 -right-1 flex h-2 w-2"><span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span><span class="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span></span>
                </div>
            {/if}
        </button>
        <span class="text-xs opacity-40 font-mono hidden sm:inline">MD Viewer v0.1</span>
    </div>
  </div>

  <div class="flex flex-1 overflow-hidden relative">
    <!-- Editor Column -->
    <div class="flex flex-col border-r relative {editorClass}" style="width: {splitWidth}%;">
      <div class="p-2 text-xs font-bold uppercase tracking-wider opacity-50 border-b shrink-0 {toolbarClass}">
        Editor
      </div>
      <textarea
        bind:value={markdown}
        spellcheck="false" autocorrect="off" autocapitalize="off"
        class="flex-1 p-4 focus:outline-none resize-none font-mono text-sm select-text bg-transparent"
        placeholder="Type markdown here..."
      ></textarea>
      <div on:mousedown={startResizing} class="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-blue-500/50 transition-colors z-10"></div>
    </div>

    <!-- Preview Column -->
    <div class="flex-1 flex flex-col relative">
      {#if isResizing} <div class="absolute inset-0 z-50"></div> {/if}
      <div class="p-2 text-xs font-bold uppercase tracking-wider opacity-50 border-b shrink-0 {toolbarClass}">
        Preview
      </div>
      <Preview html={htmlContent} css={highlightingCSS} theme={currentPreviewTheme} fontSize={fontSize} />
    </div>
  </div>
</main>

<style>
  :global(body) { user-select: none; margin: 0; }
  select option {
    background-color: white;
    color: black;
  }
</style>
