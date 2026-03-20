<script lang="ts">
  /**
   * Main Application component for MD Viewer.
   * Manages layout (resizable panes), dual-theming, file I/O, and debounced rendering.
   * Now supports multiple open files via a Tab interface.
   * Features: Word count, Focus Mode, Print to PDF.
   */
  import { onMount, tick } from 'svelte';
  import { EventsOn, OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime.js';
  import * as backend from './lib/backend';
  import Preview from './components/Preview.svelte';
  import { themes } from './themes';
  import { t, locale } from './i18n';
  import { APP_THEME, STYLE, DEFAULTS } from './lib/constants';
  import type { AppTheme_t } from './lib/constants';
  import { c_initialmd } from './lib/devdefmd.js';

  interface Tab {
    id: string;
    title: string;
    path: string | null;
    content: string;
    isDirty: boolean;
  }

  // State: Tabs
  let tabs: Tab[] = [];
  let activeTabIndex: number = 0;

  // Set initial default markdown
  const defaultMarkdown = () => $t('welcomeTitle')+c_initialmd;

  function createNewTab(title = "Untitled", content = "", path: string | null = null): Tab {
    return {
      id: Math.random().toString(36).substring(2, 11),
      title,
      path,
      content,
      isDirty: false
    };
  }

  function addNewTab() {
    const newTab = createNewTab($t('untitled'), defaultMarkdown());
    tabs = [...tabs, newTab];
    activeTabIndex = tabs.length - 1;
  }

  function handleCloseTab(index: number, event?: MouseEvent) {
    if (event) event.stopPropagation();
    
    // If it's the last tab, we might want to keep one empty one or just close
    if (tabs.length === 1) {
        tabs = [createNewTab($t('untitled'), defaultMarkdown())];
        activeTabIndex = 0;
        return;
    }

    const wasActive = index === activeTabIndex;
    tabs = tabs.filter((_, i) => i !== index);
    
    if (wasActive) {
      activeTabIndex = Math.max(0, index - 1);
    } else if (index < activeTabIndex) {
      activeTabIndex--;
    }
  }

  // Reactive derived values for active tab
  $: activeTab = tabs[activeTabIndex] || (tabs.length > 0 ? tabs[0] : null);
  $: markdown = activeTab ? activeTab.content : "";

  // Feature 6: Word & Character Count
  $: wordCount = markdown ? (markdown.trim().split(/\s+/).filter(Boolean).length) : 0;
  $: charCount = markdown ? markdown.length : 0;
  $: readingTime = Math.ceil(wordCount / 225); // Average 225 wpm

  // Mark tab as dirty when content changes via binding
  $: if (tabs[activeTabIndex] && tabs[activeTabIndex].content) {
      // This reactive block triggers on any content change in the active tab
      // We don't want to mark it dirty on the very first load, 
      // but Svelte handles the initial assignment as a change.
      // For simplicity, we just track it.
  }

  function onContentInput() {
    if (tabs[activeTabIndex]) {
        tabs[activeTabIndex].isDirty = true;
        // Trigger debounced update via reactive markdown dependency
    }
  }

  // State: Rendering results
  let htmlContent: string = "";
  let highlightingCSS: string = "";

  // State: Visual Preferences
  let currentPreviewTheme = themes[0]; // Active Markdown theme
  let fontSize: number = DEFAULTS.fontSize; // Active zoom level
  // UI: transient drop/toast message
  let dropMessage: string | null = null;

  // Feature 7: Focus Mode & Editor Toggle
  let isFocusMode = false;
  let isEditorHidden = false;

  // Feature 5: Print State
  let isPrinting = false;

  // App Frame Theming (Toolbar/Editor frame)
  let appTheme: AppTheme_t = APP_THEME.DARK;
  let effectiveAppTheme: 'dark' | 'light' = APP_THEME.DARK as 'dark';

  // Initialization flag
  let isReady = false;

  /**
   * checkWailsReady verifies if the Wails 'go' object is available on window.
   */
  const checkWailsReady = backend.isWailsReady;

  function updateEffectiveTheme() {
    if (appTheme === 'auto') {
      effectiveAppTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    } else {
      effectiveAppTheme = appTheme;
    }
  }

  function toggleAppTheme() {
    if (appTheme === APP_THEME.DARK) appTheme = APP_THEME.LIGHT;
    else if (appTheme === APP_THEME.LIGHT) appTheme = APP_THEME.AUTO;
    else appTheme = APP_THEME.DARK;
  }

  // Layout: Resizable Split Panes
  let splitWidth: number = 50;
  let isResizing = false;

  function startResizing() { isResizing = true; }
  function stopResizing() { isResizing = false; }
  function onMouseMove(event: MouseEvent) {
    if (!isResizing || isEditorHidden) return;
    splitWidth = (event.clientX / window.innerWidth) * 100;
    if (splitWidth < 10) splitWidth = 10;
    if (splitWidth > 90) splitWidth = 90;
  }

  // Engine: Debounced Rendering
  let timeout: ReturnType<typeof setTimeout>;

  function debouncedUpdate(value: string, themeStyle: string) {
    if (!isReady || !checkWailsReady()) return;
    clearTimeout(timeout);
    timeout = setTimeout(async () => {
      try {
        htmlContent = await backend.renderMarkdown(value, themeStyle);
      } catch (err) {
        console.error("Failed to render markdown:", err);
      }
    }, 300);
  }

  async function updateHighlightingCSS(style: string) {
    if (!isReady || !checkWailsReady()) return;
    try {
      highlightingCSS = await backend.getStyleCSS(style);
    } catch (err) {
      console.error("Failed to get highlighting CSS:", err);
    }
  }

  // Native Bindings: File Operations
  async function handleOpen() {
    if (!isReady || !checkWailsReady()) return;
    try {
      const result = await backend.openFile();
      if (result) {
        const title = await backend.getFileTitle(result.path);
        const newTab = createNewTab(title, result.content, result.path);
        tabs = [...tabs, newTab];
        activeTabIndex = tabs.length - 1;
      }
    } catch (err) { console.error("Failed to open file:", err); }
  }

  async function handleSave() {
    if (!isReady || !checkWailsReady()) return;
    try { 
      const path = await backend.saveFile(markdown); 
      if (path && activeTab) {
        const title = await backend.getFileTitle(path);
        tabs[activeTabIndex].path = path;
        tabs[activeTabIndex].title = title;
        tabs[activeTabIndex].isDirty = false;
        tabs = [...tabs]; // trigger reactivity
      }
    } catch (err) { console.error("Failed to save file:", err); }
  }

  async function handleExport() {
    if (!isReady || !checkWailsReady()) return;
    try {
      // Extract current theme colors from the DOM for the export
      const previewEl = document.querySelector('.prose');
      const containerEl = previewEl?.parentElement;

      let themeVars = "";
      if (previewEl && containerEl) {
        const pStyle = window.getComputedStyle(previewEl);
        const cStyle = window.getComputedStyle(containerEl);
        const aStyle = window.getComputedStyle(previewEl.querySelector('a') || previewEl);
        const codeStyle = window.getComputedStyle(previewEl.querySelector('code') || previewEl);
        const isDark = effectiveAppTheme === 'dark' || currentPreviewTheme.id === 'dark';

        themeVars = `
        :root {
            --bg-color: ${cStyle.backgroundColor};
            --text-color: ${pStyle.color};
            --link-color: ${aStyle.color !== pStyle.color ? aStyle.color : (isDark ? '#58a6ff' : '#0969da')};
            --border-color: ${isDark ? '#30363d' : '#dfe2e5'};
            --code-bg: ${codeStyle.backgroundColor !== 'rgba(0, 0, 0, 0)' ? codeStyle.backgroundColor : (isDark ? '#161b22' : '#f6f8fa')};
            --alert-bg: ${isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)'};
        }
        `;
      }

      await backend.exportHTML(htmlContent, themeVars + highlightingCSS);
    } catch (err) {
      console.error("Failed to export HTML:", err);
    }
  }

  // Feature 5: Print to PDF
  async function handlePrint() {
    const originalTheme = currentPreviewTheme;
    const monochromeTheme = themes.find(t => t.id === 'monochrome') || originalTheme;
    
    isPrinting = true;
    currentPreviewTheme = monochromeTheme;
    
    // Wait for Svelte to update DOM with monochrome theme
    await tick();
    // Small delay to ensure Mermaid/Katex re-renders if needed (though they react to theme changes)
    setTimeout(() => {
        window.print();
        isPrinting = false;
        currentPreviewTheme = originalTheme;
    }, 100);
  }

  function adjustFontSize(delta: number) {
    fontSize = Math.min(Math.max(fontSize + delta, 50), 200);
  }

  // Lifecycle & Reactivity
  onMount(() => {
    const init = async () => {
      if (checkWailsReady()) {
        isReady = true;

        // Listen for Wails native Drag and Drop files
        OnFileDrop(async (x: number, y: number, paths: string[]) => {
          if (!paths || paths.length === 0) return;
          const allowedExt = /\.(md|markdown|mdown|mkd|mdx)$/i;
          
          let loadedCount = 0;
          for (const path of paths) {
            if (allowedExt.test(path)) {
              try {
                const content = await backend.readFile(path);
                if (content !== undefined && content !== null) {
                  const title = await backend.getFileTitle(path);
                  const newTab = createNewTab(title, content, path);
                  tabs = [...tabs, newTab];
                  activeTabIndex = tabs.length - 1;
                  loadedCount++;
                }
              } catch (err) {
                console.error("Failed to read dropped file:", err);
              }
            }
          }
          if (loadedCount > 0) {
            dropMessage = `Loaded ${loadedCount} files`;
            setTimeout(() => dropMessage = null, 3000);
          }
        }, false);

        // Check for file passed via command line
        const result = await backend.getInitialContent();
        if (result) {
          const title = await backend.getFileTitle(result.path);
          tabs = [createNewTab(title, result.content, result.path)];
        } else {
          tabs = [createNewTab($t('welcomeTitle'), defaultMarkdown())];
        }
        activeTabIndex = 0;

        updateHighlightingCSS(currentPreviewTheme.chromaStyle);
        debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
      } else {
        setTimeout(init, 50);
      }
    };

    init();
    updateEffectiveTheme();
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => { if (appTheme === 'auto') updateEffectiveTheme(); };
    mediaQuery.addEventListener('change', handler);
    return () => { mediaQuery.removeEventListener('change', handler); OnFileDropOff(); };
  });

  $: if (appTheme) updateEffectiveTheme();
  $: if (isReady) updateHighlightingCSS(currentPreviewTheme.chromaStyle);
  $: if (isReady && (markdown !== undefined || currentPreviewTheme !== undefined)) {
    debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
  }

  $: toolbarClass = STYLE.toolbar[effectiveAppTheme];
  $: editorClass = STYLE.editor[effectiveAppTheme];
  $: buttonClass = STYLE.button[effectiveAppTheme];
  $: dividerClass = STYLE.divider[effectiveAppTheme];
  $: activeTabClass = STYLE.tab.active[effectiveAppTheme];
  $: inactiveTabClass = STYLE.tab.inactive[effectiveAppTheme];
  $: statusClass = STYLE.status[effectiveAppTheme];
  $: focusButtonClass = STYLE.focusButton[effectiveAppTheme];
</script>

<svelte:window on:mousemove={onMouseMove} on:mouseup={stopResizing} />

<main class="flex h-screen w-full overflow-hidden flex-col select-none {isPrinting ? 'is-printing' : ''} {effectiveAppTheme === 'dark' ? 'bg-slate-900' : 'bg-white'}">
  <!-- Top Navigation / Toolbar -->
  {#if !isFocusMode && !isPrinting}
  <div class="h-12 border-b flex items-center px-4 gap-4 shrink-0 z-20 {toolbarClass} print:hidden">
    <div class="flex gap-2">
      <button on:click={handleOpen} title={$t('open')} class="p-2 rounded transition-colors {buttonClass}">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
      </button>
      <button on:click={handleSave} title={$t('save')} class="p-2 rounded transition-colors {buttonClass}">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
      </button>
      <button on:click={addNewTab} title="New Tab" class="p-2 rounded transition-colors {buttonClass}">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
      </button>
    </div>

    <div class="flex-1"></div>

    <div class="flex items-center gap-4">
        <div class="flex items-center">
          <select bind:value={$locale} class="text-[10px] font-bold uppercase rounded border-none py-1 px-2 cursor-pointer bg-transparent focus:ring-1 focus:ring-blue-500 {effectiveAppTheme === 'dark' ? 'text-slate-100' : 'text-slate-900'}">
            <option value="en" class={effectiveAppTheme === 'dark' ? 'bg-slate-800 text-white' : 'bg-white text-black'}>EN</option>
            <option value="de" class={effectiveAppTheme === 'dark' ? 'bg-slate-800 text-white' : 'bg-white text-black'}>DE</option>
            <option value="es" class={effectiveAppTheme === 'dark' ? 'bg-slate-800 text-white' : 'bg-white text-black'}>ES</option>
            <option value="fr" class={effectiveAppTheme === 'dark' ? 'bg-slate-800 text-white' : 'bg-white text-black'}>FR</option>
          </select>
        </div>

        <div class="h-6 w-px {dividerClass}"></div>

        <button on:click={toggleAppTheme} title={$t('toggleTheme')} class="p-1.5 rounded-full transition-colors {buttonClass}">
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
        <span class="text-xs opacity-40 font-mono hidden sm:inline">MD Viewer v0.5.4</span>
    </div>
  </div>

  <!-- Tabs Bar -->
  <div class="flex overflow-x-auto no-scrollbar border-b {toolbarClass} print:hidden">
    {#each tabs as tab, i}
      <div 
        class="flex items-center px-4 h-9 cursor-pointer transition-colors border-r text-xs font-medium min-w-[120px] max-w-[200px] {i === activeTabIndex ? activeTabClass : inactiveTabClass} {dividerClass}"
        on:click={() => activeTabIndex = i}
      >
        <span class="truncate flex-1">{tab.title}{tab.isDirty ? ' *' : ''}</span>
        <button 
          on:click={(e) => handleCloseTab(i, e)}
          class="ml-2 p-0.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        </button>
      </div>
    {/each}
  </div>
  {/if}

  <div class="flex flex-1 overflow-hidden relative print:block">
    {#if !isEditorHidden && !isFocusMode && !isPrinting}
    <div class="flex flex-col border-r relative {editorClass} print:hidden" style="width: {splitWidth}%;">
      <div class="p-2 text-xs font-bold uppercase tracking-wider opacity-50 border-b shrink-0 {toolbarClass}">
        {$t('editor')}
      </div>
      {#if tabs[activeTabIndex]}
      <textarea
        bind:value={tabs[activeTabIndex].content}
        on:input={onContentInput}
        spellcheck="false" autocorrect="off" autocapitalize="off"
        class="flex-1 p-4 focus:outline-none resize-none font-mono text-sm select-text bg-transparent"
        placeholder={$t('placeholder')}
      ></textarea>
      {/if}
      <div on:mousedown={startResizing} class="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-blue-500/50 transition-colors z-10"></div>
    </div>
    {/if}

    <div class="flex-1 flex flex-col relative print:block">
      {#if isResizing} <div class="absolute inset-0 z-50"></div> {/if}
      
      <!-- Preview Toolbar -->
      {#if !isPrinting}
      <div class="p-2 h-10 border-b flex items-center px-4 gap-4 shrink-0 {toolbarClass} print:hidden">
        <div class="flex items-center gap-2">
            <button 
                on:click={() => isEditorHidden = !isEditorHidden} 
                class="p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors {isEditorHidden ? focusButtonClass : ''}"
                title={isEditorHidden ? "Show Editor" : "Hide Editor"}
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="9" y1="3" x2="9" y2="21"></line></svg>
            </button>
            <button 
                on:click={() => isFocusMode = !isFocusMode} 
                class="p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors {isFocusMode ? focusButtonClass : ''}"
                title={isFocusMode ? "Exit Focus Mode" : "Focus Mode"}
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"></path></svg>
            </button>
        </div>

        <div class="text-xs font-bold uppercase tracking-wider opacity-50">{$t('preview')}</div>
        <div class="flex-1"></div>

        <button on:click={handlePrint} title="Print / PDF" class="p-1 rounded transition-colors {buttonClass}">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 6 2 18 2 18 9"></polyline><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"></path><rect x="6" y="14" width="12" height="8"></rect></svg>
        </button>

        <button on:click={handleExport} title={$t('export')} class="p-1 rounded transition-colors {buttonClass}">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
        </button>

        <div class="h-4 w-px {dividerClass}"></div>

        <div class="flex gap-2 items-center">
          <span class="text-[10px] uppercase opacity-60 font-bold">Theme:</span>
          <select bind:value={currentPreviewTheme} class="text-xs rounded border-none py-0.5 cursor-pointer bg-transparent focus:ring-1 focus:ring-blue-500">
            {#each themes as theme}
              <option value={theme}>{theme.name}</option>
            {/each}
          </select>
        </div>
        <div class="h-4 w-px {dividerClass}"></div>
        <div class="flex gap-1 items-center">
          <button on:click={() => adjustFontSize(-5)} class="w-5 h-5 flex items-center justify-center rounded text-xs font-bold {buttonClass}">-</button>
          <span class="text-[10px] opacity-60 w-8 text-center font-mono">{fontSize}%</span>
          <button on:click={() => adjustFontSize(5)} class="w-5 h-5 flex items-center justify-center rounded text-xs font-bold {buttonClass}">+</button>
        </div>
      </div>
      {/if}
      <Preview html={htmlContent} css={highlightingCSS} theme={currentPreviewTheme} fontSize={fontSize} />
    </div>
  </div>

  <!-- Feature 6: Status Bar -->
  {#if !isPrinting}
  <div class="h-6 border-t flex items-center px-4 gap-6 shrink-0 text-[10px] font-medium {statusClass} print:hidden">
    <div class="flex gap-4">
        <span>{wordCount} words</span>
        <span>{charCount} characters</span>
    </div>
    <div class="flex-1"></div>
    <div>Approx. {readingTime} min read</div>
    <div class="h-3 w-px {dividerClass}"></div>
    <div class="uppercase tracking-tighter opacity-80">{activeTab?.path || 'Untitled'}</div>
  </div>
  {/if}

  {#if dropMessage}
    <div class="drop-toast">{dropMessage}</div>
  {/if}
</main>

<style>
  :global(body) { margin: 0; }
  select option { background-color: white; color: black; }
  :global(.bg-slate-900) select option, :global(.bg-slate-800) select option { background-color: #1e293b; color: white; }
  
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .no-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }

  .drop-toast {
    position: absolute;
    bottom: 3.5rem;
    left: 50%;
    transform: translateX(-50%);
    background: #3b82f6;
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    font-weight: 500;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
    z-index: 50;
    animation: slideUp 0.3s ease-out;
  }

  @keyframes slideUp {
    from { transform: translate(-50%, 100%); opacity: 0; }
    to { transform: translate(-50%, 0); opacity: 1; }
  }

  @media print {
    :global(body), main.is-printing {
        background: white !important;
        background-color: white !important;
        color: black !important;
        height: auto !important;
        overflow: visible !important;
        display: block !important;
        filter: grayscale(100%) !important;
    }
  }
</style>