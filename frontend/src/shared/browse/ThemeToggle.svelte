<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  let { theme = 'dark' }: { theme?: string } = $props();
  const dispatch = createEventDispatcher();

  function handleClick() {
    dispatch('toggle');
  }
</script>

<button
  class="theme-toggle"
  onclick={handleClick}
  title="Theme: {theme} — klicken zum Wechseln"
>
  {#if theme === 'dark'}
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
      stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
    </svg>
    <span>Dark</span>
  {:else}
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
      stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <circle cx="12" cy="12" r="5" />
      <line x1="12" y1="1" x2="12" y2="3" />
      <line x1="12" y1="21" x2="12" y2="23" />
      <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
      <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
      <line x1="1" y1="12" x2="3" y2="12" />
      <line x1="21" y1="12" x2="23" y2="12" />
      <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
      <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
    </svg>
    <span>Light</span>
  {/if}
</button>

<style>
  .theme-toggle {
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.45rem 0.9rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: var(--surface);
    color: var(--text);
    font-family: var(--font-sans);
    font-size: 0.875rem;
    box-shadow: var(--shadow);
  }
  .theme-toggle:hover { border-color: var(--accent); color: var(--accent); }
  .theme-toggle svg { width: 16px; height: 16px; }
</style>
