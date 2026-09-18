<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  export let theme: string = 'dark';
  const dispatch = createEventDispatcher();
  
  function handleClick() {
    dispatch('toggle');
  }
</script>

<button 
  class="theme-toggle" 
  on:click={handleClick}
  title="Theme: {$theme}"
>
  {#if $theme === 'dark'}🌙 Dark{:else if $theme === 'light'}☀️ Light{:else}🌓 Auto{/if}
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
    font-size: 1rem;
    transition: opacity 0.2s, background 0.2s;
  }
  .theme-toggle:hover { opacity: 0.8; background: var(--bg); }
</style>
