<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import { fetchTree, getStoredTheme } from './lib/backend';
  import Sidebar from './components/Sidebar.svelte';
  import Content from './components/Content.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';

  // State
  let tabs = writable([]);
  let htmlContent = writable('');
  let isReady = writable(false);
  let pageTitle = writable('MarkSafe Browse');

  onMount(async () => {
    // Initialize theme
    const theme = getStoredTheme();
    document.body.className = theme;

    // Fetch TOC tree
    const tree = await fetchTree();
    tabs.set(tree);

    // Auto-load index.md if available
    const rootEntry = tree.find(e => e.path === 'index.md');
    if (rootEntry) {
      const result = await fetch(`/render?path=index.md`).then(r => r.json());
      htmlContent.set(result.html);
      pageTitle.set(result.title);
    }

    isReady.set(true);
  });

  async function loadMarkdown(path: string) {
    const result = await fetch(`/render?path=${encodeURIComponent(path)}`).then(r => r.json());
    if (result) {
      htmlContent.set(result.html);
      pageTitle.set(result.title);
    }
  }

  function handleSelectEntry(path: string) {
    loadMarkdown(path);
  }

  function handleThemeToggle() {
    const current = document.body.className;
    const next = current === 'dark' ? 'light' : 'dark';
    document.body.className = next;
    localStorage.setItem('marksafe-theme', next);
  }
</script>

<div class="app-container">
  <Sidebar {tabs} onSelect={handleSelectEntry} />
  
  <main class="content">
    <ThemeToggle on:toggle={handleThemeToggle} />
    
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
