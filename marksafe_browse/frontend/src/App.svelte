<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchTree } from './lib/backend';
  import Sidebar from './components/Sidebar.svelte';
  import Content from './components/Content.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';

  // State — plain variables, no stores needed
  let tabs: any[] = [];
  let htmlContent = '';
  let pageTitle = 'MarkSafe Browse';
  let effectiveTheme = 'dark';
  let isReady = false;

  onMount(async () => {
    // Initialize theme from localStorage or system
    const stored = localStorage.getItem('marksafe-theme');
    if (stored) {
      document.body.className = stored;
      effectiveTheme = stored;
    } else {
      const system = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
      document.body.className = system;
      effectiveTheme = system;
    }

    // Fetch TOC tree
    try {
      const resp = await fetch('/tree');
      const data = await resp.json();
      tabs = data;
    } catch (e) {
      console.error('Failed to fetch tree:', e);
    }

    // Auto-load index.md if available
    try {
      const resp = await fetch('/render?path=index.md');
      const data = await resp.json();
      if (data) {
        htmlContent = data.html;
        pageTitle = data.title;
      }
    } catch (e) {
      console.error('Failed to load index.md:', e);
    }

    isReady = true;
  });

  async function loadMarkdown(path: string) {
    try {
      const resp = await fetch(`/render?path=${encodeURIComponent(path)}`);
      const data = await resp.json();
      if (data) {
        htmlContent = data.html;
        pageTitle = data.title;
      }
    } catch (e) {
      console.error('Failed to load markdown:', e);
    }
  }

  function handleSelectEntry(path: string) {
    loadMarkdown(path);
  }

  function handleThemeToggle() {
    const next = effectiveTheme === 'dark' ? 'light' : 'dark';
    document.body.className = next;
    localStorage.setItem('marksafe-theme', next);
    effectiveTheme = next;
  }
</script>

<div class="app-container">
  <Sidebar />
  
  <main class="content">
    <ThemeToggle 
      on:toggle={handleThemeToggle}
      theme={effectiveTheme}
    />
    
    {#if !isReady}
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
