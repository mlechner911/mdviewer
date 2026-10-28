<script lang="ts">
  /**
   * StatusBar component for MD Viewer.
   * Refactored for Svelte 5 Runes.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import { effectiveAppTheme } from '../lib/stores';
  import type { MLCDiagnostic } from '../lib/backend';

  // --- Svelte 5 Runes: Props ---
  let { 
    wordCount, 
    charCount, 
    readingTime, 
    activeTab,
    diagnostics = [],
    extraItems = []
  } = $props<{
    wordCount: number;
    charCount: number;
    readingTime: number;
    activeTab: any;
    diagnostics?: MLCDiagnostic[];
    /** Entries of built-in modules (e.g. the Sensor Monitor's comparison). */
    extraItems?: { text: string; tone?: 'ok' | 'warn'; title?: string }[];
  }>();

  // --- Svelte 5 Runes: Derived ---
  const statusClass = $derived(STYLE.status[$effectiveAppTheme]);
  const dividerClass = $derived(STYLE.divider[$effectiveAppTheme]);

  // MLC Markdown Extension (non-standard): only documents that use it get
  // diagnostics; the tooltip lists them with their line.
  const mlcErrors = $derived(diagnostics.filter((d: MLCDiagnostic) => d.severity === 'error').length);
  const mlcWarnings = $derived(diagnostics.filter((d: MLCDiagnostic) => d.severity === 'warning').length);
  const mlcTitle = $derived(
    diagnostics.map((d: MLCDiagnostic) => `${$t('line')} ${d.line}: ${d.message}`).join('\n')
  );
</script>

<div class="h-6 border-t flex items-center px-4 gap-6 shrink-0 text-[10px] font-medium {statusClass} py-3">
  <div class="flex gap-4">
      <span>{wordCount} {$t('words')}</span>
      <span>{charCount} {$t('characters')}</span>
  </div>
  <div class="flex-1"></div>
  {#if diagnostics.length > 0}
    <div class="flex gap-2 items-center" title={mlcTitle}>
      <span class="opacity-70">{$t('mlcExtension')}</span>
      {#if mlcErrors > 0}<span class="text-red-500 font-semibold">{$t('mlcErrors', mlcErrors)}</span>{/if}
      {#if mlcWarnings > 0}<span class="text-amber-500 font-semibold">{$t('mlcWarnings', mlcWarnings)}</span>{/if}
      {#if mlcErrors === 0 && mlcWarnings === 0}<span class="text-green-500">{$t('mlcOk')}</span>{/if}
    </div>
    <div class="h-3 w-px {dividerClass}"></div>
  {/if}
  {#each extraItems as item}
    <div class={item.tone === 'ok' ? 'text-green-500' : item.tone === 'warn' ? 'text-amber-500' : ''} title={item.title}>{item.text}</div>
    <div class="h-3 w-px {dividerClass}"></div>
  {/each}
  <div>{$t('readingTime', readingTime)}</div>
  <div class="h-3 w-px {dividerClass}"></div>
  <div class="uppercase tracking-tighter opacity-80">{activeTab?.path || $t('untitled')}</div>
</div>

