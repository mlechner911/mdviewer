<script lang="ts">
  export let htmlContent: string = '';
</script>

<section class="doc">
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
    max-width: 768px;
    padding: 2rem;
  }
</style>
