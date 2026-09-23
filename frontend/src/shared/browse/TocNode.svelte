<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import TocNode from './TocNode.svelte';

  // Recursive tree node: renders files as buttons and directories as
  // collapsible <details> — at ANY depth (Sidebar only handled one level,
  // which left grandchildren as dead <strong> text).
  let {
    entry,
    activePath = null,
    depth = 0,
  }: {
    entry: any;
    activePath?: string | null;
    depth?: number;
  } = $props();

  const dispatch = createEventDispatcher();

  // A section opens automatically when it holds the active page, so
  // following a content link always reveals (and marks) its target.
  // bind:open keeps manual toggles in sync; navigation only ever forces
  // closed sections open, never closes anything behind the user's back.
  // (Depth 0 stays open as before.)
  let open = $state(false);

  function subtreeHas(e: any, p: string | null): boolean {
    if (!p) return false;
    if (e.path === p) return true;
    return (e.children ?? []).some((c: any) => subtreeHas(c, p));
  }

  let wantOpen = $derived(depth < 1 || subtreeHas(entry, activePath));
  $effect(() => {
    if (wantOpen) open = true;
  });

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
    <details bind:open>
      <summary title={entry.path}>{entry.title}</summary>
      <ul class="toc sub">
        {#each entry.children ?? [] as child}
          <TocNode entry={child} {activePath} depth={depth + 1} on:select={forward} />
        {/each}
      </ul>
    </details>
  </li>
{:else}
  <li class="toc-entry">
    <button
      class:active={entry.path === activePath}
      title={entry.path}
      onclick={() => select(entry)}
      aria-current={entry.path === activePath ? 'page' : undefined}
    >{entry.title}</button>
  </li>
{/if}
