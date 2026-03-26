<script lang="ts">
  /**
   * WhitelistModal component for MD Viewer.
   * Refactored for Svelte 5 Runes.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';

  // --- Svelte 5 Runes: Props ---
  let { 
    show = false, 
    type = 'path', 
    resource = "", 
    onConfirm, 
    onCancel, 
    theme = 'dark' 
  } = $props<{
    show?: boolean;
    type?: 'path' | 'url';
    resource?: string;
    onConfirm: () => void;
    onCancel: () => void;
    theme?: 'dark' | 'light';
  }>();

  // --- Svelte 5 Runes: Derived ---
  const toolbarClass = $derived(STYLE.toolbar[theme]);
  const buttonClass = $derived(STYLE.button[theme]);
  const editorClass = $derived(STYLE.editor[theme]);
</script>

{#if show}
<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <div class="p-6">
      <h3 class="text-lg font-bold mb-2">
        {type === 'path' ? $t('securityWarningPath') : $t('securityWarningURL')}
      </h3>
      <p class="text-sm opacity-80 mb-4">
        {type === 'path' ? $t('securityDescPath') : $t('securityDescURL')}
      </p>
      
      <div class="p-3 rounded font-mono text-xs break-all mb-6 {editorClass}">
        {resource}
      </div>

      <div class="flex justify-end gap-3">
        <button 
          onclick={onCancel}
          class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}"
        >
          {$t('deny')}
        </button>
        <button 
          onclick={onConfirm}
          class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors"
        >
          {$t('allow')}
        </button>
      </div>
    </div>
  </div>
</div>
{/if}
