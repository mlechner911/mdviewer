<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar from './components/Sidebar.svelte';
  import Content from './components/Content.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';

  const MIN_WIDTH = 180;
  const MAX_WIDTH = 520;

  // State — plain variables, no stores needed
  let tabs: any[] = [];
  let htmlContent = '';
  let pageTitle = 'MarkSafe Browse';
  let effectiveTheme = 'dark';
  let isReady = false;

  // Sidebar layout state (persisted)
  let sidebarWidth = 280;
  let sidebarHidden = false;
  let dragging = false;

  try {
    const w = parseInt(localStorage.getItem('marksafe-sidebar-width') || '', 10);
    if (Number.isFinite(w)) sidebarWidth = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, w));
    sidebarHidden = localStorage.getItem('marksafe-sidebar-hidden') === '1';
  } catch {
    // localStorage unavailable — fall back to defaults
  }

  function startResize(e: PointerEvent) {
    dragging = true;
    (e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId);
    e.preventDefault();
  }

  function onPointerMove(e: PointerEvent) {
    if (!dragging) return;
    sidebarWidth = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, e.clientX));
  }

  function stopResize() {
    if (!dragging) return;
    dragging = false;
    try {
      localStorage.setItem('marksafe-sidebar-width', String(sidebarWidth));
    } catch {
      // ignore
    }
  }

  function toggleSidebar() {
    sidebarHidden = !sidebarHidden;
    try {
      localStorage.setItem('marksafe-sidebar-hidden', sidebarHidden ? '1' : '0');
    } catch {
      // ignore
    }
  }

  onMount(() => {
    window.addEventListener('pointermove', onPointerMove);
    window.addEventListener('pointerup', stopResize);
    window.addEventListener('pointercancel', stopResize);
    return () => {
      window.removeEventListener('pointermove', onPointerMove);
      window.removeEventListener('pointerup', stopResize);
      window.removeEventListener('pointercancel', stopResize);
    };
  });

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

  function handleSelectEntry(event: CustomEvent<{ path: string }>) {
    loadMarkdown(event.detail.path);
  }

  function handleThemeToggle() {
    const next = effectiveTheme === 'dark' ? 'light' : 'dark';
    document.body.className = next;
    localStorage.setItem('marksafe-theme', next);
    effectiveTheme = next;
  }
</script>

<svelte:head>
  <title>{pageTitle} — MarkSafe Browse</title>
</svelte:head>

<div class="app-container">
  {#if sidebarHidden}
    <button class="expand-btn" on:click={toggleSidebar} title="Verzeichnis einblenden"
      aria-label="Verzeichnis einblenden">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <rect x="3" y="3" width="18" height="18" rx="2" />
        <line x1="9" y1="3" x2="9" y2="21" />
      </svg>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <polyline points="9 18 15 12 9 6" />
      </svg>
    </button>
  {:else}
    <Sidebar
      {tabs}
      width={sidebarWidth}
      on:select={handleSelectEntry}
      on:collapse={toggleSidebar}
    />
    <div
      class="resizer"
      class:active={dragging}
      style="left: {sidebarWidth}px"
      on:pointerdown={startResize}
      on:dblclick={toggleSidebar}
      title="Ziehen zum Anpassen · Doppelklick zum Ausblenden"
    ></div>
  {/if}

  <main class="content" style={sidebarHidden ? 'margin-left: 0;' : `margin-left: ${sidebarWidth + 20}px;`}>
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
  .resizer {
    position: fixed;
    top: 0;
    bottom: 0;
    width: 10px;
    margin-left: -5px;
    cursor: col-resize;
    z-index: 10;
    touch-action: none;
    background: transparent;
  }
  /* Always-visible grip line so the zone is discoverable */
  .resizer::after {
    content: "";
    position: absolute;
    left: 50%;
    top: 8px;
    bottom: 8px;
    width: 2px;
    margin-left: -1px;
    border-radius: 2px;
    background: var(--border, #334155);
    opacity: 0.8;
    transition: background 0.15s, width 0.15s;
  }
  .resizer:hover::after,
  .resizer.active::after {
    background: var(--accent, #60a5fa);
    opacity: 1;
    width: 4px;
    margin-left: -2px;
  }
  .expand-btn {
    position: fixed;
    top: 1rem;
    left: 1rem;
    z-index: 10;
    cursor: pointer;
    padding: 0.5rem 0.75rem;
    border-radius: 0.5rem;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text);
    font-size: 1rem;
  }
  .expand-btn:hover { background: var(--surface-hover); color: var(--accent); border-color: var(--accent); }
  .expand-btn svg { width: 16px; height: 16px; }
</style>
