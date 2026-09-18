<script lang="ts">
  import { onMount } from 'svelte';
  import { writable } from 'svelte/store';
  import { fetchTree, getStoredTheme, getSystemTheme, applyTheme, onSystemThemeChange } from './lib/backend';
  import Sidebar from './components/Sidebar.svelte';
  import Content from './components/Content.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';

  // State
  let tabs = writable([]);
  let htmlContent = writable('');
  let isReady = writable(false);
  let pageTitle = writable('MarkSafe Browse');
  let effectiveTheme = writable('dark');

  onMount(async () => {
    // Initialize theme: check localStorage first, then system preference
    const stored = getStoredTheme();
    if (stored) {
      applyTheme(stored);
      effectiveTheme.set(stored);
    } else {
      const system = getSystemTheme();
      applyTheme('auto');
      effectiveTheme.set(system);
    }

    // Listen for system theme changes when in auto mode
    const cleanup = onSystemThemeChange((theme) => {
      if (!getStoredTheme()) {
        effectiveTheme.set(theme);
        applyTheme('auto');
      }
    });

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
    applyTheme(next);
    effectiveTheme.set(next);
  }

  function handleThemeChange(theme: string) {
    effectiveTheme.set(theme);
  }
</script>

<div class="app-container">
  <Sidebar {tabs} onSelect={handleSelectEntry} />
  
  <main class="content">
    <ThemeToggle 
      on:toggle={handleThemeToggle}
      theme={$effectiveTheme}
    />
    
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
