<script lang="ts">
  /**
   * A yes/no question, e.g. before saving an edited signed document. With
   * altLabel a third answer (e.g. "Don't save" between Save and Cancel).
   */
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';

  let {
    title = '',
    message = '',
    confirmLabel = 'OK',
    cancelLabel = 'Cancel',
    onConfirm,
    onCancel,
    altLabel = '',
    onAlt,
    confirmClass = 'bg-red-600 hover:bg-red-500',
    theme = 'dark',
  } = $props<{
    title?: string;
    message?: string;
    confirmLabel?: string;
    cancelLabel?: string;
    onConfirm: () => void;
    onCancel: () => void;
    altLabel?: string;
    onAlt?: () => void;
    confirmClass?: string;
    theme?: EffectiveTheme_t;
  }>();

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') onCancel();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" role="dialog" aria-modal="true" aria-label={title}>
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <div class="p-6">
      <h3 class="text-lg font-bold mb-2">{title}</h3>
      <p class="text-sm opacity-80 mb-6">{message}</p>
      <div class="flex justify-end gap-3">
        <button onclick={onCancel} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{cancelLabel}</button>
        {#if altLabel}
          <button onclick={() => onAlt?.()} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{altLabel}</button>
        {/if}
        <button onclick={onConfirm} class="px-4 py-2 rounded text-sm font-medium text-white transition-colors {confirmClass}">{confirmLabel}</button>
      </div>
    </div>
  </div>
</div>
