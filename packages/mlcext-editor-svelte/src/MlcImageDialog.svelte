<script module lang="ts">
  import type { MlcImageChoice } from './imageFile.js';

  export interface MlcImageRequest {
    choice: MlcImageChoice | null;
    dataUri: string | null;
    /** The alt text with the size ("Logo|200"). */
    alt: string;
    embed: boolean;
  }
</script>

<script lang="ts">
  // Insert an image: a file (chosen or dropped) or a picture from the
  // clipboard, as a reference or embedded in the document. With `existing`,
  // an image of the document is edited: without another file only its alt
  // text (description and size) changes. Files come from the host
  // (`choose`, `chosen`), so the dialog runs on a desktop and in a browser.
  import { onMount } from 'svelte';
  import { formatImageSize, imageAlt, parseImageSize, splitImageAlt } from 'mlc-markdown';
  import MlcDialog from './MlcDialog.svelte';
  import { dialogTexts, formatBytes } from './labels.js';
  import { chooseImageFile, imageMarkdown } from './imageFile.js';

  const LARGE = 5 * 1024 * 1024;

  let {
    chosen = null,
    dataUri = null,
    existing = null,
    preferEmbed = false,
    choose = () => chooseImageFile(),
    toMarkdown,
    theme = 'light',
    locale = 'en',
    labels = {},
    onApply,
    onClose,
  }: {
    /** A file the user already gave (dropped); a promise while the host inspects it. */
    chosen?: MlcImageChoice | Promise<MlcImageChoice | null> | null;
    /** A picture from the clipboard. */
    dataUri?: string | null;
    /** An image of the document, ![alt](target): its text and where the alt text lies in it. */
    existing?: { text: string; alt: string; target: string; altFrom: number; altTo: number } | null;
    /** Start with "embedded" even when a reference works (a form being filled in). */
    preferEmbed?: boolean;
    /** The file choice of the host; default the browser's, embedded as data: URL. */
    choose?: () => Promise<MlcImageChoice | null>;
    /** The Markdown for the image; default ![alt](data: URL or reference). */
    toMarkdown?: (request: MlcImageRequest) => Promise<string>;
    theme?: 'light' | 'dark';
    locale?: string;
    labels?: Record<string, string>;
    onApply: (text: string) => void;
    onClose: () => void;
  } = $props();

  const tr = $derived(dialogTexts(locale, labels));

  let choice = $state<MlcImageChoice | null>(null);
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
  const size = $derived(
    pasted
      ? Math.floor(((dataUri ?? '').length * 3) / 4)
      : keep && embedded
        ? Math.floor(((existing?.target ?? '').length * 3) / 4)
        : (choice?.size ?? 0),
  );
  const parsedSize = $derived(sizeText.trim() ? parseImageSize(sizeText) : {});
  const ready = $derived((pasted || !!choice || keep) && parsedSize !== null);

  const message = (err: unknown) => String((err as Error)?.message ?? err).replace(/^\w*Error:\s*/, '');

  async function use(c: MlcImageChoice | Promise<MlcImageChoice | null> | null) {
    error = '';
    try {
      const got = await c;
      if (!got) return;
      choice = got;
      embed = got.embeddable && (preferEmbed || !got.reference);
      if (!alt) alt = got.name.replace(/\.[^.]+$/, '');
    } catch (err) {
      choice = null;
      error = message(err);
    }
  }

  async function pick() {
    await use(choose());
  }

  function defaultMarkdown(r: MlcImageRequest): string {
    if (r.dataUri) return imageMarkdown(r.alt, r.dataUri);
    const c = r.choice;
    if (r.embed && c?.dataUri) return imageMarkdown(r.alt, c.dataUri);
    if (!r.embed && c?.reference) return imageMarkdown(r.alt, c.reference);
    throw new Error('no image data: pass toMarkdown');
  }

  async function apply() {
    if (!ready) return;
    const newAlt = imageAlt(alt.trim(), parsedSize ?? {});
    if (keep && existing) {
      onApply(existing.text.slice(0, existing.altFrom) + newAlt + existing.text.slice(existing.altTo));
      return;
    }
    const request: MlcImageRequest = { choice, dataUri, alt: newAlt, embed: pasted || embed };
    try {
      onApply(toMarkdown ? await toMarkdown(request) : defaultMarkdown(request));
    } catch (err) {
      error = message(err);
    }
  }

  onMount(() => {
    if (existing) {
      const a = splitImageAlt(existing.alt);
      alt = a.text;
      sizeText = a.error ? existing.alt.slice(existing.alt.lastIndexOf('|') + 1).trim() : formatImageSize(a.size);
    } else if (chosen) void use(chosen);
    else if (!dataUri) void pick();
    firstInput?.focus();
  });
