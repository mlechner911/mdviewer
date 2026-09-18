<script lang="ts">
  export let htmlContent: string = '';
</script>

<section class="content">
  {#if htmlContent}
    {@html htmlContent}
  {:else}
    <h1>MarkSafe Browse</h1>
    <p>Wähle ein Dokument aus dem Inhaltsverzeichnis links.</p>
  {/if}
</section>

<style>
  .content {
    margin-left: 300px;
    padding: 2rem;
    max-width: 900px;
    flex: 1;
  }
</style>
