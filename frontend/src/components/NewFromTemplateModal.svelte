<script lang="ts">
  /**
   * NewFromTemplateModal creates a new document from a template: one of
   * the MLC Markdown Extension (non-standard) or a general one (README,
   * CHANGELOG …). The form is built from the template's schema (mlcext:
   * *.json + Markdown body) in the user's template folder, so templates
   * change without a new program. Values are checked by mlcext; a document
   * with errors is not created.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as backend from '../lib/backend';
  import type { MLCTemplate, MLCTemplateField, MLCDiagnostic } from '../lib/backend';

  let { show = false, theme = 'dark', docPath = null, onCreate, onCancel } = $props<{
    show?: boolean;
    theme?: EffectiveTheme_t;
    docPath?: string | null;
    onCreate: (content: string, fileName: string) => void;
    onCancel: () => void;
  }>();

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);
  const editorClass = $derived(STYLE.editor[currentTheme]);

  let templates = $state<MLCTemplate[]>([]);
  let selectedId = $state('');
  let values = $state<Record<string, any>>({});
  let edited = $state<Record<string, boolean>>({});
  let fieldErrors = $state<Record<string, string>>({});
  let diagnostics = $state<MLCDiagnostic[]>([]);

  const selected = $derived(templates.find((x) => x.id === selectedId));
  const groups = $derived([
    { label: $t('tplGroupMlc'), items: templates.filter((x) => x.category === 'mlc') },
    { label: $t('tplGroupGeneral'), items: templates.filter((x) => x.category !== 'mlc') },
  ].filter((g) => g.items.length));
  let folder = $state('');
  // Templates of the document folder (_vorlagen) are marked; the rest come
  // from the user's template folder.
  const fromDocFolder = (tpl: MLCTemplate) => /[\\/]_vorlagen$/.test(tpl.source ?? '');
  const fields = $derived<MLCTemplateField[]>(selected?.fields ?? []);

  // Reload on every opening: the document folder's templates may change.
  $effect(() => {
    if (!show) return;
    backend.templatesFolder().then((f) => (folder = f));
    backend.mlcTemplates().then((list) => {
      templates = list;
      if (!list.some((x) => x.id === selectedId)) selectedId = list.find((x) => !x.problems?.length)?.id ?? '';
    });
  });

  // Fresh defaults whenever the template changes.
  $effect(() => {
    const fs = fields;
    const v: Record<string, any> = {};
    for (const f of fs) {
      v[f.id] = isMulti(f) ? (f.defaultValue ? f.defaultValue.split(',') : []) : f.defaultValue ?? '';
    }
    values = v;
    edited = {};
    fieldErrors = {};
    diagnostics = [];
  });

  // Document IDs: propose the next free number until the user edits it.
  $effect(() => {
    if (!show) return;
    for (const f of fields) {
      if (f.type !== 'refno' || edited[f.id]) continue;
      const area = f.area ? String(values[f.area] ?? '') : '';
      backend.suggestMLCRefNo(f.prefix ?? '', area, docPath ?? '').then((s) => {
        if (!edited[f.id] && s) values[f.id] = s;
      });
    }
  });

  const isMulti = (f: MLCTemplateField) => f.type === 'multiselect' || (f.type === 'docref' && !!f.multiple);
  const optionLabel = (f: MLCTemplateField, i: number) => f.optionLabels?.[i] || f.optionList?.[i] || '';

  function toggleOption(f: MLCTemplateField, opt: string) {
    const cur: string[] = values[f.id] ?? [];
    values[f.id] = cur.includes(opt) ? cur.filter((x) => x !== opt) : [...cur, opt];
  }

  async function create() {
    if (!selected) return;
    const res = await backend.renderMLCTemplate(selected.id, $state.snapshot(values));
    fieldErrors = (res.fieldErrors ?? {}) as Record<string, string>;
    diagnostics = (res.diagnostics ?? []).filter((d) => d.severity === 'error');
    if (Object.keys(fieldErrors).length > 0 || diagnostics.length > 0) return;
    onCreate(res.content, res.fileName);
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onCancel();
  }
</script>

{#if show}
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" onkeydown={onKeydown}>
  <div class="max-w-xl w-full max-h-[90vh] flex flex-col rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <div class="p-6 flex flex-col gap-3 text-sm overflow-y-auto">
      <h3 class="text-lg font-bold">{$t('tplHeading')}</h3>
      <p class="text-xs opacity-70">
        {$t('tplIntro')}
        <button type="button" class="underline hover:opacity-100" title={folder} onclick={() => backend.openTemplatesFolder()}>{$t('tplOpenFolder')}</button>
      </p>

      <label class="flex flex-col gap-1">
        <span class="text-xs font-semibold opacity-80">{$t('tplTemplate')}</span>
        <select bind:value={selectedId} class="px-2 py-1.5 rounded border {editorClass}">
          {#each groups as g}
            <optgroup label={g.label}>
              {#each g.items as tpl}
                <option value={tpl.id} disabled={!!tpl.problems?.length}>
                  {tpl.name}{fromDocFolder(tpl) ? ` · ${$t('tplOwn')}` : ''}{tpl.problems?.length ? ` · ${$t('tplBroken')}` : ''}
                </option>
              {/each}
            </optgroup>
          {/each}
        </select>
        {#if selected?.description}<span class="text-[11px] opacity-60">{selected.description}</span>{/if}
      </label>

      {#if selected?.warnings?.length}
        <div class="p-2 rounded text-[11px] bg-amber-500/10 opacity-90">{selected.warnings.join(' · ')}</div>
      {/if}

      {#each templates.filter((x) => x.problems?.length) as broken}
        <div class="p-2 rounded text-[11px] bg-amber-500/15 text-amber-600">
          <strong>{broken.name}</strong> ({broken.id}.json): {broken.problems?.join(' · ')}
        </div>
      {/each}

      {#each fields as f (f.id)}
        <div class="flex flex-col gap-1">
          <span class="text-xs font-semibold opacity-80">{f.label}{f.required ? ' *' : ''}</span>
          {#if f.type === 'multiline'}
            <textarea bind:value={values[f.id]} rows="4" placeholder={f.placeholder ?? ''} class="px-2 py-1.5 rounded border {editorClass}"></textarea>
          {:else if f.type === 'number'}
            <input type="number" bind:value={values[f.id]} min={f.min ?? undefined} max={f.max ?? undefined} step={f.step ?? 'any'} class="px-2 py-1.5 rounded border font-mono {editorClass}" />
          {:else if f.type === 'date'}
            <input type="date" bind:value={values[f.id]} class="px-2 py-1.5 rounded border font-mono {editorClass}" />
          {:else if f.type === 'time'}
            <input type="time" bind:value={values[f.id]} class="px-2 py-1.5 rounded border font-mono {editorClass}" />
          {:else if f.type === 'select'}
            <select bind:value={values[f.id]} class="px-2 py-1.5 rounded border {editorClass}">
              <option value="">—</option>
              {#each f.optionList ?? [] as opt}<option value={opt}>{opt}</option>{/each}
            </select>
          {:else if f.type === 'multiselect'}
            <div class="flex flex-wrap gap-x-4 gap-y-1">
              {#each f.optionList ?? [] as opt}
                <label class="flex items-center gap-1.5 text-xs font-mono">
                  <input type="checkbox" checked={(values[f.id] ?? []).includes(opt)} onchange={() => toggleOption(f, opt)} />{opt}
                </label>
              {/each}
            </div>
          {:else if f.type === 'docref'}
            {#if !(f.optionList ?? []).length}
              <span class="text-[11px] opacity-60">{$t('tplNoDocuments')}</span>
            {:else if f.multiple}
              <div class="flex flex-col gap-1 max-h-40 overflow-y-auto">
                {#each f.optionList ?? [] as opt, i}
                  <label class="flex items-center gap-2 text-xs">
                    <input type="checkbox" checked={(values[f.id] ?? []).includes(opt)} onchange={() => toggleOption(f, opt)} />{optionLabel(f, i)}
                  </label>
                {/each}
              </div>
            {:else}
              <select bind:value={values[f.id]} class="px-2 py-1.5 rounded border {editorClass}">
                <option value="">—</option>
                {#each f.optionList ?? [] as opt, i}<option value={opt}>{optionLabel(f, i)}</option>{/each}
              </select>
            {/if}
          {:else if f.type === 'refno'}
            <input bind:value={values[f.id]} oninput={() => (edited[f.id] = true)} placeholder="{f.prefix}-QM-001" class="px-2 py-1.5 rounded border font-mono {editorClass}" />
          {:else}
            <input bind:value={values[f.id]} placeholder={f.placeholder ?? ''} class="px-2 py-1.5 rounded border {editorClass}" />
          {/if}
          {#if fieldErrors[f.id]}<span class="text-[11px] text-red-500">{fieldErrors[f.id]}</span>{/if}
          {#if f.help}<span class="text-[11px] opacity-60">{f.help}</span>{/if}
        </div>
      {/each}

      {#if fieldErrors['']}
        <div class="p-2 rounded text-xs bg-red-600/15 text-red-500">{fieldErrors['']}</div>
      {/if}
      {#if diagnostics.length > 0}
        <div class="p-2 rounded text-xs bg-red-600/15 text-red-500">
          <strong>{$t('tplCheckFailed')}</strong>
          {#each diagnostics as d}<div>{$t('line')} {d.line}: {d.message}</div>{/each}
        </div>
      {/if}
    </div>
    <div class="flex justify-end gap-3 px-6 py-3 border-t border-black/10">
      <button onclick={onCancel} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('tplCancel')}</button>
      <button onclick={create} disabled={!selected} class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors disabled:opacity-50">{$t('tplCreate')}</button>
    </div>
  </div>
</div>
{/if}