</script>

<MlcDialog
  title={existing ? tr('imageEditTitle') : tr('imageInsert')}
  {theme}
  submitLabel={existing ? tr('fieldApplyEdit') : tr('fieldApplyInsert')}
  cancelLabel={tr('cancel')}
  submitDisabled={!ready}
  onSubmit={() => void apply()}
  {onClose}
>
  <div class="mlc-d-grid">
    <span class="mlc-d-label">{tr('imageFile')}</span>
    <div class="mlc-d-row">
      {#if pasted}
        <img src={dataUri} alt="" class="mlc-d-thumb" />
        <span class="mlc-d-label">{tr('imagePasted')}</span>
      {:else if keep && existing}
        {#if embedded}
          <img src={existing.target} alt="" class="mlc-d-thumb" />
          <span class="mlc-d-label">{tr('imageEmbedded')}</span>
        {:else}
          <code class="mlc-d-truncate mlc-d-label" style="font-size: 0.75rem" title={existing.target}>{existing.target}</code>
        {/if}
        <button type="button" class="mlc-d-button mlc-d-small" onclick={pick}>{tr('imageChooseOther')}</button>
      {:else}
        <button type="button" class="mlc-d-button mlc-d-small" onclick={pick}>{tr('imageChoose')}</button>
        <span class="mlc-d-truncate mlc-d-label" title={choice?.path ?? ''} data-testid="image-name">{choice?.name ?? ''}</span>
      {/if}
      {#if size}<span style="opacity: 0.6; flex-shrink: 0">{formatBytes(size, locale)}</span>{/if}
    </div>

    <label for="mlc-img-alt" class="mlc-d-label">{tr('imageAlt')}</label>
    <input id="mlc-img-alt" type="text" bind:this={firstInput} bind:value={alt} autocomplete="off" />

    <label for="mlc-img-size" class="mlc-d-label">{tr('imageSize')}</label>
    <div class="mlc-d-row">
      <input
        id="mlc-img-size"
        type="text"
        bind:value={sizeText}
        autocomplete="off"
        placeholder="200 · 200x100 · 50%"
        aria-invalid={parsedSize === null}
        style="width: 9rem"
      />
      <span style="font-size: 0.75rem; opacity: 0.6">{parsedSize === null ? tr('imageSizeInvalid') : tr('imageSizeHint')}</span>
    </div>

    {#if !pasted && choice}
      <span class="mlc-d-label">{tr('imageHow')}</span>
      <div style="display: flex; flex-direction: column; gap: 0.25rem">
        <label class="mlc-d-row" style:opacity={canReference ? 1 : 0.5}>
          <input type="radio" name="mlc-img-how" checked={!embed} disabled={!canReference} onchange={() => (embed = false)} />
          {tr('imageReference')}{#if choice.reference}<code class="mlc-d-truncate" style="font-size: 0.75rem; opacity: 0.7">{choice.reference}</code>{/if}
        </label>
        <label class="mlc-d-row" style:opacity={choice.embeddable ? 1 : 0.5}>
          <input type="radio" name="mlc-img-how" checked={embed} disabled={!choice.embeddable} onchange={() => (embed = true)} />
          {tr('imageEmbed')}
        </label>
      </div>
    {/if}
  </div>
  {#if !pasted && choice && !canReference}<p class="mlc-d-note" style="margin-top: 0.75rem">{tr('imageEmbedOnly')}</p>{/if}
  {#if size > LARGE && (pasted || embed)}<p class="mlc-d-warn" style="margin-top: 0.75rem">{tr('imageLarge', formatBytes(size, locale))}</p>{/if}
  {#if error}<p class="mlc-d-error" style="margin-top: 0.75rem" role="alert">{error}</p>{/if}
</MlcDialog>

<style>
  .mlc-d-thumb {
    max-height: 4rem;
    max-width: 8rem;
    border: 1px solid rgba(0, 0, 0, 0.2);
    border-radius: 0.25rem;
  }
  .mlc-d-truncate {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
  }
</style>
