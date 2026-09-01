<script lang="ts">
  /**
   * Toolbar component for MD Viewer.
   * Refactored for Svelte 5 Runes.
   */
  import { t, locale } from '../i18n';
  import { STYLE, APP_THEME } from '../lib/constants';
  import { appTheme, effectiveAppTheme, isFocusMode, isPrinting, appVersion } from '../lib/stores';

  // --- Svelte 5 Runes: Props ---
  let { 
    onOpen, 
    onSave, 
    onNewTab,
    onCreateNewTab
  } = $props<{
    onOpen: () => void;
    onSave: () => void;
    onNewTab: () => void;
    onCreateNewTab?: () => void;
  }>();

  // --- Svelte 5 Runes: Derived ---
  const toolbarClass = $derived(STYLE.toolbar[$effectiveAppTheme]);
  const buttonClass = $derived(STYLE.button[$effectiveAppTheme]);

  function toggleTheme() {
    appTheme.update(current => {
      if (current === APP_THEME.DARK) return APP_THEME.LIGHT;
      if (current === APP_THEME.LIGHT) return APP_THEME.AUTO;
      return APP_THEME.DARK;
    });
  }
</script>

{#if !$isFocusMode && !$isPrinting}
<div class="h-12 border-b flex items-center px-4 gap-4 shrink-0 z-20 {toolbarClass} print:hidden">
  <span class="text-sm font-semibold opacity-70">MarkSafe v{$appVersion}</span>

  <div class="flex-1"></div>

  <div class="flex items-center gap-3">
    <button 
      onclick={toggleTheme} 
      title={$t('toggleTheme') + ' (' + ($appTheme === 'auto' ? $t('menuThemeAuto') : ($appTheme === 'dark' ? $t('menuThemeDark') : $t('menuThemeLight'))) + ')'} 
      class="p-2 rounded-full transition-colors flex items-center gap-1.5 {buttonClass}"
    >
      {#if $appTheme === 'dark'}
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
      {:else if $appTheme === 'light'}
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
      {:else}
        <div class="relative">
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
          <span class="absolute -top-1 -right-1 flex h-2 w-2"><span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span><span class="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span></span>
        </div>
      {/if}
    </button>
  </div>
</div>
{/if}
