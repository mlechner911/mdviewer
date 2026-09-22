<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import type { BrowseClient, BrowseHit } from './browseClient';

  export let client: BrowseClient;

  const dispatch = createEventDispatcher();
  let q = '';
  let results: BrowseHit[] = [];
  let active = 0;
  let loading = false;
  let timer: ReturnType<typeof setTimeout> | null = null;
  let input: HTMLInputElement | null = null;

  onMount(() => {
    input?.focus();
    return () => {
      if (timer) clearTimeout(timer);
    };
  });

  function onInput() {
    if (timer) clearTimeout(timer);
    timer = setTimeout(run, 180);
  }

  async function run() {
    const query = q.trim();
    if (!query) {
      results = [];
      active = 0;
      loading = false;
      return;
    }
    loading = true;
    try {
      results = (await client.search(query)) ?? [];
    } catch {
      results = [];
    } finally {
      loading = false;
      active = 0;
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (results.length) active = (active + 1) % results.length;
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (results.length) active = (active - 1 + results.length) % results.length;
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (results[active]) choose(active);
    }
  }

  function onWindowKey(e: KeyboardEvent) {
    if (e.key === 'Escape') close();
  }

  function choose(i: number) {
    const hit = results[i];
    if (hit) dispatch('open', { path: hit.path });
  }

  function close() {
    dispatch('close');
  }
</script>

<svelte:window on:keydown={onWindowKey} />

<div class="search-overlay" role="presentation" on:click|self={close}>
  <div class="search-modal" role="dialog" aria-modal="true" aria-label="Dokumente durchsuchen">
    <div class="search-box">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="11" cy="11" r="8" />
        <line x1="21" y1="21" x2="16.65" y2="16.65" />
      </svg>
      <input
        bind:this={input}
        bind:value={q}
        on:input={onInput}
        on:keydown={onKey}
        placeholder="Dokumente durchsuchen …"
        aria-label="Suchbegriff"
        autocomplete="off"
        spellcheck="false"
      />
      <kbd>esc</kbd>
    </div>
    <div class="search-results" role="listbox" aria-label="Suchergebnisse">
      {#if loading}
        <div class="search-state">Suche läuft …</div>
      {:else if !q.trim()}
        <div class="search-state">Titel, Dateinamen und Volltext werden durchsucht — Fuzzy inklusive.</div>
      {:else if results.length === 0}
        <div class="search-state">Keine Treffer für „{q.trim()}“.</div>
      {:else}
        {#each results as hit, i}
          <button
            class="search-row"
            class:selected={i === active}
            role="option"
            aria-selected={i === active}
            on:click={() => choose(i)}
            on:mousemove={() => (active = i)}
          >
            <div class="search-title">{hit.title}</div>
            <div class="search-snippet">{@html hit.snippet}</div>
            <div class="search-path">{hit.path}</div>
          </button>
        {/each}
      {/if}
    </div>
    <div class="search-footer"><kbd>Pfeiltasten</kbd> navigieren · <kbd>Enter</kbd> öffnen · <kbd>Esc</kbd> schließen</div>
  </div>
</div>

<style>
  .search-overlay {
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding: 12vh 1rem 1rem;
    background: rgba(15, 23, 42, 0.55);
  }
  .search-modal {
    width: min(640px, 92vw);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 4px;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.35);
    overflow: hidden;
  }
  .search-box {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin: 0.6rem 0.6rem 0;
    padding: 0.75rem 0.9rem 0.75rem 1rem;
    border: 1px solid var(--border);
    border-radius: 4px;
    background: var(--bg);
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .search-box:focus-within {
    border-color: var(--accent);
    box-shadow: 0 0 0 2px var(--accent-soft);
  }
  .search-box svg {
    width: 18px;
    height: 18px;
    flex: none;
    color: var(--muted);
  }
  .search-box input {
    flex: 1;
    min-width: 0;
    border: none;
    outline: none;
    background: transparent;
    color: var(--text);
    font-family: var(--font-sans);
    font-size: 1.05rem;
  }
  .search-box input::placeholder { color: var(--muted); }
  /* No fat browser outline: the field ring above signals focus. */
  .search-box input:focus-visible { outline: none; }
  kbd {
    font-family: var(--font-mono);
    font-size: 0.7rem;
    color: var(--muted);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 0.1rem 0.4rem;
    background: var(--bg);
  }
  .search-results {
    max-height: 50vh;
    overflow-y: auto;
    padding: 0.4rem;
    scrollbar-width: thin;
    scrollbar-color: var(--border) transparent;
  }
  .search-results::-webkit-scrollbar { width: 6px; }
  .search-results::-webkit-scrollbar-track { background: transparent; }
  .search-results::-webkit-scrollbar-thumb {
    background: var(--border);
    border-radius: 3px;
  }
  .search-state {
    padding: 1.5rem 1rem;
    text-align: center;
    color: var(--muted);
    font-size: 0.9rem;
  }
  .search-row {
    display: block;
    width: 100%;
    text-align: left;
    cursor: pointer;
    border: none;
    border-radius: var(--radius-md);
    background: transparent;
    color: var(--text);
    padding: 0.6rem 0.8rem;
  }
  .search-row:hover { background: var(--surface-hover); }
  .search-row.selected {
    background: var(--accent-soft);
    box-shadow: inset 2px 0 0 var(--accent);
  }
  .search-title {
    font-weight: 700;
    font-size: 0.95rem;
    line-height: 1.4;
  }
  .search-row.selected .search-title { color: var(--accent-strong); }
  .search-snippet {
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--text-muted);
    margin-top: 0.2rem;
  }
  .search-snippet :global(mark) {
    background: var(--accent-soft);
    color: var(--accent-strong);
    border-radius: 3px;
    padding: 0 2px;
  }
  .search-path {
    font-family: var(--font-mono);
    font-size: 0.72rem;
    color: var(--muted);
    margin-top: 0.25rem;
  }
  .search-footer {
    padding: 0.6rem 1rem;
    border-top: 1px solid var(--border);
    background: var(--code-bg);
    color: var(--text-muted);
    font-size: 0.75rem;
  }
</style>
