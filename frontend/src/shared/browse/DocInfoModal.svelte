<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let title: string = '';
  export let path: string = '';
  export let size: number = 0;
  export let modified: string = '';
  export let words: number = 0;
  export let chars: number = 0;

  const dispatch = createEventDispatcher();

  function formatSize(b: number): string {
    if (!Number.isFinite(b) || b < 0) return '–';
    if (b < 1024) return `${b} B`;
    const units = ['KB', 'MB', 'GB'];
    let v = b;
    let i = -1;
    do {
      v /= 1024;
      i++;
    } while (v >= 1024 && i < units.length - 1);
    return `${v.toFixed(1)} ${units[i]}`;
  }

  function formatDate(iso: string): string {
    if (!iso) return '–';
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return '–';
    return d.toLocaleString('de-DE', { dateStyle: 'medium', timeStyle: 'short' });
  }

  $: readingMins = Math.max(1, Math.ceil(words / 225));

  function close() {
    dispatch('close');
  }

  function onWindowKey(e: KeyboardEvent) {
    if (e.key === 'Escape') close();
  }
</script>

<svelte:window on:keydown={onWindowKey} />

<div class="info-overlay" role="presentation" on:click|self={close}>
  <div class="info-modal" role="dialog" aria-modal="true" aria-label="Dokumentinformationen">
    <div class="info-head">
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="16" x2="12" y2="12" />
        <line x1="12" y1="8" x2="12.01" y2="8" />
      </svg>
      <strong>Dokumentinfo</strong>
      <button class="info-x" on:click={close} aria-label="Schließen">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
          stroke-linecap="round" aria-hidden="true">
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      </button>
    </div>
    <dl class="info-grid">
      <dt>Titel</dt>
      <dd>{title || '–'}</dd>
      <dt>Pfad</dt>
      <dd class="mono" title={path}>{path || '–'}</dd>
      <dt>Größe</dt>
      <dd class="mono">{formatSize(size)}</dd>
      <dt>Geändert</dt>
      <dd class="mono">{formatDate(modified)}</dd>
      <dt>Wörter</dt>
      <dd class="mono">{words}</dd>
      <dt>Zeichen</dt>
      <dd class="mono">{chars}</dd>
      <dt>Lesezeit</dt>
      <dd class="mono">ca. {readingMins} Min.</dd>
    </dl>
  </div>
</div>

<style>
  .info-overlay {
    position: fixed;
    inset: 0;
    z-index: 60;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding: 18vh 1rem 1rem;
    background: rgba(15, 23, 42, 0.55);
  }
  .info-modal {
    width: min(440px, 92vw);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.35);
    padding: 1.25rem;
  }
  .info-head {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    color: var(--accent-strong);
    font-size: 1.05rem;
    margin-bottom: 0.75rem;
  }
  .info-head svg { width: 20px; height: 20px; }
  .info-x {
    margin-left: auto;
    cursor: pointer;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--muted);
    padding: 0.15rem 0.5rem;
    font-size: 0.85rem;
    line-height: 1.4;
  }
  .info-x:hover { color: var(--accent); border-color: var(--accent); }
  .info-x svg { width: 12px; height: 12px; display: block; }
  .info-grid {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.45rem 1rem;
    margin: 0;
  }
  .info-grid dt { color: var(--muted); font-size: 0.85rem; }
  .info-grid dd {
    margin: 0;
    font-size: 0.9rem;
    overflow-wrap: anywhere;
  }
  .info-grid dd.mono { font-family: var(--font-mono); font-size: 0.82rem; }
</style>
