<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let tabs: any[] = [];
  export let width: number = 280;
  export let activePath: string | null = null;

  const dispatch = createEventDispatcher();

  function selectEntry(entry: any) {
    if (!entry.isDir) {
      dispatch('select', { path: entry.path });
    }
  }
</script>

<nav class="sidebar" style="width: {width}px" aria-label="Inhaltsverzeichnis">
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
      {#if entry.isDir}
        <li class="toc-dir">
          <details open>
            <summary>{entry.title}</summary>
            <ul class="toc sub">
              {#each entry.children as child}
                <li class="toc-entry">
                  {#if child.isDir}
                    <strong>{child.title}</strong>
                  {:else}
                    <button
                      class:active={child.path === activePath}
                      on:click={() => selectEntry(child)}
                      aria-current={child.path === activePath ? 'page' : undefined}
                    >{child.title}</button>
                  {/if}
                </li>
              {/each}
            </ul>
          </details>
        </li>
      {:else}
        <li class="toc-entry">
          <button
            class:active={entry.path === activePath}
            on:click={() => selectEntry(entry)}
            aria-current={entry.path === activePath ? 'page' : undefined}
          >{entry.title}</button>
        </li>
      {/if}
    {/each}
  </ul>
</nav>

<style>
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    height: 100vh;
    background: var(--surface);
    border-right: 1px solid var(--border);
    padding: 1.25rem 1rem;
    overflow-y: auto;
  }
  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    margin-bottom: 1rem;
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
  .toc {
    list-style: none;
    padding-left: 0;
  }
  .toc-entry { margin: 1px 0; }
  .toc-dir {
    font-family: var(--font-mono);
    font-size: 0.8rem;
    font-weight: 600;
    letter-spacing: 0.02em;
    text-transform: uppercase;
    color: var(--muted);
    margin: 0.75rem 0 0.25rem;
  }
  .toc-entry button {
    color: var(--text);
    cursor: pointer;
    display: block;
    width: 100%;
    padding: 0.4rem 0.6rem;
    border: none;
    border-radius: var(--radius-sm);
    background: transparent;
    text-align: left;
    font-family: var(--font-sans);
    font-size: 0.9rem;
    line-height: 1.45;
  }
  .toc-entry button:hover { background: var(--surface-hover); color: var(--accent-strong); }
  /* Active page: accent text + tinted bg + fine vertical bar */
  .toc-entry button.active {
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-weight: 600;
    box-shadow: inset 2px 0 0 var(--accent);
  }
  details { list-style: none; }
  details summary {
    cursor: pointer;
    list-style: none;
    list-style-type: none;
    display: flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.3rem 0.4rem;
    border-radius: var(--radius-sm);
    user-select: none;
  }
  details summary::-webkit-details-marker { display: none; }
  details summary::marker { content: ''; font-size: 0; }
  details summary::before {
    content: '';
    flex: none;
    width: 0;
    height: 0;
    border-top: 4px solid transparent;
    border-bottom: 4px solid transparent;
    border-left: 6px solid var(--muted);
    transition: transform 0.15s ease;
  }
  details[open] > summary::before { transform: rotate(90deg); }
  details summary:hover { background: var(--surface-hover); }
  .sub { padding-left: 1.1rem; list-style: none; }
  .sidebar::-webkit-scrollbar { width: 8px; }
  .sidebar::-webkit-scrollbar-thumb { background: var(--border); border-radius: 4px; }
</style>
