<script lang="ts">
  /**
   * Main Application component for MD Viewer.
   * Refactored for Svelte 5 Runes, CodeMirror 6 markdown editor, dark/light themes, and crisp print output.
   */
  import { onMount, tick, untrack } from 'svelte';
  import { get } from 'svelte/store';
  import { EventsOn, EventsOff, OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime.js';
  import * as backend from './lib/backend';
  
  // Components
  import Editor from './components/Editor.svelte';
  import Preview from './components/Preview.svelte';
  import WhitelistModal from './components/WhitelistModal.svelte';
  import Toolbar from './components/Toolbar.svelte';
  import TabsBar from './components/TabsBar.svelte';
  import StatusBar from './components/StatusBar.svelte';

  // State & Config
  import { getTheme } from './themes';
  import { t, locale, translations } from './i18n';
  import { APP_THEME, STYLE, DEFAULTS } from './lib/constants';
  import { c_initialmd } from './lib/devdefmd.js';
  import { 
    appTheme, effectiveAppTheme, splitWidth, isFocusMode, 
    isEditorHidden, isPrinting, dropMessage, showToast 
  } from './lib/stores';

  // CONFIGURATION: Show HTML toolbar in Vite development mode, hide in production.
  const SHOW_HTML_TOOLBAR = import.meta.env.DEV;

  interface Tab {
    id: string;
    title: string;
    path: string | null;
    content: string;
    isDirty: boolean;
  }

  // --- Svelte 5 Runes: State ---
  let tabs = $state<Tab[]>([]);
  let activeTabIndex = $state(0);
  let isReady = $state(false);
  let htmlContent = $state("");
  let highlightingCSS = $state("");
  let fontSize = $state(DEFAULTS.fontSize);
  
  let showSecurityModal = $state(false);
  let securityType = $state<'path' | 'url'>('path');
  let securityResource = $state("");

  let isSyncScroll = $state(true);
  let editorComponent: any = $state();
  let previewComponent: any = $state();
  let scrollLock = false;

  function handleEditorScroll(percentage: number) {
    if (!isSyncScroll || scrollLock || !previewComponent) return;
    scrollLock = true;
    previewComponent.setScrollPercentage(percentage);
    setTimeout(() => { scrollLock = false; }, 50);
  }

  function handlePreviewScroll() {
    if (!isSyncScroll || scrollLock || !editorComponent || !previewComponent) return;
    scrollLock = true;
    const percentage = previewComponent.getScrollPercentage();
    editorComponent.setScrollPercentage(percentage);
    setTimeout(() => { scrollLock = false; }, 50);
  }

  // --- Svelte 5 Runes: Derived ---
  const markdown = $derived(tabs[activeTabIndex]?.content || "");
  const activeTab = $derived(tabs[activeTabIndex] || null);
  const currentPreviewTheme = $derived(getTheme($effectiveAppTheme));

  const wordCount = $derived(markdown ? (markdown.trim().split(/\s+/).filter(Boolean).length) : 0);
  const charCount = $derived(markdown ? markdown.length : 0);
  const readingTime = $derived(Math.ceil(wordCount / 225));

  const toolbarClass = $derived(STYLE.toolbar[$effectiveAppTheme]);
  const editorClass = $derived(STYLE.editor[$effectiveAppTheme]);
  const buttonClass = $derived(STYLE.button[$effectiveAppTheme]);
  const dividerClass = $derived(STYLE.divider[$effectiveAppTheme]);
  const focusButtonClass = $derived(STYLE.focusButton[$effectiveAppTheme]);

  const defaultMarkdown = () => $t('welcomeTitle') + c_initialmd;

  function createNewTab(title = $t('untitled'), content = "", path: string | null = null): Tab {
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

  function onContentInput() {
    if (tabs[activeTabIndex]) {
      tabs[activeTabIndex].isDirty = true;
    }
  }

  // Formatting Helpers for native menu and toolbar shortcuts
  function wrapSelection(prefix: string, suffix: string) {
    if (editorComponent?.wrapSelection) {
      editorComponent.wrapSelection(prefix, suffix);
      onContentInput();
    }
  }

  function prefixSelection(prefix: string) {
    if (editorComponent?.prefixSelection) {
      editorComponent.prefixSelection(prefix);
      onContentInput();
    }
  }

  async function handleSecurityRequest(detail: { type: 'path' | 'url', resource: string }) {
    securityType = detail.type;
    securityResource = detail.resource;
    showSecurityModal = true;
  }

  async function confirmSecurityRequest() {
    if (securityType === 'path') {
      await backend.addPathToWhitelist(securityResource);
    } else {
      await backend.addURLToWhitelist(securityResource);
    }
    showSecurityModal = false;
    debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
  }

  async function handleOpenExternalMD(detail: { path: string }) {
    const path = detail.path;
    try {
      const content = await backend.readFile(path);
      if (content !== undefined) {
        const title = await backend.getFileTitle(path);
        const newTab = createNewTab(title, content, path);
        tabs = [...tabs, newTab];
        activeTabIndex = tabs.length - 1;
      }
    } catch (err) {
      console.error("Failed to open linked markdown:", err);
    }
  }

  const checkWailsReady = backend.isWailsReady;

  function updateEffectiveTheme() {
    if ($appTheme === 'auto') {
      const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      effectiveAppTheme.set(isDark ? 'dark' : 'light');
    } else {
      effectiveAppTheme.set($appTheme as 'dark' | 'light');
    }
  }

  function toggleAppTheme() {
    appTheme.update(current => {
      if (current === APP_THEME.DARK) return APP_THEME.LIGHT;
      if (current === APP_THEME.LIGHT) return APP_THEME.AUTO;
      return APP_THEME.DARK;
    });
  }

  // Native Menu Translation Sync
  async function updateNativeMenu(currentLocale: string) {
    if (!isReady || !checkWailsReady()) return;
    const tMap = translations[currentLocale];
    if (!tMap) return;
    const menuTranslations = {
      menuFile: tMap.menuFile,
      menuEdit: tMap.menuEdit,
      menuView: tMap.menuView,
      menuFormat: tMap.menuFormat,
      menuLanguage: tMap.menuLanguage,
      menuAppearance: tMap.menuAppearance,
      menuThemeDark: tMap.menuThemeDark,
      menuThemeLight: tMap.menuThemeLight,
      menuThemeAuto: tMap.menuThemeAuto,
      menuNewTab: tMap.menuNewTab,
      menuOpen: tMap.menuOpen,
      menuRecentFiles: tMap.menuRecentFiles,
      menuNoRecentFiles: tMap.menuNoRecentFiles,
      menuSave: tMap.menuSave,
      menuUndo: tMap.menuUndo,
      menuRedo: tMap.menuRedo,
      menuCut: tMap.menuCut,
      menuCopy: tMap.menuCopy,
      menuPaste: tMap.menuPaste,
      menuBold: tMap.menuBold,
      menuItalic: tMap.menuItalic,
      menuCodeBlock: tMap.menuCodeBlock,
      menuAbout: tMap.menuAbout,
      aboutTitle: tMap.aboutTitle,
      aboutBody: tMap.aboutBody
    };
    await backend.updateMenu(menuTranslations);
  }

  let isResizing = $state(false);
  function startResizing() { isResizing = true; }
  function stopResizing() { isResizing = false; }
  function onMouseMove(event: MouseEvent) {
    if (!isResizing || $isEditorHidden) return;
    splitWidth.set((event.clientX / window.innerWidth) * 100);
    if ($splitWidth < 10) splitWidth.set(10);
    if ($splitWidth > 90) splitWidth.set(90);
  }

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

  async function handleOpen() {
    if (!isReady || !checkWailsReady()) return;
    try {
      const result = await backend.openFile();
      if (result) {
        const title = await backend.getFileTitle(result.path);
        const newTab = createNewTab(title, result.content, result.path);
        tabs = [...tabs, newTab];
        activeTabIndex = tabs.length - 1;
        const parentDir = await backend.getParentDir(result.path);
        await backend.addPathToWhitelist(parentDir);
        updateNativeMenu(get(locale));
      }
    } catch (err) { console.error("Failed to open file:", err); }
  }

  async function handleSave() {
    if (!isReady || !checkWailsReady()) return;
    try { 
      const path = await backend.saveFile(markdown); 
      if (path && activeTab) {
        tabs[activeTabIndex].path = path;
        tabs[activeTabIndex].title = await backend.getFileTitle(path);
        tabs[activeTabIndex].isDirty = false;
        const parentDir = await backend.getParentDir(path);
        await backend.addPathToWhitelist(parentDir);
        updateNativeMenu(get(locale));
      }
    } catch (err) { console.error("Failed to save file:", err); }
  }

  async function handleExport() {
    if (!isReady || !checkWailsReady()) return;
    try {
      const isDark = $effectiveAppTheme === 'dark';
      const themeVars = `
        :root {
            --bg-color: ${isDark ? '#0f172a' : '#ffffff'};
            --text-color: ${isDark ? '#f1f5f9' : '#0f172a'};
            --link-color: ${isDark ? '#60a5fa' : '#0969da'};
            --border-color: ${isDark ? '#334155' : '#e2e8f0'};
            --code-bg: ${isDark ? '#1e293b' : '#f8fafc'};
            --alert-bg: ${isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)'};
        }
      `;
      await backend.exportHTML(htmlContent, themeVars + highlightingCSS);
    } catch (err) {
      console.error("Failed to export HTML:", err);
    }
  }

  function handlePrint() {
    window.print();
  }

  function adjustFontSize(delta: number) {
    fontSize = Math.min(Math.max(fontSize + delta, 50), 200);
  }

  async function handleOpenRecent(path: string) {
    if (!isReady || !checkWailsReady()) return;
    try {
      const content = await backend.readFile(path);
      if (content !== undefined) {
        const title = await backend.getFileTitle(path);
        const newTab = createNewTab(title, content, path);
        tabs = [...tabs, newTab];
        activeTabIndex = tabs.length - 1;
        const parentDir = await backend.getParentDir(path);
        await backend.addPathToWhitelist(parentDir);
        updateNativeMenu(get(locale));
      }
    } catch (err) {
      console.error("Failed to open recent file:", err);
    }
  }

  onMount(() => {
    const init = async () => {
      const wailsReady = checkWailsReady();
      if (wailsReady || import.meta.env.DEV) {
        if (wailsReady) {
          isReady = true;
          EventsOn("menu-open-file", handleOpen);
          EventsOn("menu-open-recent", handleOpenRecent);
          EventsOn("menu-save-file", handleSave);
          EventsOn("menu-new-tab", addNewTab);
          EventsOn("format-bold", () => wrapSelection('**', '**'));
          EventsOn("format-italic", () => wrapSelection('*', '*'));
          EventsOn("format-h1", () => prefixSelection('# '));
          EventsOn("format-h2", () => prefixSelection('## '));
          EventsOn("format-h3", () => prefixSelection('### '));
          EventsOn("format-code", () => wrapSelection('\n```\n', '\n```\n'));
          EventsOn("set-locale", (l: string) => locale.set(l));
          EventsOn("set-theme", (t: string) => appTheme.set(t as any));

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
                    const parentDir = await backend.getParentDir(path);
                    await backend.addPathToWhitelist(parentDir);
                  }
                } catch (err) { console.error("Failed to read dropped file:", err); }
              }
            }
            if (loadedCount > 0) {
              showToast($t('filesLoaded', loadedCount));
              updateNativeMenu(get(locale));
            }
          }, false);
        }

        const result = wailsReady ? await backend.getInitialContent() : null;
        if (result) {
          const title = await backend.getFileTitle(result.path);
          tabs = [createNewTab(title, result.content, result.path)];
          const parentDir = await backend.getParentDir(result.path);
          await backend.addPathToWhitelist(parentDir);
        } else {
          tabs = [createNewTab($t('untitled'), defaultMarkdown())];
        }
        activeTabIndex = 0;

        if (import.meta.env.DEV && !wailsReady) isReady = true;

        updateHighlightingCSS(currentPreviewTheme.chromaStyle);
        debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
        if (wailsReady) updateNativeMenu(get(locale));
      } else {
        setTimeout(init, 50);
      }
    };
    init();
    updateEffectiveTheme();
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => { if ($appTheme === 'auto') updateEffectiveTheme(); };
    mediaQuery.addEventListener('change', handler);
    return () => { 
      mediaQuery.removeEventListener('change', handler); 
      OnFileDropOff(); 
      EventsOff("menu-open-file"); EventsOff("menu-open-recent"); EventsOff("menu-save-file"); EventsOff("menu-new-tab");
      EventsOff("set-locale"); EventsOff("set-theme");
      EventsOff("format-bold"); EventsOff("format-italic"); EventsOff("format-h1"); EventsOff("format-h2"); EventsOff("format-h3"); EventsOff("format-code");
    };
  });

  // --- Svelte 5 Runes: Effects ---
  $effect(() => {
    if ($appTheme) untrack(() => updateEffectiveTheme());
  });

  $effect(() => {
    if ($locale) untrack(() => updateNativeMenu($locale));
  });

  $effect(() => {
    if (isReady && $effectiveAppTheme) {
      const theme = getTheme($effectiveAppTheme);
      updateHighlightingCSS(theme.chromaStyle);
      debouncedUpdate(markdown, theme.chromaStyle);
    }
  });

  $effect(() => {
    if (isReady && markdown !== undefined) {
      debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
    }
  });

