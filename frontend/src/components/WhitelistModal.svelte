<script lang="ts">
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';

  export let show = false;
  export let type: 'path' | 'url' = 'path';
  export let resource = "";
  export let onConfirm: () => void;
  export let onCancel: () => void;
  export let theme: 'dark' | 'light' = 'dark';

  $: toolbarClass = STYLE.toolbar[theme];
  $: buttonClass = STYLE.button[theme];
  $: editorClass = STYLE.editor[theme];
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
          on:click={onCancel}
          class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}"
        >
          {$t('deny')}
        </button>
        <button 
          on:click={onConfirm}
          class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors"
        >
          {$t('allow')}
        </button>
      </div>
    </div>
  </div>
</div>
{/if}
