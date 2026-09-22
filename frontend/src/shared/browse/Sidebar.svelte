<script lang="ts">
  import { createEventDispatcher, tick } from 'svelte';
  import TocNode from './TocNode.svelte';

  export let tabs: any[] = [];
  export let width: number = 280;
  export let activePath: string | null = null;

  const dispatch = createEventDispatcher();

  // Re-emit node selections so they bubble to App.
  function forward(e: CustomEvent<{ path: string }>) {
    dispatch('select', e.detail);
  }

  let nav: HTMLElement | null = null;

  // Scroll the marked entry into view inside the sidebar whenever
  // navigation lands on another page (content links included).
  async function revealActive() {
    await tick();
    nav
      ?.querySelector(':scope .toc-entry button.active')
      ?.scrollIntoView({ block: 'nearest', inline: 'nearest' });
  }
  $: if (activePath && tabs.length) void revealActive();
</script>

<nav class="sidebar" style="width: {width}px" aria-label="Inhaltsverzeichnis" bind:this={nav}>
  <div class="sidebar-header">
    <h2>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <rect x="3" y="3" width="18" height="18" rx="2" />
        <line x1="9" y1="3" x2="9" y2="21" />
      </svg>
      <span>Inhalt</span>
    </h2>
    <button
      class="collapse-btn"
      on:click={() => dispatch('collapse')}
      title="Verzeichnis ausblenden"
      aria-label="Verzeichnis ausblenden"
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <polyline points="15 18 9 12 15 6" />
      </svg>
    </button>
  </div>
  <ul class="toc">
    {#each tabs as entry}
      <TocNode {entry} {activePath} depth={0} on:select={forward} />
    {/each}
  </ul>
</nav>

<style>
  /* Tree styling (.toc, buttons, details triangle) lives in the global
     stylesheet so recursive TocNode levels share one definition. */
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    height: 100vh;
    background: var(--surface);
    border-right: 1px solid var(--border);
    padding: 0.875rem 0.75rem;
    overflow-y: auto;
  }
  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
    padding: 0 0.25rem;
  }
  .sidebar-header h2 {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0;
    font-family: var(--font-mono);
    font-size: 0.75rem;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--muted);
  }
  .sidebar-header h2 svg {
    width: 15px;
    height: 15px;
    color: var(--accent);
  }
  .collapse-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--muted);
    padding: 0.25rem;
  }
  .collapse-btn:hover { background: var(--surface-hover); color: var(--accent); border-color: var(--accent); }
  .collapse-btn svg { width: 14px; height: 14px; }
  .sidebar::-webkit-scrollbar { width: 8px; }
  .sidebar::-webkit-scrollbar-thumb { background: var(--border); border-radius: 4px; }
</style>
