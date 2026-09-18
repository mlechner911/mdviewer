<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';

  export let htmlContent: string = '';

  const dispatch = createEventDispatcher();
  let docEl: HTMLElement | null = null;

  // Visible-text check: an "empty" render (frontmatter-only file,
  // tag-only fragment) shows the friendly placeholder instead of a
  // blank section that looks like a broken renderer.
  function hasVisibleText(html: string): boolean {
    if (!html) return false;
    return (
      html
        .replace(/<[^>]*>/g, '')
        .replace(/\s+/g, ' ')
        .trim().length > 0
    );
  }
  $: showDoc = hasVisibleText(htmlContent);

  // In-app navigation: links the server validated (a[data-md]) open via
  // /render instead of a full page load (which would 404). Broken links
  // carry no href at all, external links behave natively.
  function onClick(e: MouseEvent) {
    const target = e.target as HTMLElement | null;
    // External links always confirm first (see ExternalLinkModal).
    const ext = target?.closest?.(
      'a.external-link[href],a[href^="http://"],a[href^="https://"]',
    ) as HTMLAnchorElement | null;
    if (ext && ext.href) {
      e.preventDefault();
      dispatch('external', { url: ext.href });
      return;
    }
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
  {#if showDoc}
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
