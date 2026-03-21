<script lang="ts">
  import { STYLE } from '../lib/constants';
  import { effectiveAppTheme, isFocusMode, isPrinting } from '../lib/stores';

  export let tabs: any[];
  export let activeTabIndex: number;
  export let onCloseTab: (index: number, event?: MouseEvent) => void;

  $: toolbarClass = STYLE.toolbar[$effectiveAppTheme];
  $: dividerClass = STYLE.divider[$effectiveAppTheme];
  $: activeTabClass = STYLE.tab.active[$effectiveAppTheme];
  $: inactiveTabClass = STYLE.tab.inactive[$effectiveAppTheme];
</script>

{#if !$isFocusMode && !$isPrinting}
<div class="flex overflow-x-auto no-scrollbar border-b {toolbarClass} print:hidden">
  {#each tabs as tab, i}
    <button 
      type="button"
      class="flex items-center px-4 h-9 cursor-pointer transition-colors border-r text-xs font-medium min-w-[120px] max-w-[200px] {i === activeTabIndex ? activeTabClass : inactiveTabClass} {dividerClass}"
      on:click={() => activeTabIndex = i}
      on:keydown={(e) => e.key === 'Enter' && (activeTabIndex = i)}
    >
      <span class="truncate flex-1 text-left">{tab.title}{tab.isDirty ? ' *' : ''}</span>
      <button 
        type="button"
        on:click={(e) => onCloseTab(i, e)}
        class="ml-2 p-0.5 rounded-full hover:bg-black/10 dark:hover:bg-white/10"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
      </button>
    </button>
  {/each}
</div>
{/if}

<style>
  .no-scrollbar::-webkit-scrollbar { display: none; }
  .no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
</style>
