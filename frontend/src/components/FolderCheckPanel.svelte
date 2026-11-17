<script lang="ts">
  /**
   * Findings of the folder check (.marksafe.yaml) for the whole folder,
   * grouped by file. A click opens the file at the line.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import type { FolderRun } from '../lib/folderCheck.svelte';
  import type { FolderFinding } from '../lib/backend';

  let {
    run,
    theme = 'dark',
    onOpen,
    onRerun,
    onClose,
  }: {
    run: FolderRun;
    theme?: EffectiveTheme_t;
    /** Open path at line (0: no line). */
    onOpen: (path: string, line: number) => void;
    onRerun: () => void;
    onClose: () => void;
  } = $props();

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);

  // Files in the order they come, findings without a file first.
  const groups = $derived.by(() => {
    const map = new Map<string, FolderFinding[]>();
    for (const f of run.result?.findings ?? []) {
      const key = f.file || '';
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(f);
    }
    return [...map.entries()].sort(([a], [b]) => (a === '' ? -1 : b === '' ? 1 : 0));
  });

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') onClose();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" role="dialog" aria-modal="true" aria-label={$t('folderCheckTitle')} data-testid="foldercheck-panel">
  <div class="max-w-2xl w-full max-h-[80vh] flex flex-col rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <div class="p-4 border-b border-current/10">
      <h3 class="text-lg font-bold">{$t('folderCheckTitle')}</h3>
      <p class="text-xs opacity-60 break-all">{run.dir}</p>
    </div>
    <div class="p-4 overflow-auto flex-1 text-sm">
      {#if run.running}
        <p class="opacity-70" data-testid="foldercheck-running">{$t('folderCheckRunning')}</p>
      {:else if run.refused}
        <p class="opacity-70" data-testid="foldercheck-refused">{$t('folderCheckRefused')}</p>
      {:else if run.error}
        <p class="text-red-500 whitespace-pre-line">{run.error}</p>
      {:else if run.result?.failed}
        <p class="text-red-500 font-semibold mb-2" data-testid="foldercheck-failed">{$t('folderCheckFailed')}</p>
        <pre data-testid="foldercheck-stderr" class="text-xs whitespace-pre-wrap opacity-80">{run.result.stderr}</pre>
      {:else if run.result?.error}
        <!-- Exit 2: the check's own message is the only guidance for the
             user (e.g. "rebuild the check program"), so it stands out,
             wraps and can be copied. -->
        <p class="font-semibold mb-2 {currentTheme === 'light' ? 'text-amber-700' : 'text-amber-400'}">{$t('folderCheckNotPossible')}</p>
        <p data-testid="foldercheck-error" class="select-text whitespace-pre-line break-words rounded border-l-4 border-amber-500 bg-amber-500/10 px-3 py-2">{run.result.error}</p>
      {:else if groups.length === 0}
        <p class="text-green-500" data-testid="foldercheck-ok">{$t('folderCheckOk')}</p>
      {:else}
        {#each groups as [file, findings]}
          <h4 class="font-semibold mt-3 first:mt-0 break-all">{file || $t('folderCheckGeneral')}</h4>
          <ul class="mt-1" data-testid="foldercheck-findings">
            {#each findings as f}
              <li data-testid="foldercheck-finding" data-code={f.code} data-severity={f.severity}>
                <button
                  class="w-full text-left px-2 py-1 rounded hover:bg-current/10 flex gap-2 disabled:cursor-default"
                  disabled={!f.path}
                  onclick={() => f.path && onOpen(f.path, f.line ?? 0)}
                >
                  <span class={f.severity === 'warning' ? 'text-amber-500' : 'text-red-500'}>●</span>
                  <span class="opacity-60 w-14 shrink-0">{f.line ? `${$t('line')} ${f.line}` : ''}</span>
                  <span class="flex-1">{f.message}</span>
                  {#if f.code}<span class="opacity-50 text-xs">{f.code}</span>{/if}
                </button>
              </li>
            {/each}
          </ul>
        {/each}
      {/if}
    </div>
    <div class="p-4 flex justify-end gap-3 border-t border-current/10">
      <button onclick={onRerun} data-testid="foldercheck-rerun" disabled={run.running} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('folderCheckRerun')}</button>
      <button onclick={onClose} data-testid="foldercheck-close" class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('close')}</button>
    </div>
  </div>
</div>
