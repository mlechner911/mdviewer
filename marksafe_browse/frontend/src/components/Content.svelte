<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';

  export let htmlContent: string = '';

  const dispatch = createEventDispatcher();
  let docEl: HTMLElement | null = null;

  // In-app navigation: links the server validated (a[data-md]) open via
  // /render instead of a full page load (which would 404). Broken links
  // carry no href at all, external links behave natively.
  function onClick(e: MouseEvent) {
    const target = e.target as HTMLElement | null;
    const anchor = target?.closest?.('a[data-md]') as HTMLAnchorElement | null;
    if (!anchor) return;
    e.preventDefault();
    const path = anchor.getAttribute('data-md');
    if (path) dispatch('open', { path });
  }

  // Attached imperatively (not via on:click): the section itself is NOT an
  // interactive element — the handler only delegates to dynamic {@html}
  // links. This keeps a11y checkers (correctly) quiet.
  onMount(() => {
    docEl?.addEventListener('click', onClick);
    return () => docEl?.removeEventListener('click', onClick);
  });
</script>

<section class="doc" bind:this={docEl}>
  {#if htmlContent}
    {@html htmlContent}
  {:else}
    <h1>MarkSafe Browse</h1>
    <p>Wähle ein Dokument aus dem Inhaltsverzeichnis links.</p>
  {/if}
</section>

<style>
  /* NOTE: class intentionally NOT named "content" — that name carries a
     margin-left rule for the fixed sidebar; nesting it doubled the offset. */
  .doc {
    width: 100%;
    max-width: 896px;
    padding: 2rem;
  }
</style>
