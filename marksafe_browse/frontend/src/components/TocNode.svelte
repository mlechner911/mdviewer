<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  // Recursive tree node: renders files as buttons and directories as
  // collapsible <details> — at ANY depth (Sidebar only handled one level,
  // which left grandchildren as dead <strong> text).
  export let entry: any;
  export let activePath: string | null = null;
  export let depth: number = 0;

  const dispatch = createEventDispatcher();

  function select(e: any) {
    if (!e.isDir && e.path) dispatch('select', { path: e.path });
  }

  // Re-emit child selections so they bubble to Sidebar -> App.
  function forward(e: CustomEvent<{ path: string }>) {
    dispatch('select', e.detail);
  }
</script>

{#if entry.isDir}
  <li class="toc-dir">
    <details open={depth < 1}>
      <summary>{entry.title}</summary>
      <ul class="toc sub">
        {#each entry.children ?? [] as child}
          <svelte:self entry={child} {activePath} depth={depth + 1} on:select={forward} />
        {/each}
      </ul>
    </details>
  </li>
{:else}
  <li class="toc-entry">
    <button
      class:active={entry.path === activePath}
      on:click={() => select(entry)}
      aria-current={entry.path === activePath ? 'page' : undefined}
    >{entry.title}</button>
  </li>
{/if}
