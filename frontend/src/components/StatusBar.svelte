<script lang="ts">
  /**
   * StatusBar component for MD Viewer.
   * Refactored for Svelte 5 Runes.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import { effectiveAppTheme, isPrinting } from '../lib/stores';

  // --- Svelte 5 Runes: Props ---
  let { 
    wordCount, 
    charCount, 
    readingTime, 
    activeTab 
  } = $props<{
    wordCount: number;
    charCount: number;
    readingTime: number;
    activeTab: any;
  }>();

  // --- Svelte 5 Runes: Derived ---
  const statusClass = $derived(STYLE.status[$effectiveAppTheme]);
  const dividerClass = $derived(STYLE.divider[$effectiveAppTheme]);
</script>

{#if !$isPrinting}
<div class="h-6 border-t flex items-center px-4 gap-6 shrink-0 text-[10px] font-medium {statusClass} print:hidden">
  <div class="flex gap-4">
      <span>{wordCount} {$t('words')}</span>
      <span>{charCount} {$t('characters')}</span>
  </div>
  <div class="flex-1"></div>
  <div>{$t('readingTime', readingTime)}</div>
  <div class="h-3 w-px {dividerClass}"></div>
  <div class="uppercase tracking-tighter opacity-80">{activeTab?.path || $t('untitled')}</div>
</div>
{/if}
