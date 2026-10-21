<script lang="ts">
  /**
   * Sign-in to the Sensor Monitor by device login: MarkSafe opens the
   * confirmation page in the browser and shows the code to compare. The
   * backend waits and reports with the event "sensormonitor-login".
   */
  import { onMount } from 'svelte';
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as backend from '../lib/backend';
  import { EventsOn, BrowserOpenURL } from '../lib/wails';

  let { onDone, onClose, theme = 'dark' } = $props<{
    /** The sign-in succeeded. */
    onDone: () => void;
    onClose: () => void;
    theme?: EffectiveTheme_t;
  }>();

  let code = $state<backend.SensorLoginCode | null>(null);
  let error = $state('');
  let starting = $state(false);

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);

  async function start() {
    starting = true;
    error = '';
    code = null;
    try {
      code = await backend.sensorMonitorStartLogin();
    } catch (err: any) {
      error = $t('smLoginFailed', String(err?.message ?? err));
    } finally {
      starting = false;
    }
  }

  function close() {
    if (code) void backend.sensorMonitorCancelLogin();
    onClose();
  }

  onMount(() => {
    const off = EventsOn('sensormonitor-login', (res: any) => {
      const r = Array.isArray(res) ? res[0] : res;
      if (!r) return;
      if (r.ok) { code = null; onDone(); return; }
      if (r.error === 'cancelled' || r.error === 'signed-out') return;
      code = null;
      error = r.error === 'denied' ? $t('smLoginDenied') : r.error === 'expired' ? $t('smLoginExpired') : $t('smLoginFailed', r.error);
    });
    void start();
    return off;
  });

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Escape') close();
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm print:hidden" role="dialog" aria-modal="true" aria-label={$t('smLoginTitle')}>
  <div class="max-w-md w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <div class="p-6">
      <h3 class="text-lg font-bold mb-4">{$t('smLoginTitle')}</h3>
      {#if code}
        <p class="text-sm opacity-80 mb-3">{$t('smLoginStep')}</p>
        <div class="text-3xl font-mono font-bold tracking-widest text-center py-3 select-text">{code.userCode}</div>
        <p class="text-xs opacity-70 text-center mb-4">{$t('smLoginWaiting')}</p>
        <div class="text-center">
          <button onclick={() => code && BrowserOpenURL(code.url)} class="text-xs underline opacity-80 hover:opacity-100">{$t('smLoginOpenAgain')}</button>
        </div>
      {:else if starting}
        <p class="text-sm opacity-70">…</p>
      {/if}
      {#if error}<p class="text-sm text-red-500 mt-2">{error}</p>{/if}
      <div class="flex justify-end gap-3 mt-6">
        {#if error}
          <button onclick={start} class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors">{$t('smRetry')}</button>
        {/if}
        <button onclick={close} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{code ? $t('cancel') : $t('close')}</button>
      </div>
    </div>
  </div>
</div>
