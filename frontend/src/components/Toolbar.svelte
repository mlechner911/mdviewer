<script lang="ts">
  import { t, locale } from '../i18n';
  import { STYLE, APP_THEME } from '../lib/constants';
  import { appTheme, effectiveAppTheme, isFocusMode, isPrinting } from '../lib/stores';

  export let onOpen: () => void;
  export let onSave: () => void;
  export let onNewTab: () => void;

  $: toolbarClass = STYLE.toolbar[$effectiveAppTheme];
  $: buttonClass = STYLE.button[$effectiveAppTheme];

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
  <div class="flex gap-2">
    <button on:click={onOpen} title={$t('open')} class="p-2 rounded transition-colors {buttonClass}">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
    </button>
    <button on:click={onSave} title={$t('save')} class="p-2 rounded transition-colors {buttonClass}">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
    </button>
    <button on:click={onNewTab} title={$t('newTab')} class="p-2 rounded transition-colors {buttonClass}">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
    </button>
  </div>

  <div class="flex-1"></div>

  <div class="flex items-center gap-4">
      <div class="flex items-center">
        <select bind:value={$locale} class="text-[10px] font-bold uppercase rounded border-none py-1 px-2 cursor-pointer bg-transparent focus:ring-1 focus:ring-blue-500 {$effectiveAppTheme === 'dark' ? 'text-slate-100' : 'text-slate-900'}">
          <option value="en">EN</option>
          <option value="de">DE</option>
          <option value="es">ES</option>
          <option value="fr">FR</option>
        </select>
      </div>

      <div class="h-6 w-px {STYLE.divider[$effectiveAppTheme]}"></div>

      <button on:click={toggleTheme} title={$t('toggleTheme')} class="p-1.5 rounded-full transition-colors {buttonClass}">
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
      <span class="text-xs opacity-40 font-mono hidden sm:inline">MarkSafe v1.0.3</span>
  </div>
</div>
{/if}
