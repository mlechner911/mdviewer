<script lang="ts">
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import { effectiveAppTheme, isPrinting } from '../lib/stores';

  export let wordCount: number;
  export let charCount: number;
  export let readingTime: number;
  export let activeTab: any;

  $: statusClass = STYLE.status[$effectiveAppTheme];
  $: dividerClass = STYLE.divider[$effectiveAppTheme];
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
