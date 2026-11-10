<script lang="ts">
  /**
   * Insert an image: a file (chosen or dropped) or a picture from the
   * clipboard, as a reference or embedded in the document. A reference works
   * only inside the document's folder: the PDF reads nothing outside it.
   * With existing, an image in the document is edited: without another file
   * only its alt text (text and size) changes.
   */
  import { onMount } from 'svelte';
  import { t, locale } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as backend from '../lib/backend';
  import { formatBytes } from '../lib/editor/dataUri';
  import { formatImageSize, imageAlt, parseImageSize, splitImageAlt } from 'mlc-markdown';

  const LARGE = 5 * 1024 * 1024;

  let { path = null, dataUri = null, existing = null, docPath = null, preferEmbed = false, theme = 'dark', onApply, onClose }: {
    /** A dropped file; null starts with the file choice. */
    path?: string | null;
    /** A picture from the clipboard. */
    dataUri?: string | null;
    /** An image of the document, ![alt](target): its text and where the alt text lies in it. */
    existing?: { text: string; alt: string; target: string; altFrom: number; altTo: number } | null;
    /** The document's file: references start from its folder. */
    docPath?: string | null;
    /** A form being filled in: the record should stay complete on its own. */
    preferEmbed?: boolean;
    theme?: EffectiveTheme_t;
    onApply: (text: string) => void;
    onClose: () => void;
  } = $props();

  const toolbarClass = $derived(STYLE.toolbar[theme]);
  const buttonClass = $derived(STYLE.button[theme]);
  const inputClass = $derived(theme === 'dark' ? 'bg-slate-800 border-slate-600' : 'bg-white border-gray-300');

  let choice = $state<backend.MLCImageChoice | null>(null);
  let alt = $state('');
  /** Size after "|" in the alt text: "200", "200x100" or "50%"; empty = own size. */
  let sizeText = $state('');
  let embed = $state(true);
  let error = $state('');
  let firstInput = $state<HTMLInputElement>();

  const pasted = $derived(!!dataUri);
  /** Editing an image and keeping its file. */
  const keep = $derived(!!existing && !pasted && !choice);
  const embedded = $derived(!!existing?.target.startsWith('data:'));
  const canReference = $derived(!!choice?.reference);
  const size = $derived(pasted ? Math.floor(((dataUri ?? '').length * 3) / 4) : keep && embedded ? Math.floor(((existing?.target ?? '').length * 3) / 4) : (choice?.size ?? 0));
  const parsedSize = $derived(sizeText.trim() ? parseImageSize(sizeText) : {});
  const ready = $derived((pasted || !!choice || keep) && parsedSize !== null);

  async function use(p: string) {
    error = '';
    try {
      const c = await backend.inspectImage(p, docPath ?? '');
      if (!c) return;
      choice = c;
      embed = c.embeddable && (preferEmbed || !c.reference);
      if (!alt) alt = c.name.replace(/\.[^.]+$/, '');
    } catch (err) {
      choice = null;
      error = String(err).replace(/^\w*Error:\s*/, '');
    }
  }
  async function pick() {
    try {
      const p = await backend.chooseImage(docPath ?? '');
      if (p) await use(p);
    } catch (err) {
      error = String(err).replace(/^\w*Error:\s*/, '');
    }
  }
  async function apply() {
    if (!ready) return;
    const newAlt = imageAlt(alt.trim(), parsedSize ?? {});
    if (keep && existing) {
      onApply(existing.text.slice(0, existing.altFrom) + newAlt + existing.text.slice(existing.altTo));
      return;
    }
    try {
      const text = await backend.imageMarkdown({ path: choice?.path ?? '', dataUri: dataUri ?? '', alt: newAlt, embed: pasted || embed }, docPath ?? '');
      onApply(text);
    } catch (err) {
      error = String(err).replace(/^\w*Error:\s*/, '');
    }
  }

  onMount(() => {
    if (existing) {
      const a = splitImageAlt(existing.alt);
      alt = a.text;
      sizeText = a.error ? existing.alt.slice(existing.alt.lastIndexOf('|') + 1).trim() : formatImageSize(a.size);
    } else if (path) void use(path);
    else if (!dataUri) void pick();
    firstInput?.focus();
  });
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
  class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
  role="dialog"
  aria-modal="true"
  aria-label={existing ? $t('imageEditTitle') : $t('imageInsert')}
  tabindex="-1"
  onkeydown={(e) => { if (e.key === 'Escape') onClose(); }}
