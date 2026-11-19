<script module lang="ts">
  /** mlcext.FieldSpec (JSON). */
  export interface MlcFieldSpec {
    kind: string;
    name: string;
    chars?: number;
    rows?: number;
    fillPage?: boolean;
    pattern?: string;
    value?: string;
    checked?: boolean;
    role?: string;
    tooltip?: string;
  }
</script>

<script lang="ts">
  // Form field of the MLC Markdown Extension (non-standard) as a dialog:
  // kind, name, width, preset, tooltip … instead of the syntax. Inserts a new
  // field or replaces the one being edited. Writing the field is mlcext's
  // FormatField, called by the host through `format`, so the syntax is the
  // checker's and this dialog knows none of it.
  import { onMount } from 'svelte';
  import { renderMarkdown } from 'mlc-markdown';
  import MlcDialog from './MlcDialog.svelte';
  import { dialogTexts } from './labels.js';

  type UiKind = 'text' | 'multiline' | 'check' | 'date' | 'time' | 'pattern' | 'signature';

  let {
    initial = null,
    format,
    usedNames = [],
    prefill = [],
    properties = [],
    theme = 'light',
    locale = 'en',
    labels = {},
    preview = true,
    onApply,
    onClose,
  }: {
    /** The field being edited (mlcext.FieldAt().spec); null inserts a new one. */
    initial?: MlcFieldSpec | null;
    /** mlcext.FormatField of the host: the written field, or why there is none. */
    format: (spec: MlcFieldSpec) => Promise<{ text: string; error: string }>;
    /** Names of the other fields in the document (duplicate warning). */
    usedNames?: string[];
    /** Field names the Sensor Monitor fills from an event. */
    prefill?: { name: string; label: string }[];
    /** Properties of the document's schema, with their description. */
    properties?: { name: string; label: string }[];
    theme?: 'light' | 'dark';
    locale?: string;
    /** Texts replacing the built-in German and English ones (keys: see labels.js). */
    labels?: Record<string, string>;
    /** Show the field as it renders (mlc-markdown). */
    preview?: boolean;
    onApply: (text: string) => void;
    onClose: () => void;
  } = $props();

  const tr = $derived(dialogTexts(locale, labels));

  const DATE = 'TT.MM.JJJJ';
  const TIME = 'HH:MM';

  // svelte-ignore state_referenced_locally
  const start = initial;
  let kind = $state<UiKind>(
    !start
      ? 'text'
      : start.kind === 'pattern'
        ? start.pattern === DATE
          ? 'date'
          : start.pattern === TIME
            ? 'time'
            : 'pattern'
        : (start.kind as UiKind) || 'text',
  );
  let name = $state(start?.name ?? '');
  let chars = $state(start?.chars || 20);
  let rows = $state(start?.rows || 3);
  let fillPage = $state(start?.fillPage ?? false);
  let pattern = $state(start?.kind === 'pattern' ? (start.pattern ?? '') : '');
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

  const spec = $derived<MlcFieldSpec>({
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

  // Written form from mlcext, again on every change; a late answer to an
  // older input is dropped.
  let seq = 0;
  $effect(() => {
    const s = spec;
    const mine = ++seq;
    format(s).then(
      (r) => {
        if (mine !== seq) return;
        text = r.text;
        error = r.error;
      },
      (err) => {
        if (mine !== seq) return;
        text = '';
        error = String(err?.message ?? err);
      },
    );
  });

  const duplicate = $derived(!!spec.name && spec.name !== (start?.name ?? '') && usedNames.includes(spec.name));
  const prefillLabel = $derived(prefill.find((p) => p.name === spec.name)?.label ?? '');
  const propertyLabel = $derived(properties.find((p) => p.name === spec.name)?.label ?? '');
  const previewHtml = $derived.by(() => {
    if (!preview || !text) return '';
    try {
      return renderMarkdown(text, { locale }).html;
    } catch {
      return '';
    }
  });

  function apply() {
    if (text && !error) onApply(text);
  }

  let firstInput = $state<HTMLSelectElement>();
  onMount(() => firstInput?.focus());
</script>

<MlcDialog
  title={start ? tr('fieldDialogEdit') : tr('fieldDialogInsert')}
  {theme}
  width="36rem"
  submitLabel={start ? tr('fieldApplyEdit') : tr('fieldApplyInsert')}
  cancelLabel={tr('cancel')}
  submitDisabled={!text || !!error}
  onSubmit={apply}
  {onClose}
>
  <div class="mlc-d-grid">
    <label for="mlc-fd-kind" class="mlc-d-label">{tr('fieldKind')}</label>
    <select id="mlc-fd-kind" bind:this={firstInput} bind:value={kind} data-testid="field-kind">
      {#each KINDS as k (k.id)}<option value={k.id}>{tr(k.label)}</option>{/each}
    </select>

    <label for="mlc-fd-name" class="mlc-d-label">{tr('fieldName')}</label>
    <div>
      <input
        id="mlc-fd-name"
        type="text"
        bind:value={name}
        spellcheck="false"
        autocomplete="off"
        placeholder={tr('fieldNameHint')}
        list="mlc-fd-names"
        class="mlc-d-mono"
        style="width: 100%"
      />
      <datalist id="mlc-fd-names">
        {#each properties as p (p.name)}<option value={p.name}>{p.label}</option>{/each}
        {#each prefill as p (p.name)}<option value={p.name}>{p.label}</option>{/each}
      </datalist>
      {#if duplicate}<p class="mlc-d-warn">{tr('fieldNameDuplicate')}</p>{/if}
      {#if propertyLabel}<p class="mlc-d-note">{propertyLabel}</p>{/if}
      {#if prefillLabel}<p class="mlc-d-note">{tr('fieldNamePrefill', prefillLabel)}</p>{/if}
    </div>

    {#if kind === 'text' || kind === 'multiline'}
      <label for="mlc-fd-chars" class="mlc-d-label">{tr('fieldChars')}</label>
      <input id="mlc-fd-chars" type="number" min="4" max="120" bind:value={chars} style="width: 6rem" />
    {/if}
    {#if kind === 'multiline'}
      <label for="mlc-fd-rows" class="mlc-d-label">{tr('fieldRows')}</label>
      <div class="mlc-d-row" style="gap: 1rem">
        <input id="mlc-fd-rows" type="number" min="1" max="40" bind:value={rows} style="width: 6rem" />
        <label class="mlc-d-check"><input type="checkbox" bind:checked={fillPage} /> {tr('fieldFillPage')}</label>
      </div>
    {/if}
    {#if kind === 'pattern'}
      <label for="mlc-fd-pattern" class="mlc-d-label">{tr('fieldPattern')}</label>
      <input id="mlc-fd-pattern" type="text" bind:value={pattern} spellcheck="false" placeholder="AA-999" class="mlc-d-mono" style="width: 10rem" />
    {/if}
    {#if kind === 'signature'}
      <label for="mlc-fd-role" class="mlc-d-label">{tr('fieldRole')}</label>
      <input id="mlc-fd-role" type="text" bind:value={role} style="width: 100%" />
    {/if}
    {#if kind === 'check'}
      <span></span>
      <label class="mlc-d-check"><input type="checkbox" bind:checked /> {tr('fieldChecked')}</label>
    {:else if kind !== 'signature'}
      <label for="mlc-fd-value" class="mlc-d-label">{tr('fieldValue')}</label>
      {#if kind === 'multiline'}
        <textarea id="mlc-fd-value" rows="2" bind:value style="width: 100%"></textarea>
      {:else}
        <input id="mlc-fd-value" type="text" bind:value placeholder={kind === 'date' ? '16.10.2026' : kind === 'time' ? '14:30' : ''} style="width: 100%" />
      {/if}
    {/if}

    <label for="mlc-fd-tooltip" class="mlc-d-label">{tr('fieldTooltip')}</label>
    <input id="mlc-fd-tooltip" type="text" bind:value={tooltip} style="width: 100%" />
  </div>

  <div class="mlc-d-heading">{tr('fieldSyntax')}</div>
  {#if error}
    <p class="mlc-d-error" style="font-size: inherit" data-testid="field-error">{error}</p>
  {:else}
    <pre class="mlc-d-box mlc-d-mono" style="margin: 0.25rem 0 0; padding: 0.5rem; font-size: 0.75rem; line-height: 1rem; white-space: pre-wrap; word-break: break-all" data-testid="field-syntax">{text}</pre>
    {#if previewHtml}
      <div class="mlc-d-heading" style="margin-top: 0.75rem">{tr('fieldPreview')}</div>
      <!-- sanitized by mlc-markdown (DOMPurify) -->
      <div class="mlc-md mlc-d-box" style="margin-top: 0.25rem; padding: 0.5rem" data-theme={theme}>{@html previewHtml}</div>
    {/if}
  {/if}
</MlcDialog>
