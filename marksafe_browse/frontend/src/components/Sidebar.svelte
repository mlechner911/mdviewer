<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  export let tabs: any[] = [];
  
  const dispatch = createEventDispatcher();
  
  function selectEntry(path: string) {
    dispatch('select', { path });
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
                    <a href="javascript:void(0)" on:click={() => selectEntry(child.path)}>{child.title}</a>
                  {/if}
                </li>
              {/each}
            </ul>
          </details>
        </li>
      {:else}
        <li class="toc-entry">
          <a href="javascript:void(0)" on:click={() => selectEntry(entry.path)}>{entry.title}</a>
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
  .toc-dir { font-weight: bold; color: var(--accent); }
  .toc-entry a { color: var(--accent); text-decoration: none; cursor: pointer; }
  .toc-entry a:hover { text-decoration: underline; }
  details summary { cursor: pointer; }
  .sub { padding-left: 1rem; list-style: none; }
</style>
