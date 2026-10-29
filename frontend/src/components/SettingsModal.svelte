<script lang="ts">
  /**
   * File › Settings…: a frame for the sections of built-in modules (so far
   * only the Sensor Monitor's "Connections"). Save asks each section to
   * store its values; the dialog closes when all succeed.
   */
  import type { Component } from 'svelte';
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';

  let { sections, onClose, theme = 'dark' }: {
    /** Section components; each exports save(): Promise<boolean>. */
    sections: Component<any>[];
    onClose: () => void;
    theme?: EffectiveTheme_t;
  } = $props();

  let instances = $state<any[]>([]);
  let saving = $state(false);

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);

  async function save() {
    saving = true;
    try {
      const results = await Promise.all(instances.map((s) => s?.save?.() ?? true));
      if (results.every(Boolean)) onClose();
    } finally {
      saving = false;
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" role="dialog" aria-modal="true" aria-label={$t('settingsTitle')}>
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <form class="p-6" onsubmit={(e) => { e.preventDefault(); save(); }}>
      <h3 class="text-lg font-bold mb-4">{$t('settingsTitle')}</h3>
      {#each sections as Section, i}
        <div class={i > 0 ? 'mt-6' : ''}><Section bind:this={instances[i]} {theme} /></div>
      {/each}
      <div class="flex justify-end gap-3 mt-6">
        <button type="button" onclick={onClose} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('cancel')}</button>
        <button type="submit" disabled={saving} class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors disabled:opacity-50">{$t('save')}</button>
      </div>
    </form>
  </div>
</div>
