<script lang="ts">
  /**
   * StatusBar component for MD Viewer.
   * Refactored for Svelte 5 Runes.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import { effectiveAppTheme, isPrinting } from '../lib/stores';
  import type { MLCDiagnostic } from '../lib/backend';

  // --- Svelte 5 Runes: Props ---
  let { 
    wordCount, 
    charCount, 
    readingTime, 
    activeTab,
    diagnostics = [],
    sensor = null
  } = $props<{
    wordCount: number;
    charCount: number;
    readingTime: number;
    activeTab: any;
    diagnostics?: MLCDiagnostic[];
    /** Sensor Monitor connection; shown only while signed in. */
    sensor?: { signedIn: boolean; email: string; expires: string } | null;
  }>();

  const sensorUntil = $derived(
    sensor?.signedIn && sensor.expires
      ? new Date(sensor.expires).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      : '',
  );

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

{#if !$isPrinting}
<div class="h-6 border-t flex items-center px-4 gap-6 shrink-0 text-[10px] font-medium {statusClass} py-3 print:hidden">
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
  {#if sensor?.signedIn}
    <div class="text-green-500" title={sensor.email}>{sensorUntil ? $t('smStatus', sensorUntil) : $t('smSignedIn')}</div>
    <div class="h-3 w-px {dividerClass}"></div>
  {/if}
  <div>{$t('readingTime', readingTime)}</div>
  <div class="h-3 w-px {dividerClass}"></div>
  <div class="uppercase tracking-tighter opacity-80">{activeTab?.path || $t('untitled')}</div>
</div>
{/if}
