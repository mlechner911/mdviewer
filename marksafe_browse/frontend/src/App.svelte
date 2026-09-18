<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import { fetchTree, toggleTheme, getStoredTheme } from './lib/backend';
  import Sidebar from './components/Sidebar.svelte';
  import Content from './components/Content.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';

  // State
  let tabs = writable([]);
  let htmlContent = writable('');
  let isReady = writable(false);
  let effectiveTheme = writable('dark');

  onMount(async () => {
    // Initialize theme
    const theme = getStoredTheme();
    document.body.className = theme;
    effectiveTheme.set(theme);

    // Fetch TOC tree
    const tree = await fetchTree();
    tabs.set(tree);

    isReady.set(true);
  });

  async function selectEntry(path: string) {
    const res = await fetch(`/render?path=${encodeURIComponent(path)}`);
    if (res.ok) {
      const data = await res.json();
      htmlContent.set(data.html);
    }
  }

  function handleTheme() {
    const current = document.body.className;
    const next = current === 'dark' ? 'light' : 'dark';
    document.body.className = next;
    localStorage.setItem('marksafe-theme', next);
    effectiveTheme.set(next);
  }
</script>

<div class="app-container">
  <Sidebar {tabs} onSelect={selectEntry} />
  
  <main class="content">
    <ThemeToggle on:toggle={handleTheme} />
    
    {#if !$isReady}
      <div class="loading">Lade Dokumentation...</div>
    {:else}
      <Content {htmlContent} />
    {/if}
  </main>
</div>

<style>
  .app-container {
    display: flex;
    min-height: 100vh;
  }
  .loading {
    text-align: center;
    padding: 4rem;
    font-size: 1.25rem;
    opacity: 0.6;
  }
</style>
