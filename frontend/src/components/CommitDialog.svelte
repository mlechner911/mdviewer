<script lang="ts">
  /**
   * Commit after a clean folder check (.marksafe.yaml with commit: true).
   * Only the saved file is committed, never pushed.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';

  let {
    path,
    message: initial,
    theme = 'dark',
    onCommit,
    onCancel,
  }: {
    path: string;
    message: string;
    theme?: EffectiveTheme_t;
    /** Returns an error text to show, or null when committed. */
    onCommit: (message: string) => Promise<string | null>;
    onCancel: () => void;
  } = $props();

  // Prefilled once; the user's edits are kept from then on.
  // svelte-ignore state_referenced_locally
  let message = $state(initial);
  let error = $state('');
  let busy = $state(false);
  let input = $state<HTMLTextAreaElement>();

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);
  const inputClass = $derived(currentTheme === 'dark' ? 'bg-slate-800 border-slate-600' : 'bg-white border-gray-300');

  $effect(() => {
    if (!input) return;
    input.focus();
    input.setSelectionRange(input.value.length, input.value.length);
  });

  async function submit() {
    if (busy || !message.trim()) return;
    busy = true;
    error = (await onCommit(message)) ?? '';
    busy = false;
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape' && !busy) onCancel();
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) void submit();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" role="dialog" aria-modal="true" aria-label={$t('commitTitle')}>
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <div class="p-6">
      <h3 class="text-lg font-bold mb-1">{$t('commitTitle')}</h3>
      <p class="text-xs opacity-60 break-all mb-3">{path}</p>
      <p class="text-sm opacity-80 mb-3">{$t('commitHint')}</p>
      <label class="text-sm font-medium" for="commit-message">{$t('commitMessage')}</label>
      <textarea
        id="commit-message"
        bind:this={input}
        bind:value={message}
        rows="3"
        class="w-full mt-1 p-2 rounded border text-sm font-mono {inputClass}"
      ></textarea>
      {#if error}<p class="text-sm text-red-500 mt-2 whitespace-pre-line break-words">{error}</p>{/if}
      <div class="flex justify-end gap-3 mt-4">
        <button onclick={onCancel} disabled={busy} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('commitLater')}</button>
        <button onclick={submit} disabled={busy || !message.trim()} class="px-4 py-2 rounded text-sm font-medium text-white transition-colors bg-blue-600 hover:bg-blue-500 disabled:opacity-50">{$t('commitButton')}</button>
      </div>
    </div>
  </div>
</div>
