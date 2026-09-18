<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let htmlContent: string = '';

  const dispatch = createEventDispatcher();

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
</script>

<section class="doc" on:click={onClick}>
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
