<script lang="ts">
  /**
   * Main Application component for MD Viewer.
   * Manages layout (resizable panes), dual-theming, file I/O, and debounced rendering.
   */
  import { onMount } from 'svelte';
  import { RenderMarkdown, OpenFile, SaveFile, GetStyleCSS } from '../wailsjs/go/main/App.js'
  import Preview from './components/Preview.svelte';
  import { themes } from './themes';
  import { t, locale } from './i18n';

  // State: Markdown Content
  let markdown: string = "";

  // Set initial markdown based on locale
  $: if (markdown === "" && $t('welcomeTitle')) {
    markdown = `${$t('welcomeTitle')}

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
  }

  // State: Rendering results
  let htmlContent: string = "";
  let highlightingCSS: string = "";
  
  // State: Visual Preferences
  let currentPreviewTheme = themes[0]; // Active Markdown theme
  let fontSize: number = 90; // Active zoom level
  
  // App Frame Theming (Toolbar/Editor frame)
  type AppTheme = 'dark' | 'light' | 'auto';
  let appTheme: AppTheme = 'dark';
  let effectiveAppTheme: 'dark' | 'light' = 'dark';

  // Initialization flag to prevent calling Go bindings before Wails is ready
  let isReady = false;

  /**
   * checkWailsReady verifies if the Wails 'go' object is available on window.
   */
  function checkWailsReady(): boolean {
    return typeof window !== 'undefined' && 
           (window as any).go !== undefined && 
           (window as any).go.main !== undefined;
  }

  /**
   * updateEffectiveTheme calculates the actual theme based on 'auto' settings.
   */
  function updateEffectiveTheme() {
    if (appTheme === 'auto') {
      effectiveAppTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    } else {
      effectiveAppTheme = appTheme;
    }
  }

  /**
   * toggleAppTheme cycles through dark -> light -> auto.
   */
  function toggleAppTheme() {
    if (appTheme === 'dark') appTheme = 'light';
    else if (appTheme === 'light') appTheme = 'auto';
    else appTheme = 'dark';
  }

  // Layout: Resizable Split Panes
  let splitWidth: number = 50; // Current width of the editor in %
  let isResizing = false;

  function startResizing() { isResizing = true; }
  function stopResizing() { isResizing = false; }
  function onMouseMove(event: MouseEvent) {
    if (!isResizing) return;
    splitWidth = (event.clientX / window.innerWidth) * 100;
    // Enforce min/max constraints
    if (splitWidth < 10) splitWidth = 10;
    if (splitWidth > 90) splitWidth = 90;
  }

  // Engine: Debounced Rendering to prevent UI flicker
  let timeout: ReturnType<typeof setTimeout>;
  
  /**
   * debouncedUpdate sends content to the backend for conversion after a short delay.
   */
  function debouncedUpdate(value: string, themeStyle: string) {
    if (!isReady || !checkWailsReady()) return;
    clearTimeout(timeout);
    timeout = setTimeout(async () => {
      try {
        htmlContent = await RenderMarkdown(value, themeStyle);
      } catch (err) {
        console.error("Failed to render markdown:", err);
      }
    }, 300);
  }

  /**
   * updateHighlightingCSS fetches the raw CSS required for the selected Chroma style.
   */
  async function updateHighlightingCSS(style: string) {
    if (!isReady || !checkWailsReady()) return;
    try {
      highlightingCSS = await GetStyleCSS(style);
    } catch (err) {
      console.error("Failed to get highlighting CSS:", err);
    }
  }

  // Native Bindings: File Operations
  async function handleOpen() {
    if (!isReady || !checkWailsReady()) return;
    try {
      const content = await OpenFile();
      if (content !== undefined && content !== null) { markdown = content; }
    } catch (err) { console.error("Failed to open file:", err); }
  }

  async function handleSave() {
    if (!isReady || !checkWailsReady()) return;
    try { await SaveFile(markdown); } catch (err) { console.error("Failed to save file:", err); }
  }

  // Helpers
  function adjustFontSize(delta: number) {
    fontSize = Math.min(Math.max(fontSize + delta, 50), 200);
  }

  // Lifecycle & Reactivity
  onMount(() => {
    // Initial check
    if (checkWailsReady()) {
      isReady = true;
    } else {
      // Fallback: keep checking until it's ready (common in Vite dev mode)
      const interval = setInterval(() => {
        if (checkWailsReady()) {
          isReady = true;
          clearInterval(interval);
          // Trigger initial render once ready
          updateHighlightingCSS(currentPreviewTheme.chromaStyle);
          debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
        }
      }, 50);
    }
    
    updateEffectiveTheme();
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => { if (appTheme === 'auto') updateEffectiveTheme(); };
    mediaQuery.addEventListener('change', handler);
    
    // If already ready, trigger initial data fetch
    if (isReady) {
      updateHighlightingCSS(currentPreviewTheme.chromaStyle);
      debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
    }
    
    return () => mediaQuery.removeEventListener('change', handler);
  });

  // Automatically recalculate styles and render on state changes
  $: if (appTheme) updateEffectiveTheme();
  
  // Guard reactive statements with both isReady and checkWailsReady
  $: if (isReady && checkWailsReady()) {
    updateHighlightingCSS(currentPreviewTheme.chromaStyle);
  }
  
  $: if (isReady && checkWailsReady() && (markdown !== undefined || currentPreviewTheme !== undefined)) {
    debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
  }

  // Computed classes for the App Frame styling
  $: toolbarClass = effectiveAppTheme === 'dark' ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-slate-100 border-slate-300 text-slate-900';
  $: editorClass = effectiveAppTheme === 'dark' ? 'bg-slate-900 text-slate-100 border-slate-700' : 'bg-white text-slate-900 border-slate-300';
  $: buttonClass = effectiveAppTheme === 'dark' ? 'bg-slate-700 hover:bg-slate-600' : 'bg-slate-200 hover:bg-slate-300';
  $: dividerClass = effectiveAppTheme === 'dark' ? 'bg-slate-700' : 'bg-slate-300';
</script>

<svelte:window on:mousemove={onMouseMove} on:mouseup={stopResizing} />

<main class="flex h-screen w-full overflow-hidden flex-col select-none {effectiveAppTheme === 'dark' ? 'bg-slate-900' : 'bg-white'}">
  <!-- Top Navigation / Toolbar -->
  <div class="h-12 border-b flex items-center px-4 gap-4 shrink-0 z-20 {toolbarClass}">
    <div class="flex gap-2">
      <button on:click={handleOpen} title={$t('open')} class="p-2 rounded transition-colors {buttonClass}">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
      </button>
      <button on:click={handleSave} title={$t('save')} class="p-2 rounded transition-colors {buttonClass}">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
      </button>
    </div>
    
    <div class="flex-1"></div>
    
    <!-- Top Right Controls (Language & Theme) -->
    <div class="flex items-center gap-4">
        <!-- Language Selector -->
        <div class="flex items-center">
          <select 
            bind:value={$locale} 
            class="text-[10px] font-bold uppercase rounded border-none py-1 px-2 cursor-pointer bg-transparent focus:ring-1 focus:ring-blue-500 {effectiveAppTheme === 'dark' ? 'text-slate-100' : 'text-slate-900'}"
          >
            <option value="en" class={effectiveAppTheme === 'dark' ? 'bg-slate-800 text-white' : 'bg-white text-black'}>EN</option>
            <option value="de" class={effectiveAppTheme === 'dark' ? 'bg-slate-800 text-white' : 'bg-white text-black'}>DE</option>
            <option value="es" class={effectiveAppTheme === 'dark' ? 'bg-slate-800 text-white' : 'bg-white text-black'}>ES</option>
            <option value="fr" class={effectiveAppTheme === 'dark' ? 'bg-slate-800 text-white' : 'bg-white text-black'}>FR</option>
          </select>
        </div>

        <div class="h-6 w-px {dividerClass}"></div>

        <!-- App Theme Toggle (SVG Icons) -->
        <button 
            on:click={toggleAppTheme} 
            title={$t('toggleTheme')} 
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
    <!-- Editor Pane -->
    <div class="flex flex-col border-r relative {editorClass}" style="width: {splitWidth}%;">
      <div class="p-2 text-xs font-bold uppercase tracking-wider opacity-50 border-b shrink-0 {toolbarClass}">
        {$t('editor')}
      </div>
      <textarea
        bind:value={markdown}
        spellcheck="false" autocorrect="off" autocapitalize="off"
        class="flex-1 p-4 focus:outline-none resize-none font-mono text-sm select-text bg-transparent"
        placeholder={$t('placeholder')}
      ></textarea>
      
      <!-- Draggable Splitter Handle -->
      <div on:mousedown={startResizing} class="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-blue-500/50 transition-colors z-10"></div>
    </div>

    <!-- Preview Pane -->
    <div class="flex-1 flex flex-col relative">
      {#if isResizing} <div class="absolute inset-0 z-50"></div> {/if}
      
      <!-- Preview Header with local controls -->
      <div class="p-2 h-10 border-b flex items-center px-4 gap-4 shrink-0 {toolbarClass}">
        <div class="text-xs font-bold uppercase tracking-wider opacity-50">
            {$t('preview')}
        </div>
        <div class="flex-1"></div>
        
        <!-- Preview Theme Selector -->
        <div class="flex gap-2 items-center">
          <span class="text-[10px] uppercase opacity-60 font-bold">Theme:</span>
          <select bind:value={currentPreviewTheme} class="text-xs rounded border-none py-0.5 cursor-pointer bg-transparent focus:ring-1 focus:ring-blue-500">
            {#each themes as theme}
              <option value={theme}>{theme.name}</option>
            {/each}
          </select>
        </div>

        <div class="h-4 w-px {dividerClass}"></div>

        <!-- Zoom Controls -->
        <div class="flex gap-1 items-center">
          <button on:click={() => adjustFontSize(-5)} class="w-5 h-5 flex items-center justify-center rounded text-xs font-bold {buttonClass}">-</button>
          <span class="text-[10px] opacity-60 w-8 text-center font-mono">{fontSize}%</span>
          <button on:click={() => adjustFontSize(5)} class="w-5 h-5 flex items-center justify-center rounded text-xs font-bold {buttonClass}">+</button>
        </div>
      </div>

      <Preview html={htmlContent} css={highlightingCSS} theme={currentPreviewTheme} fontSize={fontSize} />
    </div>
  </div>
</main>

<style>
  :global(body) { user-select: none; margin: 0; }
  
  /* Ensure dropdown options are styled correctly in dark mode */
  select option {
    background-color: white;
    color: black;
  }
  
  :global(.bg-slate-900) select option,
  :global(.bg-slate-800) select option {
    background-color: #1e293b;
    color: white;
  }
</style>
