<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
  
  function handleClick() {
    dispatch('toggle');
  }
</script>

<button class="theme-toggle" on:click={handleClick}>
  🌓
</button>

<style>
  .theme-toggle {
    position: absolute;
    top: 1rem;
    right: 1rem;
    cursor: pointer;
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text);
    font-size: 1.2rem;
  }
  .theme-toggle:hover { opacity: 0.8; }
</style>
