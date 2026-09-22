<script lang="ts">
  /**
   * AboutModal: themed in-app replacement for the native About dialog,
   * so it follows the active Dark/Light mode like the rest of the UI.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';

  // --- Svelte 5 Runes: Props ---
  let {
    show = false,
    title = "",
    message = "",
    onClose,
    theme = 'dark'
  } = $props<{
    show?: boolean;
    title?: string;
    message?: string;
    onClose: () => void;
    theme?: EffectiveTheme_t;
  }>();

  // --- Svelte 5 Runes: Derived ---
  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  let closeBtn = $state<HTMLButtonElement>();

  $effect(() => {
    if (show) closeBtn?.focus();
  });
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);
</script>

<svelte:window
  onkeydown={(e) => {
    if (show && e.key === 'Escape') onClose();
  }}
/>

{#if show}
<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm print:hidden">
  <div class="max-w-md w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}" role="dialog" aria-modal="true" aria-label={title}>
    <div class="p-6">
      <h3 class="text-lg font-bold mb-2">
        {title}
      </h3>
      <p class="text-sm opacity-80 mb-6 whitespace-pre-line">
        {message}
      </p>

      <div class="flex justify-end">
        <button
          bind:this={closeBtn}
          onclick={onClose}
          class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}"
        >
          {$t('close')}
        </button>
      </div>
    </div>
  </div>
</div>
{/if}
