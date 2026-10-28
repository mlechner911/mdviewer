<script lang="ts">
  /**
   * Save the document in the Sensor Monitor as a draft (never approved).
   * Reason always offered, PIN only after the server asks for it.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as api from './api';
  import { sensorErrorText, needsPin } from './errors';

  let { markdown, filename = '', refNo = '', version = '', title = '', hasErrors = false, onSaved, onSignIn, onClose, theme = 'dark' } = $props<{
    markdown: string;
    filename?: string;
    refNo?: string;
    version?: string;
    title?: string;
    /** The local check reports errors (the server would refuse). */
    hasErrors?: boolean;
    onSaved: (draft: api.SavedDraft) => void;
    onSignIn: () => void;
    onClose: () => void;
    theme?: EffectiveTheme_t;
  }>();

  let reason = $state('');
  let pin = $state('');
  let showPin = $state(false);
  let error = $state('');
  let busy = $state(false);

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);
  const inputClass = $derived(currentTheme === 'dark' ? 'bg-slate-800 border-slate-600' : 'bg-white border-gray-300');

  async function save() {
    busy = true; error = '';
    const res = await api.saveDraft(markdown, filename, reason.trim(), showPin ? pin : '');
    busy = false;
    if (res.error) {
      if (res.error.signIn) { onSignIn(); return; }
      if (needsPin(res.error)) { showPin = true; pin = ''; }
      error = sensorErrorText(res.error, $t);
      return;
    }
    if (res.draft) onSaved(res.draft);
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm print:hidden" role="dialog" aria-modal="true" aria-label={$t('smDraftTitle')}>
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <form class="p-6" onsubmit={(e) => { e.preventDefault(); save(); }}>
      <h3 class="text-lg font-bold mb-2">{$t('smDraftTitle')}</h3>
      <p class="text-xs opacity-70 mb-4">{$t('smDraftInfo')}</p>
      {#if refNo}
        <div class="text-sm mb-4">
          <span class="font-mono font-medium">{refNo}</span>
          {#if version} · {$t('smVersion', version)}{/if}
          {#if title} · {title}{/if}
        </div>
      {:else}
        <p class="text-sm text-amber-500 mb-4">{$t('smDraftNoId')}</p>
      {/if}
      {#if hasErrors}<p class="text-sm text-amber-500 mb-4">{$t('smDraftHasErrors')}</p>{/if}

      <label class="block text-sm font-medium mb-1" for="sm-reason">{$t('smReason')}</label>
      <textarea id="sm-reason" bind:value={reason} rows="3" class="w-full px-3 py-2 rounded border text-sm select-text {inputClass}"></textarea>
      <p class="text-xs opacity-70 mt-1">{$t('smReasonHint')}</p>

      {#if showPin}
        <label class="block text-sm font-medium mt-4 mb-1" for="sm-pin">{$t('smPin')}</label>
        <!-- svelte-ignore a11y_autofocus -->
        <input id="sm-pin" type="password" bind:value={pin} autocomplete="off" autofocus class="w-40 px-3 py-2 rounded border text-sm {inputClass}" />
      {/if}

      {#if error}<p class="text-sm text-red-500 mt-3">{error}</p>{/if}
      <div class="flex justify-end gap-3 mt-6">
        <button type="button" onclick={onClose} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('cancel')}</button>
        <button type="submit" disabled={busy || !refNo || (showPin && !pin.trim())} class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white disabled:opacity-50">{$t('smSaveDraftBtn')}</button>
      </div>
    </form>
  </div>
</div>
