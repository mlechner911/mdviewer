<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';

  let {
    url,
    showReuse = true,
  }: {
    url: string;
    showReuse?: boolean;
  } = $props();

  const dispatch = createEventDispatcher();
  let reuse = $state(false);
  let openBtn = $state<HTMLButtonElement | null>(null);

  try {
    reuse = localStorage.getItem('marksafe-external-reuse') === '1';
  } catch {
    // ignore
  }

  onMount(() => {
    openBtn?.focus();
  });

  function confirm() {
    try {
      localStorage.setItem('marksafe-external-reuse', reuse ? '1' : '0');
    } catch {
      // ignore
    }
    dispatch('confirm', { url, reuse });
  }

  function cancel() {
    dispatch('cancel');
  }

  function onWindowKey(e: KeyboardEvent) {
    if (e.key === 'Escape') cancel();
  }
</script>

<svelte:window onkeydown={onWindowKey} />

<div class="confirm-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) cancel(); }}>
  <div
    class="confirm-modal"
    role="alertdialog"
    aria-modal="true"
    aria-label="Externer Link"
    aria-describedby="ext-url"
  >
    <div class="confirm-head">
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
        <line x1="12" y1="9" x2="12" y2="13" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
      <strong>Externer Link</strong>
    </div>
    <p class="confirm-text">
      Diese Seite führt zu einer Website außerhalb deines Verzeichnisses. Wirklich öffnen?
    </p>
    <div id="ext-url" class="confirm-url" title={url}>{url}</div>
    {#if showReuse}
    <label class="confirm-reuse">
      <input type="checkbox" bind:checked={reuse} />
      <span>Immer im selben externen Fenster öffnen (statt neuem Tab)</span>
    </label>
    {/if}
    <div class="confirm-actions">
      <button class="btn-ghost" onclick={cancel}>Abbrechen</button>
      <button class="btn-primary" bind:this={openBtn} onclick={confirm}>Extern öffnen</button>
    </div>
  </div>
</div>

<style>
  .confirm-overlay {
    position: fixed;
    inset: 0;
    z-index: 60;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding: 18vh 1rem 1rem;
    background: rgba(15, 23, 42, 0.55);
  }
  .confirm-modal {
    width: min(480px, 92vw);
    background: var(--surface);
    border: 1px solid var(--border);
    border-left: 3px solid var(--accent);
    border-radius: var(--radius-md);
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.35);
    padding: 1.25rem;
  }
  .confirm-head {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    color: var(--accent-strong);
    font-size: 1.05rem;
  }
  .confirm-head svg { width: 20px; height: 20px; }
  .confirm-text { margin: 0.75rem 0; line-height: 1.55; }
  .confirm-url {
    font-family: var(--font-mono);
    font-size: 0.8rem;
    color: var(--muted);
    background: var(--code-bg);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 0.5rem 0.7rem;
    overflow-wrap: anywhere;
  }
  .confirm-reuse {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    margin-top: 0.8rem;
    font-size: 0.85rem;
    color: var(--muted);
    cursor: pointer;
  }
  .confirm-reuse input { margin-top: 0.2rem; accent-color: var(--accent); }
  .confirm-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.6rem;
    margin-top: 1.1rem;
  }
  .btn-ghost {
    cursor: pointer;
    padding: 0.5rem 1rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text);
    font-family: var(--font-sans);
    font-size: 0.9rem;
  }
  .btn-ghost:hover { border-color: var(--accent); color: var(--accent); }
  .btn-primary {
    cursor: pointer;
    padding: 0.5rem 1rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--accent);
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-family: var(--font-sans);
    font-size: 0.9rem;
    font-weight: 600;
  }
  .btn-primary:hover { background: var(--accent); border-color: var(--accent); color: #fff; }
</style>
