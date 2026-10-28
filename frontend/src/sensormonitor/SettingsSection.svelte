<script lang="ts">
  /**
   * Section "Connections" in File › Settings…: the address of an MLC
   * Sensor Monitor. Empty keeps the connection off, and MarkSafe shows
   * nothing of it. The settings dialog calls save().
   */
  import { onMount } from 'svelte';
  import { t } from '../i18n';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as api from './api';

  let { theme = 'dark' } = $props<{ theme?: EffectiveTheme_t }>();

  let value = $state('');
  let error = $state('');

  const inputClass = $derived(theme === 'light' ? 'bg-white border-gray-300' : 'bg-slate-800 border-slate-600');

  onMount(async () => { value = (await api.status()).address; });

  /** Stores the address; false (with the reason shown) when it is not valid. */
  export async function save(): Promise<boolean> {
    error = '';
    try {
      await api.setAddress(value.trim());
      // SensorMonitor.svelte updates status, menu and comparison.
      window.dispatchEvent(new CustomEvent('sensormonitor-changed'));
      return true;
    } catch (err: any) {
      error = String(err?.message ?? err);
      return false;
    }
  }
</script>

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
