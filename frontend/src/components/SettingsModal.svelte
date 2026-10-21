<script lang="ts">
  /**
   * Settings. For now one section, "Connections": the address of an MLC
   * Sensor Monitor. Empty keeps the connection off, and MarkSafe shows
   * nothing of it (plan P-MARKS-20261021-01).
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as backend from '../lib/backend';

  let { address = '', onSaved, onClose, theme = 'dark' } = $props<{
    address?: string;
    onSaved: (address: string) => void;
    onClose: () => void;
    theme?: EffectiveTheme_t;
  }>();

  // svelte-ignore state_referenced_locally
  let value = $state(address);
  let error = $state('');
  let saving = $state(false);

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);
  const inputClass = $derived(currentTheme === 'dark' ? 'bg-slate-800 border-slate-600' : 'bg-white border-gray-300');

  async function save() {
    saving = true;
    error = '';
    try {
      const normalized = await backend.setSensorMonitorURL(value.trim());
      onSaved(normalized);
    } catch (err: any) {
      error = String(err?.message ?? err);
    } finally {
      saving = false;
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm print:hidden" role="dialog" aria-modal="true" aria-label={$t('settingsTitle')}>
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <form class="p-6" onsubmit={(e) => { e.preventDefault(); save(); }}>
      <h3 class="text-lg font-bold mb-4">{$t('settingsTitle')}</h3>
      <h4 class="text-xs font-semibold uppercase tracking-wide opacity-70 mb-2">{$t('settingsConnections')}</h4>
      <label class="block text-sm font-medium mb-1" for="sm-address">{$t('smAddress')}</label>
      <input
        id="sm-address"
        type="text"
        bind:value
        placeholder={$t('smAddressPlaceholder')}
        spellcheck="false"
        autocomplete="off"
        class="w-full px-3 py-2 rounded border text-sm select-text {inputClass}"
      />
      <p class="text-xs opacity-70 mt-2">{$t('smAddressHint')}</p>
      {#if error}<p class="text-xs text-red-500 mt-2">{error}</p>{/if}
      <div class="flex justify-end gap-3 mt-6">
        <button type="button" onclick={onClose} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('cancel')}</button>
        <button type="submit" disabled={saving} class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors disabled:opacity-50">{$t('save')}</button>
      </div>
    </form>
  </div>
</div>
