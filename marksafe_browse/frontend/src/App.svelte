<script lang="ts">
  import { onMount } from 'svelte';
  import BrowseView from '../../../libs/browse/BrowseView.svelte';
  import { createHttpClient } from '../../../libs/browse/httpClient';
  import type { BrowseTheme } from '../../../libs/browse/browseClient';

  // Thin host shell: theme ownership + external navigation live here,
  // everything else is the shared BrowseView (also used embedded).
  const client = createHttpClient();
  let theme: BrowseTheme = 'dark';

  function systemTheme(): BrowseTheme {
    try {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    } catch {
      return 'dark';
    }
  }

  function syncBodyBg(t: BrowseTheme) {
    try {
      document.body.style.background = t === 'light' ? '#ffffff' : '#0d0e11';
    } catch {
      // ignore
    }
  }

  function initTheme() {
    let stored: string | null = null;
    try {
      stored = localStorage.getItem('marksafe-theme');
    } catch {
      // ignore
    }
    theme = stored === 'light' || stored === 'dark' ? stored : systemTheme();
    syncBodyBg(theme);
  }

  function handleToggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark';
    try {
      localStorage.setItem('marksafe-theme', theme);
    } catch {
      // ignore
    }
    syncBodyBg(theme);
  }

  function handleOpenExternal(event: CustomEvent<{ url: string; reuse: boolean }>) {
    const { url, reuse } = event.detail;
    try {
      // Named window reuses one tab; _blank opens a fresh one. Opener is
      // always severed (backend also sets rel=noopener).
      window.open(url, reuse ? 'marksafe-external' : '_blank', 'noopener');
    } catch (e) {
      console.error(`Failed to open ${url}:`, e);
    }
  }

  onMount(initTheme);
</script>

<BrowseView
  {client}
  {theme}
  syncUrl
  on:toggleTheme={handleToggleTheme}
  on:openExternal={handleOpenExternal}
/>