</script>

<svelte:window onmousemove={onMouseMove} onmouseup={stopResizing} />

<WhitelistModal 
  show={showSecurityModal} 
  type={securityType} 
  resource={securityResource} 
  theme={$effectiveAppTheme}
  onConfirm={confirmSecurityRequest}
  onCancel={() => showSecurityModal = false}
/>

<main class="flex h-screen w-full overflow-hidden flex-col select-none {$effectiveAppTheme === 'dark' ? 'bg-slate-900' : 'bg-white'}">
  {#if SHOW_HTML_TOOLBAR}
    <Toolbar onOpen={handleOpen} onSave={handleSave} onNewTab={addNewTab} />
  {/if}
  
  <TabsBar tabs={tabs} bind:activeTabIndex={activeTabIndex} onCloseTab={handleCloseTab} />

  <div class="flex flex-1 overflow-hidden relative">
    {#if !$isEditorHidden && !$isFocusMode}
    <div class="flex flex-col min-w-0 border-r relative {editorClass} print:hidden" style="width: {$splitWidth}%;">
      <div class="p-2 text-xs font-bold uppercase tracking-wider opacity-50 border-b shrink-0 {toolbarClass}">
        {$t('editor')}
      </div>
      {#if tabs[activeTabIndex]}
      <div class="flex-1 min-h-0 relative">
        <Editor
          bind:this={editorComponent}
          bind:value={tabs[activeTabIndex].content}
          theme={$effectiveAppTheme}
          placeholder={$t('placeholder')}
          onchange={onContentInput}
          onscroll={handleEditorScroll}
        />
      </div>
      {/if}
      <div 
        role="slider"
        aria-label="Resize editor and preview"
        aria-valuenow={$splitWidth}
        aria-valuemin={10}
        aria-valuemax={90}
        tabindex="0"
        onmousedown={startResizing} 
        onkeydown={(e) => {
          if (e.key === 'ArrowLeft') splitWidth.set(Math.max(10, $splitWidth - 1));
          if (e.key === 'ArrowRight') splitWidth.set(Math.min(90, $splitWidth + 1));
        }}
        class="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-blue-500/50 transition-colors z-10"
      ></div>
    </div>
    {/if}

    <div class="flex-1 min-w-0 flex flex-col relative">
      {#if isResizing} <div class="absolute inset-0 z-50"></div> {/if}
      <div class="p-2 h-10 border-b flex items-center px-4 gap-4 shrink-0 {toolbarClass} print:hidden">
        <div class="flex items-center gap-2">
            <button 
                onclick={() => isEditorHidden.update(v => !v)} 
                class="p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors {$isEditorHidden ? focusButtonClass : ''}"
                title={$isEditorHidden ? $t('showEditor') : $t('hideEditor')}
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="9" y1="3" x2="9" y2="21"></line></svg>
            </button>
            <button 
                onclick={() => isFocusMode.update(v => !v)} 
                class="p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors {$isFocusMode ? focusButtonClass : ''}"
                title={$isFocusMode ? $t('exitFocusMode') : $t('focusMode')}
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"></path></svg>
            </button>
            <button 
                onclick={() => isSyncScroll = !isSyncScroll} 
                class="p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors {isSyncScroll ? focusButtonClass : ''}"
                title={$t('syncScroll')}
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 15l5 5 5-5"/><path d="M7 9l5-5 5 5"/></svg>
            </button>
        </div>
        <div class="text-xs font-bold uppercase tracking-wider opacity-50">{$t('preview')}</div>
        <div class="flex-1"></div>
        <button onclick={handlePrint} title={$t('print')} class="p-1 rounded transition-colors {buttonClass}">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 6 2 18 2 18 9"></polyline><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"></path><rect x="6" y="14" width="12" height="8"></rect></svg>
        </button>
        <button onclick={handleExport} title={$t('export')} class="p-1 rounded transition-colors {buttonClass}">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
        </button>
        <div class="h-4 w-px {dividerClass}"></div>
        <button 
          onclick={toggleAppTheme} 
          title={$t('toggleTheme') + ' (' + ($appTheme === 'auto' ? $t('menuThemeAuto') : ($appTheme === 'dark' ? $t('menuThemeDark') : $t('menuThemeLight'))) + ')'}
          class="flex items-center gap-1.5 px-2 py-0.5 rounded text-xs transition-colors {buttonClass}"
        >
          {#if $appTheme === 'dark'}
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
            <span>{$t('menuThemeDark')}</span>
          {:else if $appTheme === 'light'}
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
            <span>{$t('menuThemeLight')}</span>
          {:else}
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
            <span>{$t('menuThemeAuto')}</span>
          {/if}
        </button>
        <div class="h-4 w-px {dividerClass}"></div>
        <div class="flex gap-1 items-center">
          <button onclick={() => adjustFontSize(-5)} class="w-5 h-5 flex items-center justify-center rounded text-xs font-bold {buttonClass}">-</button>
          <span class="text-[10px] opacity-60 w-8 text-center font-mono">{fontSize}%</span>
          <button onclick={() => adjustFontSize(5)} class="w-5 h-5 flex items-center justify-center rounded text-xs font-bold {buttonClass}">+</button>
        </div>
      </div>
      <Preview 
        bind:this={previewComponent}
        html={htmlContent} 
        css={highlightingCSS} 
        theme={currentPreviewTheme} 
        fontSize={fontSize} 
        currentFilePath={tabs[activeTabIndex]?.path}
        onsecurity_request={handleSecurityRequest}
        onopen_file={handleOpenExternalMD}
        onscroll={handlePreviewScroll}
      />
    </div>
  </div>

  <StatusBar {wordCount} {charCount} {readingTime} activeTab={tabs[activeTabIndex]} />

  {#if $dropMessage}
    <div class="drop-toast">{$dropMessage}</div>
  {/if}
</main>

<style>
  :global(body) { margin: 0; }
  .no-scrollbar::-webkit-scrollbar { display: none; }
  .no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
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
  @keyframes slideUp { from { transform: translate(-50%, 100%); opacity: 0; } to { transform: translate(-50%, 0); opacity: 1; } }

  @media print {
    :global(body), main {
      background: white !important;
      background-color: white !important;
      color: #111827 !important;
      height: auto !important;
      overflow: visible !important;
      display: block !important;
    }
  }
</style>