>
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <form class="p-6 text-sm" onsubmit={(e) => { e.preventDefault(); void apply(); }}>
      <h3 class="text-lg font-bold mb-4">{existing ? $t('imageEditTitle') : $t('imageInsert')}</h3>
      <div class="grid grid-cols-[max-content_1fr] gap-x-3 gap-y-2 items-center">
        <span class="opacity-80">{$t('imageFile')}</span>
        <div class="flex items-center gap-2 min-w-0">
          {#if pasted}
            <img src={dataUri} alt="" class="max-h-16 max-w-32 rounded border border-black/20" />
            <span class="opacity-80">{$t('imagePasted')}</span>
          {:else if keep && existing}
            {#if embedded}
              <img src={existing.target} alt="" class="max-h-16 max-w-32 rounded border border-black/20" />
              <span class="opacity-80">{$t('imageEmbedded')}</span>
            {:else}
              <code class="truncate opacity-80 text-xs" title={existing.target}>{existing.target}</code>
            {/if}
            <button type="button" onclick={pick} class="px-3 py-1 rounded shrink-0 {buttonClass}">{$t('imageChooseOther')}</button>
          {:else}
            <button type="button" onclick={pick} class="px-3 py-1 rounded shrink-0 {buttonClass}">{$t('imageChoose')}</button>
            <span class="truncate opacity-80" title={choice?.path ?? ''}>{choice?.name ?? ''}</span>
          {/if}
          {#if size}<span class="opacity-60 shrink-0">{formatBytes(size, $locale)}</span>{/if}
        </div>

        <label for="img-alt" class="opacity-80">{$t('imageAlt')}</label>
        <input id="img-alt" bind:this={firstInput} bind:value={alt} autocomplete="off" class="px-2 py-1 rounded border {inputClass}" />

        <label for="img-size" class="opacity-80">{$t('imageSize')}</label>
        <div class="flex items-center gap-2">
          <input id="img-size" bind:value={sizeText} autocomplete="off" placeholder="200 · 200x100 · 50%" aria-invalid={parsedSize === null} class="w-36 px-2 py-1 rounded border {inputClass}" class:border-red-500={parsedSize === null} />
          <span class="text-xs opacity-60">{parsedSize === null ? $t('imageSizeInvalid') : $t('imageSizeHint')}</span>
        </div>

        {#if !pasted && choice}
          <span class="opacity-80">{$t('imageHow')}</span>
          <div class="flex flex-col gap-1">
            <label class="flex items-center gap-2" class:opacity-50={!canReference}>
              <input type="radio" name="img-how" checked={!embed} disabled={!canReference} onchange={() => (embed = false)} />
              {$t('imageReference')}{#if choice.reference}<code class="opacity-70 text-xs truncate">{choice.reference}</code>{/if}
            </label>
            <label class="flex items-center gap-2" class:opacity-50={!choice.embeddable} title={choice.embeddable ? '' : $t('imageWebpHint')}>
              <input type="radio" name="img-how" checked={embed} disabled={!choice.embeddable} onchange={() => (embed = true)} />
              {$t('imageEmbed')}
            </label>
          </div>
        {/if}
      </div>
      {#if !pasted && choice && !canReference}<p class="mt-3 text-xs opacity-70">{$t('imageEmbedOnly')}</p>{/if}
      {#if size > LARGE && (pasted || embed)}<p class="mt-3 text-xs text-amber-500">{$t('imageLarge', formatBytes(size, $locale))}</p>{/if}
      {#if error}<p class="mt-3 text-xs text-red-500" role="alert">{error}</p>{/if}
      <div class="flex justify-end gap-2 mt-5">
        <button type="button" onclick={onClose} class="px-4 py-2 rounded font-medium transition-colors {buttonClass}">{$t('cancel')}</button>
        <button type="submit" disabled={!ready} class="px-4 py-2 rounded font-medium bg-blue-600 hover:bg-blue-500 text-white disabled:opacity-50">{existing ? $t('fieldApplyEdit') : $t('fieldApplyInsert')}</button>
      </div>
    </form>
  </div>
</div>
