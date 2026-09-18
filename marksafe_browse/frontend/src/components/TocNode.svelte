<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  // Recursive tree node: renders files as buttons and directories as
  // collapsible <details> — at ANY depth (Sidebar only handled one level,
  // which left grandchildren as dead <strong> text).
  export let entry: any;
  export let activePath: string | null = null;
  export let depth: number = 0;

  const dispatch = createEventDispatcher();

  // A section opens automatically when it holds the active page, so
  // following a content link always reveals (and marks) its target.
  // Manual toggles win until navigation moves into this subtree again.
  // (Depth 0 stays open as before.)
  let userOpen: boolean | null = null;

  function subtreeHas(e: any, p: string | null): boolean {
    if (!p) return false;
    if (e.path === p) return true;
    return (e.children ?? []).some((c: any) => subtreeHas(c, p));
  }

  $: autoOpen = depth < 1 || subtreeHas(entry, activePath);
  $: if (autoOpen) userOpen = null;
  $: isOpen = userOpen ?? autoOpen;

  function onToggle(e: Event) {
    userOpen = (e.currentTarget as HTMLDetailsElement).open;
  }

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
    <details open={isOpen} on:toggle={onToggle}>
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
