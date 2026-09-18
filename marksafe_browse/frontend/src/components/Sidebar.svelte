<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  export let tabs: any[] = [];
  
  const dispatch = createEventDispatcher();
  
  function selectEntry(entry: any) {
    if (!entry.isDir) {
      dispatch('select', { path: entry.path });
    }
  }
</script>

<nav class="sidebar">
  <h2>📚 Inhaltsverzeichnis</h2>
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
                    <button on:click={() => selectEntry(child)}>{child.title}</button>
                  {/if}
                </li>
              {/each}
            </ul>
          </details>
        </li>
      {:else}
        <li class="toc-entry">
          <button on:click={() => selectEntry(entry)}>{entry.title}</button>
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
    width: 280px;
    height: 100vh;
    background: inherit;
    border-right: 1px solid var(--border);
    padding: 1.5rem;
    overflow-y: auto;
  }
  .toc {
    list-style: none;
    padding-left: 0;
  }
  .toc-entry { margin: 0.25rem 0; }
  .toc-dir { font-weight: bold; color: var(--accent); margin-bottom: 0.5rem; }
  .toc-entry button {
    color: var(--accent);
    text-decoration: none;
    cursor: pointer;
    display: block;
    width: 100%;
    padding: 0.25rem 0.5rem;
    border: none;
    background: transparent;
    text-align: left;
    font-size: inherit;
  }
  .toc-entry button:hover {
    text-decoration: underline;
    background: var(--bg);
  }
  details summary { cursor: pointer; }
  .sub { padding-left: 1rem; list-style: none; }
</style>
