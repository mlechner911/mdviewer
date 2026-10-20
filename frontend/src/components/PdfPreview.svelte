<script lang="ts">
  /**
   * PdfPreview shows the pages exactly as the PDF toolkit mlcdoc renders
   * them, for any document: form fields of the MLC Markdown Extension as
   * lines and boxes, page breaks, header and footer. Nothing is re-drawn
   * here — the images come from mlcdoc via PreviewPDF, the same path as the
   * PDF export.
   */
  import { untrack } from 'svelte';
  import { t } from '../i18n';
  import * as backend from '../lib/backend';

  let { markdown, path = null, version = '', hasMath = false } = $props<{
    markdown: string;
    path?: string | null;
    version?: string;
    hasMath?: boolean;
  }>();

  let pages = $state<string[]>([]);
  let warnings = $state<string[]>([]);
  let ignored = $state<string[]>([]);
  let error = $state('');
  let loading = $state(false);
  let seq = 0;
  let timer: ReturnType<typeof setTimeout> | undefined;

  // Re-render shortly after the last change; a stale answer is dropped.
  $effect(() => {
    const md = markdown;
    const p = path;
    clearTimeout(timer);
    timer = setTimeout(async () => {
      const mine = ++seq;
      loading = true;
      const res = await backend.previewPDF(md, p ?? '');
      if (mine !== seq) return;
      loading = false;
      if (!res) return;
      error = res.error ? (res.errorLine ? `${$t('line')} ${res.errorLine}: ` : '') + res.error : '';
      warnings = res.warnings ?? [];
      ignored = res.ignored ?? [];
      if (!res.error) pages = res.pages ?? [];
      // pages is read untracked: setting it must not schedule the next
      // render (that looped every 800 ms, "wird gesetzt…" for ever).
    }, untrack(() => pages.length) === 0 ? 0 : 800);
    return () => clearTimeout(timer);
  });
</script>

<div class="pdf-preview flex-1 overflow-auto">
  <div class="pdf-bar">
    <span>{$t('pdfPreviewTitle')}</span>
    <span class="opacity-60">{version}</span>
    {#if loading}<span class="opacity-60">· {$t('pdfPreviewLoading')}</span>{/if}
  </div>
  {#if error}
    <div class="pdf-msg pdf-error">
      <strong>{$t('pdfPreviewFailed')}</strong>
      <div class="font-mono">{error}</div>
    </div>
  {/if}
  {#if hasMath}
    <div class="pdf-msg pdf-warning">{$t('pdfMathAsText')}</div>
  {/if}
  {#if ignored.length}
    <div class="pdf-msg pdf-info">{$t('pdfIgnoredKeys', ignored.join(', '))}</div>
  {/if}
  {#each warnings as w}
    <div class="pdf-msg pdf-warning font-mono">{w}</div>
  {/each}
  <div class="pdf-pages" class:stale={!!error}>
    {#each pages as src, i}
      <img {src} alt="{$t('pdfPreviewPage')} {i + 1}" />
    {/each}
  </div>
</div>

<style>
  .pdf-preview { background: #6b7280; }
  .pdf-bar {
    position: sticky; top: 0; z-index: 1;
    display: flex; gap: 0.5rem; align-items: center;
    padding: 0.3rem 0.75rem; font-size: 11px;
    background: rgba(17, 24, 39, 0.85); color: #f3f4f6;
  }
  .pdf-msg { margin: 0.75rem auto 0; max-width: 820px; padding: 0.5rem 0.75rem; border-radius: 4px; font-size: 12px; }
  .pdf-error { background: #fef2f2; color: #991b1b; border-left: 4px solid #dc2626; }
  .pdf-warning { background: #fffbeb; color: #92400e; border-left: 4px solid #d97706; }
  .pdf-info { background: #eff6ff; color: #1e3a8a; border-left: 4px solid #3b82f6; }
  .pdf-pages { display: flex; flex-direction: column; align-items: center; gap: 1rem; padding: 1rem; }
  .pdf-pages.stale { opacity: 0.45; }
  .pdf-pages img { width: 100%; max-width: 820px; background: #fff; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35); }
  /* Print: only the pages, one per sheet; status bar and notes are screen
     aids. The page images carry their own margins, so the sheet has none
     (this also drops the browser's date/title lines). */
  @page { margin: 0; }
  @media print {
    .pdf-preview { background: none; overflow: visible; }
    .pdf-bar, .pdf-msg { display: none; }
    .pdf-pages { display: block; padding: 0; }
    .pdf-pages img { max-width: none; box-shadow: none; break-after: page; display: block; }
    .pdf-pages img:last-child { break-after: auto; }
  }
</style>
