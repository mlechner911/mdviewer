<script lang="ts">
  /**
   * Open from the Sensor Monitor: look up a document id, show the approved
   * version and a waiting draft, open one of them in a new tab.
   */
  import { onMount } from 'svelte';
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as api from './api';
  import { sensorErrorText } from './errors';

  let { initialRef = '', onOpen, onSignIn, onClose, theme = 'dark' } = $props<{
    initialRef?: string;
    /** Open text as a new tab; doc is the version it came from. */
    onOpen: (text: string, doc: api.SensorDocument) => void;
    /** The sign-in ran out: sign in again. */
    onSignIn: () => void;
    onClose: () => void;
    theme?: EffectiveTheme_t;
  }>();

  // svelte-ignore state_referenced_locally
  let ref = $state(initialRef);
  let stand = $state<api.Stand | null>(null);
  let error = $state('');
  let busy = $state(false);

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);
  const inputClass = $derived(currentTheme === 'dark' ? 'bg-slate-800 border-slate-600' : 'bg-white border-gray-300');
  const cardClass = $derived(currentTheme === 'dark' ? 'border-slate-600' : 'border-gray-300');

  function fail(e: api.SensorError) {
    if (e.signIn) { onSignIn(); return; }
    error = sensorErrorText(e, $t);
  }

  async function search() {
    const r = ref.trim();
    if (!r) return;
    busy = true; error = ''; stand = null;
    const res = await api.lookup(r);
    busy = false;
    if (res.error) { fail(res.error); return; }
    stand = res.stand;
  }

  async function open(doc: api.SensorDocument) {
    busy = true; error = '';
    const res = await api.source(doc.id);
    busy = false;
    if (res.error) { fail(res.error); return; }
    onOpen(res.text, doc);
  }

  function day(s: string): string {
    if (!s) return '';
    const d = new Date(s);
    return isNaN(d.getTime()) ? s : d.toLocaleDateString();
  }

  onMount(() => { if (ref.trim()) void search(); });

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm print:hidden" role="dialog" aria-modal="true" aria-label={$t('smOpenTitle')}>
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <div class="p-6">
      <h3 class="text-lg font-bold mb-4">{$t('smOpenTitle')}</h3>
      <form class="flex gap-2" onsubmit={(e) => { e.preventDefault(); search(); }}>
        <label class="sr-only" for="sm-ref">{$t('smRefLabel')}</label>
        <input id="sm-ref" type="text" bind:value={ref} placeholder={$t('smRefLabel')} spellcheck="false" autocomplete="off"
          class="flex-1 px-3 py-2 rounded border text-sm font-mono select-text {inputClass}" />
        <button type="submit" disabled={busy || !ref.trim()} class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white disabled:opacity-50">{$t('smSearch')}</button>
      </form>

      {#if stand}
        {#if !stand.approved && !stand.draft}
          <p class="text-sm opacity-70 mt-4">{$t('smNone')}</p>
        {/if}
        {#each [{ doc: stand.approved, label: $t('smApproved') }, { doc: stand.draft, label: $t('smDraft') }] as item}
          {#if item.doc}
            <div class="mt-4 p-3 rounded border flex items-center gap-3 {cardClass}">
              <div class="flex-1 min-w-0 text-sm">
                <div class="text-xs uppercase tracking-wide opacity-70">{item.label}</div>
                <div class="font-medium truncate">{item.doc.title || item.doc.ref_no}</div>
                <div class="text-xs opacity-70">
                  {#if item.doc.version}{$t('smVersion', item.doc.version)}{/if}
                  {#if item.doc.approved_at} · {$t('smApprovedAt', day(item.doc.approved_at))}{/if}
                  {#if item.doc.valid_until} · {$t('smReviewDue', day(item.doc.valid_until))}{/if}
                  {#if !item.doc.approved_at && item.doc.uploaded_by_label} · {$t('smUploadedBy', item.doc.uploaded_by_label)}{/if}
                </div>
              </div>
              <button onclick={() => item.doc && open(item.doc)} disabled={busy} class="px-3 py-1.5 rounded text-sm font-medium transition-colors {buttonClass}">{$t('smOpenBtn')}</button>
            </div>
          {/if}
        {/each}
      {/if}

      {#if error}<p class="text-sm text-red-500 mt-3">{error}</p>{/if}
      <div class="flex justify-end mt-6">
        <button onclick={onClose} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('close')}</button>
      </div>
    </div>
  </div>
</div>
