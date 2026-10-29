<script lang="ts">
  /**
   * Form field of the MLC Markdown Extension (non-standard) as a dialog:
   * kind, name, width, preset, tooltip … instead of the syntax. Inserts a
   * new field or replaces the one being edited. Writing and reading is
   * mlcext (FormatField / FieldAt), so the syntax is the checker's.
   */
  import { onMount } from 'svelte';
  import { renderMarkdown } from 'mlc-markdown';
  import { t, locale } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as backend from '../lib/backend';

  type UiKind = 'text' | 'multiline' | 'check' | 'date' | 'time' | 'pattern' | 'signature';

  let { initial = null, usedNames = [], prefill = [], theme = 'dark', onApply, onClose }: {
    /** The field being edited; null inserts a new one. */
    initial?: backend.MLCFieldSpec | null;
    /** Names of the other fields in the document (duplicate warning). */
    usedNames?: string[];
    /** Field names the Sensor Monitor fills from an event. */
    prefill?: { name: string; label: string }[];
    theme?: EffectiveTheme_t;
    onApply: (text: string) => void;
    onClose: () => void;
  } = $props();

  const DATE = 'TT.MM.JJJJ';
  const TIME = 'HH:MM';

  // svelte-ignore state_referenced_locally
  const start = initial;
  let kind = $state<UiKind>(
    !start ? 'text'
      : start.kind === 'pattern' ? (start.pattern === DATE ? 'date' : start.pattern === TIME ? 'time' : 'pattern')
      : (start.kind as UiKind) || 'text',
  );
  let name = $state(start?.name ?? '');
  let chars = $state(start?.chars || 20);
  let rows = $state(start?.rows || 3);
  let fillPage = $state(start?.fillPage ?? false);
  let pattern = $state(start?.kind === 'pattern' ? start.pattern ?? '' : '');
  let value = $state(start?.value ?? '');
  let checked = $state(start?.checked ?? false);
  let role = $state(start?.role ?? '');
  let tooltip = $state(start?.tooltip ?? '');

  let text = $state('');
  let error = $state('');

  const KINDS: { id: UiKind; label: string }[] = [
    { id: 'text', label: 'fieldKindText' },
    { id: 'multiline', label: 'fieldKindMultiline' },
    { id: 'check', label: 'fieldKindCheck' },
    { id: 'date', label: 'fieldKindDate' },
    { id: 'time', label: 'fieldKindTime' },
    { id: 'pattern', label: 'fieldKindPattern' },
    { id: 'signature', label: 'fieldKindSignature' },
  ];

  const spec = $derived<backend.MLCFieldSpec>({
    kind: kind === 'date' || kind === 'time' ? 'pattern' : kind,
    name: name.trim(),
    chars: kind === 'text' || kind === 'multiline' ? Number(chars) || 0 : 0,
    rows: kind === 'multiline' ? Number(rows) || 1 : 0,
    fillPage: kind === 'multiline' && fillPage,
    pattern: kind === 'date' ? DATE : kind === 'time' ? TIME : kind === 'pattern' ? pattern : '',
    value: kind === 'check' || kind === 'signature' ? '' : value,
    checked: kind === 'check' && checked,
    role: kind === 'signature' ? role : '',
    tooltip,
  });

  // Written form from mlcext, again on every change.
  let seq = 0;
  $effect(() => {
    const s = spec;
    const mine = ++seq;
    void backend.mlcFormatField(s).then((r) => {
      if (mine !== seq) return;
      text = r.text;
      error = r.error;
    });
  });

  const duplicate = $derived(!!spec.name && spec.name !== (start?.name ?? '') && usedNames.includes(spec.name));
  const prefillLabel = $derived(prefill.find((p: { name: string; label: string }) => p.name === spec.name)?.label ?? '');
  const previewHtml = $derived.by(() => {
    if (!text) return '';
    try { return renderMarkdown(text, { locale: $locale }).html; } catch { return ''; }
  });

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);
  const inputClass = $derived(currentTheme === 'dark' ? 'bg-slate-800 border-slate-600' : 'bg-white border-gray-300');

  function apply() {
    if (!text || error) return;
    onApply(text);
  }

  let firstInput = $state<HTMLSelectElement>();
  onMount(() => firstInput?.focus());

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" role="dialog" aria-modal="true" aria-label={start ? $t('fieldDialogEdit') : $t('fieldDialogInsert')}>
  <div class="max-w-xl w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <form class="p-6 text-sm" onsubmit={(e) => { e.preventDefault(); apply(); }}>
      <h3 class="text-lg font-bold mb-4">{start ? $t('fieldDialogEdit') : $t('fieldDialogInsert')}</h3>

      <div class="grid grid-cols-[max-content_1fr] gap-x-3 gap-y-2 items-center">
        <label for="fd-kind" class="opacity-80">{$t('fieldKind')}</label>
        <select id="fd-kind" bind:this={firstInput} bind:value={kind} class="px-2 py-1 rounded border {inputClass}">
          {#each KINDS as k}<option value={k.id}>{$t(k.label)}</option>{/each}
        </select>

        <label for="fd-name" class="opacity-80">{$t('fieldName')}</label>
        <div>
          <input id="fd-name" type="text" bind:value={name} list="fd-prefill" spellcheck="false" autocomplete="off"
            placeholder={$t('fieldNameHint')} class="w-full px-2 py-1 rounded border font-mono select-text {inputClass}" />
          <datalist id="fd-prefill">
            {#each prefill as p}<option value={p.name}>{p.label}</option>{/each}
          </datalist>
          {#if duplicate}<p class="text-xs text-amber-500 mt-1">{$t('fieldNameDuplicate')}</p>{/if}
          {#if prefillLabel}<p class="text-xs opacity-70 mt-1">{$t('fieldNamePrefill', prefillLabel)}</p>{/if}
        </div>

        {#if kind === 'text' || kind === 'multiline'}
          <label for="fd-chars" class="opacity-80">{$t('fieldChars')}</label>
          <input id="fd-chars" type="number" min="4" max="120" bind:value={chars} class="w-24 px-2 py-1 rounded border {inputClass}" />
        {/if}
        {#if kind === 'multiline'}
          <label for="fd-rows" class="opacity-80">{$t('fieldRows')}</label>
          <div class="flex items-center gap-4">
            <input id="fd-rows" type="number" min="1" max="40" bind:value={rows} class="w-24 px-2 py-1 rounded border {inputClass}" />
            <label class="flex items-center gap-1.5"><input type="checkbox" bind:checked={fillPage} /> {$t('fieldFillPage')}</label>
          </div>
        {/if}
        {#if kind === 'pattern'}
          <label for="fd-pattern" class="opacity-80">{$t('fieldPattern')}</label>
          <input id="fd-pattern" type="text" bind:value={pattern} spellcheck="false" placeholder="AA-999" class="w-40 px-2 py-1 rounded border font-mono select-text {inputClass}" />
        {/if}
        {#if kind === 'signature'}
          <label for="fd-role" class="opacity-80">{$t('fieldRole')}</label>
          <input id="fd-role" type="text" bind:value={role} class="w-full px-2 py-1 rounded border select-text {inputClass}" />
        {/if}
        {#if kind === 'check'}
          <span></span>
          <label class="flex items-center gap-1.5"><input type="checkbox" bind:checked /> {$t('fieldChecked')}</label>
        {:else if kind !== 'signature'}
          <label for="fd-value" class="opacity-80">{$t('fieldValue')}</label>
          {#if kind === 'multiline'}
            <textarea id="fd-value" rows="2" bind:value class="w-full px-2 py-1 rounded border select-text {inputClass}"></textarea>
          {:else}
            <input id="fd-value" type="text" bind:value placeholder={kind === 'date' ? '16.10.2026' : kind === 'time' ? '14:30' : ''}
              class="w-full px-2 py-1 rounded border select-text {inputClass}" />
          {/if}
        {/if}

        <label for="fd-tooltip" class="opacity-80">{$t('fieldTooltip')}</label>
        <input id="fd-tooltip" type="text" bind:value={tooltip} class="w-full px-2 py-1 rounded border select-text {inputClass}" />
      </div>

      <div class="mt-5 text-xs uppercase tracking-wide opacity-60">{$t('fieldSyntax')}</div>
      {#if error}
        <p class="mt-1 text-red-500">{error}</p>
      {:else}
        <pre class="mt-1 p-2 rounded border font-mono text-xs whitespace-pre-wrap break-all select-text {inputClass}">{text}</pre>
        {#if previewHtml}
          <div class="mt-3 text-xs uppercase tracking-wide opacity-60">{$t('fieldPreview')}</div>
          <!-- eslint-disable-next-line svelte/no-at-html-tags -- sanitized by mlc-markdown (DOMPurify) -->
          <div class="mlc-md mt-1 p-2 rounded border {inputClass}" data-theme={currentTheme}>{@html previewHtml}</div>
        {/if}
      {/if}

      <div class="flex justify-end gap-3 mt-6">
        <button type="button" onclick={onClose} class="px-4 py-2 rounded font-medium transition-colors {buttonClass}">{$t('cancel')}</button>
        <button type="submit" disabled={!text || !!error} class="px-4 py-2 rounded font-medium bg-blue-600 hover:bg-blue-500 text-white disabled:opacity-50">
          {start ? $t('fieldApplyEdit') : $t('fieldApplyInsert')}
        </button>
      </div>
    </form>
  </div>
</div>
