<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { enhanceContent } from './lib/backend';
  import Sidebar from './components/Sidebar.svelte';
  import Content from './components/Content.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';
  import SearchModal from './components/SearchModal.svelte';

  const MIN_WIDTH = 180;
  const MAX_WIDTH = 520;

  // State — plain variables, no stores needed
  let tabs: any[] = [];
  let htmlContent = '';
  let pageTitle = 'MarkSafe Browse';
  let effectiveTheme = 'dark';
  let isReady = false;
  let currentPath: string | null = null;
  let searchOpen = false;

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

  function persistWidth() {
    try {
      localStorage.setItem('marksafe-sidebar-width', String(sidebarWidth));
    } catch {
      // ignore
    }
  }

  function setWidth(w: number) {
    sidebarWidth = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, w));
    persistWidth();
  }

  function stopResize() {
    if (!dragging) return;
    dragging = false;
    persistWidth();
  }

  // Keyboard resizing for the separator (Arrow keys, Home/End).
  function onResizerKey(e: KeyboardEvent) {
    const step = e.shiftKey ? 50 : 10;
    if (e.key === 'ArrowLeft') setWidth(sidebarWidth - step);
    else if (e.key === 'ArrowRight') setWidth(sidebarWidth + step);
    else if (e.key === 'Home') setWidth(MIN_WIDTH);
    else if (e.key === 'End') setWidth(MAX_WIDTH);
    else return;
    e.preventDefault();
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
    window.addEventListener('popstate', onPopState);
    window.addEventListener('keydown', onGlobalKey);
    return () => {
      window.removeEventListener('pointermove', onPointerMove);
      window.removeEventListener('pointerup', stopResize);
      window.removeEventListener('pointercancel', stopResize);
      window.removeEventListener('popstate', onPopState);
      window.removeEventListener('keydown', onGlobalKey);
    };
  });

  function urlPath(): string | null {
    try {
      return new URLSearchParams(window.location.search).get('path');
    } catch {
      return null;
    }
  }

  function pushUrl(path: string) {
    try {
      history.pushState({ path }, '', '?path=' + encodeURIComponent(path));
    } catch {
      // ignore (e.g. file:// or sandboxed iframe)
    }
  }

  // Browser back/forward buttons: reload the historic document.
  function onPopState(e: PopStateEvent) {
    const p = (e.state as { path?: string } | null)?.path ?? urlPath() ?? 'index.md';
    loadMarkdown(p, false);
  }

  // Ctrl/Cmd+K opens the search; Alt+Left / Alt+Right = back / forward.
  function onGlobalKey(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
      e.preventDefault();
      searchOpen = true;
      return;
    }
    if (!e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;
    const t = e.target as HTMLElement | null;
    if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)) return;
    if (e.key === 'ArrowLeft') {
      e.preventDefault();
      history.back();
    } else if (e.key === 'ArrowRight') {
      e.preventDefault();
      history.forward();
    }
  }

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

    setChroma(effectiveTheme);

    // Startup document: deep-link (?path=…) or index.md fallback
    await loadMarkdown(urlPath() ?? 'index.md', false);

    isReady = true;
  });

  function setChroma(theme: string) {
    const link = document.getElementById('chroma-css') as HTMLLinkElement | null;
    if (link) {
      link.href = theme === 'light' ? '/assets/chroma-light.css' : '/assets/chroma-dark.css';
    }
  }

  // Re-run KaTeX + Mermaid on the rendered document (same behavior as
  // the Wails preview; markup comes from the shared Go renderer).
  async function refreshEnhancements() {
    await tick();
    const el = document.querySelector('#app .doc') as HTMLElement | null;
    if (el) await enhanceContent(el, effectiveTheme === 'dark');
  }

  // Sequenced loader: overlapping navigations are normal (fast clicks,
  // back/forward, search). Only the latest request may touch state —
  // stale responses are dropped silently instead of blanking the view.
  // Failures become a visible retry panel, never a silent blank page.
  let loadSeq = 0;
  let loadAbort: AbortController | null = null;
  let loadError: string | null = null;

  async function loadMarkdown(path: string, push = true) {
    const seq = ++loadSeq;
    loadAbort?.abort();
    const ctrl = new AbortController();
    loadAbort = ctrl;
    loadError = null;
    try {
      const resp = await fetch(
        `/render?path=${encodeURIComponent(path)}&theme=${effectiveTheme}`,
        { signal: ctrl.signal },
      );
      if (seq !== loadSeq) return; // superseded
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const data = await resp.json();
      if (seq !== loadSeq) return; // superseded
      if (data) {
        htmlContent = data.html;
        pageTitle = data.title;
        currentPath = path;
        if (push) pushUrl(path);
        await refreshEnhancements();
      }
    } catch (e) {
      const aborted = e instanceof DOMException && e.name === 'AbortError';
      if (aborted || seq !== loadSeq) return; // superseded, silent
      const msg = e instanceof Error ? e.message : String(e);
      console.error(`Failed to load ${path}:`, e);
      loadError = `"${path}" konnte nicht geladen werden (${msg}).`;
    }
  }

  function retryLoad() {
    loadMarkdown(currentPath ?? 'index.md', false);
  }

  function handleSelectEntry(event: CustomEvent<{ path: string }>) {
    loadMarkdown(event.detail.path);
  }

  // SEO: description follows the open document (first ~160 chars of
  // visible text). Svelte updates the <meta> tag reactively, so the
  // Lighthouse "meta description" audit passes on every page.
  function excerpt(html: string): string {
    const text = html
      .replace(/<[^>]*>/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
    if (!text) return 'MarkSafe Browse – Markdown-Verzeichnis im Browser lesen.';
    return text.length > 160 ? text.slice(0, 157).trimEnd() + '…' : text;
  }
  $: pageDescription = excerpt(htmlContent);

  async function handleThemeToggle() {
    const next = effectiveTheme === 'dark' ? 'light' : 'dark';
    document.body.className = next;
    localStorage.setItem('marksafe-theme', next);
    effectiveTheme = next;
    setChroma(next);
    // Re-render so code colors match the theme, then enhance diagrams/math.
    if (currentPath) await loadMarkdown(currentPath, false);
  }
</script>

<svelte:head>
  <title>{pageTitle} — MarkSafe Browse</title>
  <meta name="description" content={pageDescription} />
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
      activePath={currentPath}
      on:select={handleSelectEntry}
      on:collapse={toggleSidebar}
    />
    <div
      class="resizer"
      class:active={dragging}
      style="left: {sidebarWidth}px"
      role="slider"
      aria-orientation="vertical"
      aria-label="Seitenleistenbreite"
      aria-valuemin={MIN_WIDTH}
      aria-valuemax={MAX_WIDTH}
      aria-valuenow={sidebarWidth}
      tabindex="0"
      on:pointerdown={startResize}
      on:dblclick={toggleSidebar}
      on:keydown={onResizerKey}
      title="Ziehen oder Pfeiltasten zum Anpassen · Doppelklick zum Ausblenden"
    ></div>
  {/if}

  <main class="content" style={sidebarHidden ? 'margin-left: 0;' : `margin-left: ${sidebarWidth + 20}px;`}>
    <div class="content-toolbar">
      <div class="nav-btns" role="group" aria-label="Navigation">
        <button class="nav-btn" on:click={() => (searchOpen = true)} title="Suchen (Strg+K)" aria-label="Suchen">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </button>
        <button class="nav-btn" on:click={() => history.back()} title="Zurück (Alt+←)" aria-label="Zurück">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <polyline points="15 18 9 12 15 6" />
          </svg>
        </button>
        <button class="nav-btn" on:click={() => history.forward()} title="Vor (Alt+→)" aria-label="Vor">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </button>
      </div>
      <ThemeToggle
        on:toggle={handleThemeToggle}
        theme={effectiveTheme}
      />
    </div>

    {#if !isReady}
      <div class="loading">Lade Dokumentation...</div>
    {:else if loadError}
      <div class="load-error" role="alert">
        <strong>Dokument konnte nicht geladen werden.</strong>
        <p>{loadError}</p>
        <button class="retry-btn" on:click={retryLoad}>Erneut versuchen</button>
      </div>
    {:else}
      <Content {htmlContent} on:open={handleSelectEntry} />

  {#if searchOpen}
    <SearchModal
      on:open={(e) => {
        searchOpen = false;
        handleSelectEntry(e);
      }}
      on:close={() => (searchOpen = false)}
    />
  {/if}
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
  .load-error {
    max-width: 560px;
    margin: 3rem auto;
    padding: 1.5rem;
    border: 1px solid var(--border);
    border-left: 3px solid var(--accent);
    border-radius: var(--radius-md);
    background: var(--surface);
  }
  .load-error p { margin: 0.5rem 0 1rem; color: var(--muted); }
  .retry-btn {
    cursor: pointer;
    padding: 0.5rem 1rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--accent);
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-family: var(--font-sans);
    font-size: 0.9rem;
    font-weight: 600;
  }
  .retry-btn:hover { background: var(--accent); color: #fff; }
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
  /* Sticky toolbar: theme toggle lives here, never overlapping text */
  .content-toolbar {
    position: sticky;
    top: 0;
    z-index: 5;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1.5rem;
    background: var(--bg);
    border-bottom: 1px solid var(--border);
  }
  .nav-btns { display: flex; gap: 0.5rem; }
  .nav-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    padding: 0.45rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: var(--surface);
    color: var(--text);
  }
  .nav-btn:hover { border-color: var(--accent); color: var(--accent); }
  .nav-btn svg { width: 16px; height: 16px; }

  .expand-btn svg { width: 16px; height: 16px; }
</style>
