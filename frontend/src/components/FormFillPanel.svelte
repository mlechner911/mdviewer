<script lang="ts">
  /**
   * Filling in a form (plan P-SENS-20261028-01 Phase 2): an input mask in
   * place of the preview. The fields come from mlcext.FormSchema; every
   * entry goes through mlcext.FillFields straight into the Markdown, so the
   * editor and the PDF preview always show the document as it will be
   * saved. The labels are guessed from the text around a field, so the
   * field name stands next to them.
   */
  import { mlcFormSchema, mlcFillForm, type MLCFormField } from '../lib/backend';
  import { t } from '../i18n';

  let { md, filledBy, onchange, onfilledby, onsave, onload, onclose } = $props<{
    md: string;
    filledBy: string;
    onchange: (md: string) => void;
    onfilledby: (name: string) => void;
    onsave: () => void;
    onload: () => void;
    onclose: () => void;
  }>();

  let fields = $state<MLCFormField[]>([]);
  // What the user is typing in a field: a value is stored trimmed, so the
  // input keeps its own text while it has the focus.
  let drafts = $state<Record<string, string>>({});
  let timer: ReturnType<typeof setTimeout>;

  $effect(() => {
    const source = md;
    clearTimeout(timer);
    timer = setTimeout(async () => (fields = await mlcFormSchema(source)), 120);
  });

  const sections = $derived.by(() => {
    const out: { title: string; fields: MLCFormField[] }[] = [];
    for (const f of fields) {
      const title = f.section ?? '';
      if (!out.length || out[out.length - 1].title !== title) out.push({ title, fields: [] });
      out[out.length - 1].fields.push(f);
    }
    return out;
  });
  const missing = $derived(fields.filter((f) => f.kind !== 'signature' && f.kind !== 'check' && !f.value).length);
  const signatures = $derived(fields.filter((f) => f.kind === 'signature').length);

  async function set(name: string, value: string) {
    onchange(await mlcFillForm(md, { [name]: value }));
  }
  function typed(name: string, value: string) {
    drafts[name] = value;
    void set(name, value);
  }
  function done(name: string) {
    delete drafts[name];
  }
  const shown = (f: MLCFormField) => drafts[f.name] ?? f.value;
</script>

<div class="fill">
  <div class="bar">
    <span class="title">{$t('formFillTitle')}</span>
    <span class="count">{fields.length} {$t('formFillFields')}{#if missing} · {missing} {$t('formFillEmpty')}{/if}</span>
    <div class="spacer"></div>
    <button onclick={onload} title={$t('formFillLoadHint')}>{$t('formFillLoad')}</button>
    <button class="primary" onclick={onsave} title={$t('formFillSaveHint')}>{$t('formFillSave')}</button>
    <button onclick={onclose} title={$t('formFillCloseHint')} aria-label={$t('formFillCloseHint')}>✕</button>
  </div>
  <div class="body">
    <label class="by">
      <span>{$t('formFillBy')}</span>
      <input value={filledBy} oninput={(e) => onfilledby(e.currentTarget.value)} placeholder={$t('formFillByHint')} autocomplete="off" />
    </label>
    {#if !fields.length}
      <p class="empty">{$t('formFillNone')}</p>
    {/if}
    {#each sections as s (s.title + s.fields[0]?.name)}
      <section>
        {#if s.title}<h3>{s.title}</h3>{/if}
        {#each s.fields as f (f.name)}
          <div class="row" class:signature={f.kind === 'signature'}>
            {#if f.kind === 'check'}
              <label class="check">
                <input type="checkbox" checked={f.value === 'x'} onchange={(e) => set(f.name, e.currentTarget.checked ? 'x' : '')} />
                <span>{f.label || f.name}</span>
              </label>
            {:else if f.kind === 'signature'}
              <span class="label">{f.label || f.role || f.name}</span>
              <span class="sign">{$t('formFillSignature')}</span>
            {:else}
              <label for="ff-{f.name}" class="label">
                {#if f.group}<span class="group">{f.group} ·</span>{/if}
                {f.label || f.name}
              </label>
              {#if f.kind === 'multiline'}
                <textarea id="ff-{f.name}" rows={Math.max(f.rows ?? 2, 2)} value={shown(f)} oninput={(e) => typed(f.name, e.currentTarget.value)} onblur={() => done(f.name)}></textarea>
              {:else}
                <input id="ff-{f.name}" value={shown(f)} placeholder={f.pattern ?? ''} maxlength={f.pattern ? [...f.pattern].length : undefined} oninput={(e) => typed(f.name, e.currentTarget.value)} onblur={() => done(f.name)} autocomplete="off" />
              {/if}
            {/if}
            <span class="name" title={f.tooltip ?? ''}>{f.name}{#if f.prefill} · {$t('formFillPrefill')}{/if}</span>
          </div>
        {/each}
      </section>
    {/each}
    {#if signatures}<p class="note">{$t('formFillSignatureNote')}</p>{/if}
  </div>
</div>

<style>
  .fill { display: flex; flex-direction: column; flex: 1; min-height: 0; font-size: 13px; }
  .bar { display: flex; align-items: center; gap: 0.5rem; padding: 0.4rem 0.75rem; border-bottom: 1px solid color-mix(in srgb, currentColor 15%, transparent); }
  .title { font-weight: 700; }
  .count { opacity: 0.6; font-size: 12px; }
  .spacer { flex: 1; }
  button { padding: 0.15rem 0.6rem; border-radius: 4px; border: 1px solid color-mix(in srgb, currentColor 25%, transparent); font-size: 12px; }
  button:hover { background: color-mix(in srgb, currentColor 10%, transparent); }
  button.primary { background: #c94c35; border-color: #c94c35; color: white; }
  .body { overflow-y: auto; padding: 0.75rem 1rem 2rem; display: flex; flex-direction: column; gap: 0.75rem; }
  .by { display: flex; align-items: center; gap: 0.5rem; padding-bottom: 0.5rem; border-bottom: 1px dashed color-mix(in srgb, currentColor 20%, transparent); }
  .by span { opacity: 0.75; }
  h3 { font-weight: 700; margin: 0.25rem 0 0.4rem; font-size: 13px; }
  .row { display: grid; grid-template-columns: minmax(8rem, 14rem) 1fr; column-gap: 0.75rem; row-gap: 0.1rem; align-items: center; padding: 0.2rem 0; }
  .row .name { grid-column: 2; font-family: var(--mlc-md-font-mono, monospace); font-size: 10.5px; opacity: 0.45; }
  .label { opacity: 0.85; }
  .group { opacity: 0.6; }
  .check { grid-column: 1 / span 2; display: flex; gap: 0.5rem; align-items: center; }
  input:not([type='checkbox']), textarea { width: 100%; padding: 0.2rem 0.4rem; border-radius: 4px; border: 1px solid color-mix(in srgb, currentColor 25%, transparent); background: transparent; color: inherit; font: inherit; }
  input:focus, textarea:focus { outline: 2px solid color-mix(in srgb, #c94c35 60%, transparent); outline-offset: 0; }
  .sign { font-style: italic; opacity: 0.6; border-bottom: 1px solid color-mix(in srgb, currentColor 40%, transparent); padding: 0.6rem 0 0.1rem; }
  .empty, .note { opacity: 0.65; font-size: 12px; }
</style>
