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
    licenses = [],
    onClose,
    theme = 'dark'
  } = $props<{
    show?: boolean;
    title?: string;
    message?: string;
    /** Licenses of built-in modules that differ from MarkSafe's (MIT). */
    licenses?: { module: string; name: string; text: string }[];
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
  let openLicense = $state(-1);
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
      {#each licenses as l, i}
        <div class="text-xs mb-4 -mt-2">
          <span class="opacity-80">{$t('aboutModuleLicense', l.module, l.name)}</span>
          <button class="underline opacity-70 hover:opacity-100 ml-1" onclick={() => (openLicense = openLicense === i ? -1 : i)}>{$t('aboutShowLicense')}</button>
          {#if openLicense === i}
            <pre class="mt-2 p-2 max-h-48 overflow-auto rounded border border-current/20 whitespace-pre-wrap text-[10px] leading-snug select-text">{l.text}</pre>
          {/if}
        </div>
      {/each}

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
