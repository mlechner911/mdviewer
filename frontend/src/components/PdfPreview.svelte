<script lang="ts">
  /**
   * PdfPreview shows the pages exactly as the PDF toolkit mlcdoc renders
   * them (MLC Markdown Extension, non-standard): form fields as lines and
   * boxes, page breaks, header and footer. Nothing is re-drawn here — the
   * images come from mlcdoc via PreviewMLCDoc.
   */
  import { t } from '../i18n';
  import * as backend from '../lib/backend';

  let { markdown, path = null, version = '' } = $props<{
    markdown: string;
    path?: string | null;
    version?: string;
  }>();

  let pages = $state<string[]>([]);
  let warnings = $state<string[]>([]);
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
      const res = await backend.previewMLCDoc(md, p ?? '');
      if (mine !== seq) return;
      loading = false;
      if (!res) return;
      error = res.error;
      warnings = res.warnings ?? [];
      if (!res.error) pages = res.pages ?? [];
    }, pages.length === 0 ? 0 : 800);
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
  .pdf-pages { display: flex; flex-direction: column; align-items: center; gap: 1rem; padding: 1rem; }
  .pdf-pages.stale { opacity: 0.45; }
  .pdf-pages img { width: 100%; max-width: 820px; background: #fff; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35); }
</style>
