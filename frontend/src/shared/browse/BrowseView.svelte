<script lang="ts">
  import { createEventDispatcher, onMount, tick } from 'svelte';
  import './browse.css';
  import { enhanceContent } from './enhance';
  import type { BrowseClient, BrowseTheme } from './browseClient';
  import Sidebar from './Sidebar.svelte';
  import Content from './Content.svelte';
  import ThemeToggle from './ThemeToggle.svelte';
  import SearchModal from './SearchModal.svelte';
  import ExternalLinkModal from './ExternalLinkModal.svelte';
  import DocInfoModal from './DocInfoModal.svelte';

  // Host-owned: data access, color scheme, URL sync.
  export let client: BrowseClient;
  export let theme: BrowseTheme = 'dark';
  export let syncUrl = true;
  // Start document when syncUrl is off (embedded mode has no URL bar).
  export let initialPath: string | null = null;

  const dispatch = createEventDispatcher();

  const MIN_WIDTH = 180;
  const MAX_WIDTH = 520;

  // State — plain variables, no stores needed
  let tabs: any[] = [];
  let htmlContent = '';
  let pageTitle = 'MarkSafe Browse';
  let isReady = false;
  let lastTheme: BrowseTheme | null = null;
  let currentPath: string | null = null;
  let searchOpen = false;
  let externalUrl: string | null = null;
  let infoOpen = false;
  let docSize = 0;
  let docModified = '';
  let docWords = 0;
  let docChars = 0;

  // Line density: compact (default, IDE-like), comfortable, spacious.
  const DENSITIES = [
    { id: 'compact', label: 'Kompakt', body: '1.42', code: '1.3', tree: '0.2rem' },
    { id: 'comfortable', label: 'Komfort', body: '1.6', code: '1.5', tree: '0.38rem' },
    { id: 'spacious', label: 'Weit', body: '1.85', code: '1.7', tree: '0.6rem' },
  ];
  let densityId = 'compact';

  function applyDensity(id: string) {
    const mode = DENSITIES.find((d) => d.id === id) ?? DENSITIES[0];
    densityId = mode.id;
    try {
      const root = document.documentElement;
      root.style.setProperty('--lh-body', mode.body);
      root.style.setProperty('--lh-code', mode.code);
      root.style.setProperty('--tree-py', mode.tree);
      localStorage.setItem('marksafe-line-height', mode.id);
    } catch {
      // ignore
    }
  }

  function cycleDensity() {
    const i = DENSITIES.findIndex((d) => d.id === densityId);
    applyDensity(DENSITIES[(i + 1) % DENSITIES.length].id);
  }

  $: densityLabel =
    (DENSITIES.find((d) => d.id === densityId) ?? DENSITIES[0]).label;

  // Plain-text stats for the info panel. NOTE: htmlContent must be
  // referenced lexically inside the $: block — Svelte does not track
  // reads hidden in called functions, which froze these at 0.
  function docText(html: string): string {
    return html
      .replace(/<[^>]*>/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
  }
  $: {
    const _t = docText(htmlContent);
    docWords = _t ? _t.split(' ').length : 0;
    docChars = _t.length;
  }

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
    if (!syncUrl) return null;
    try {
      return new URLSearchParams(window.location.search).get('path');
    } catch {
      return null;
    }
  }

  function pushUrl(path: string) {
    if (!syncUrl) return;
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
    // Theme arrives via prop (host-owned). Sidebar/tree first.
    tabs = await client.getTree();
    console.debug(`[browse] tree entries: ${tabs.length}`);

    setChroma(theme);

    try {
      const stored = localStorage.getItem('marksafe-line-height');
      if (stored) applyDensity(stored);
    } catch {
      // ignore
    }

    // Startup document: deep-link (?path=…) when the host syncs URLs,
    // host-provided initial path (embedded mode) or index.md fallback.
    await loadMarkdown(syncUrl ? (urlPath() ?? 'index.md') : (initialPath ?? 'index.md'), false);
    lastTheme = theme;

    isReady = true;
  });

  // Re-render with matching code colors whenever the host theme flips.
  // Guarded by lastTheme so startup (which loads explicitly above)
  // does not fetch twice.
  $: if (isReady && currentPath && theme !== lastTheme) {
    lastTheme = theme;
    setChroma(theme);
    loadMarkdown(currentPath, false);
  }

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
    const el = document.querySelector('.browse-view .doc') as HTMLElement | null;
    if (el) await enhanceContent(el, theme === 'dark');
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
      const doc = await client.renderDoc(path, theme, ctrl.signal);
      if (seq !== loadSeq) return; // superseded
      if (doc) {
        htmlContent = doc.html;
        pageTitle = doc.title;
        currentPath = path;
        docSize = doc.size ?? 0;
        docModified = doc.modified ?? '';
        if (push) pushUrl(path);
        console.debug(
          `[browse] rendered ${path} (${doc.html.length} chars, seq ${seq})`,
        );
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

  function handleExternal(event: CustomEvent<{ url: string }>) {
    externalUrl = event.detail.url;
  }

  // Opening itself is host-specific (window.open vs. app browser API).
  function confirmExternal(event: CustomEvent<{ url: string; reuse: boolean }>) {
    externalUrl = null;
    dispatch('openExternal', event.detail);
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

  // Theme ownership lives with the host: flip there, this view follows
  // via the theme prop (re-render handled by the reactive guard above).
  function handleThemeToggle() {
    dispatch('toggleTheme');
  }
</script>

<svelte:head>
  <title>{pageTitle} — MarkSafe Browse</title>
  <meta name="description" content={pageDescription} />
</svelte:head>

<div class="browse-view" data-theme={theme}>
<div class="app-container">
  {#if sidebarHidden}
    <button class="expand-btn" on:click={toggleSidebar} title="Verzeichnis einblenden"
      aria-label="Verzeichnis einblenden">
      <span class="expand-label">Verzeichnis</span>
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
      <div class="toolbar-right">
        <button
          class="nav-btn"
          on:click={cycleDensity}
          title="Zeilenabstand: {densityLabel} (klicken zum Wechseln)"
          aria-label="Zeilenabstand wechseln, aktuell {densityLabel}"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <line x1="3" y1="6" x2="15" y2="6" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <line x1="3" y1="18" x2="15" y2="18" />
          </svg>
          <span class="nav-label">{densityLabel}</span>
        </button>
        <button
          class="nav-btn"
          on:click={() => (infoOpen = true)}
          title="Dokumentinfo anzeigen"
          aria-label="Dokumentinfo anzeigen"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="16" x2="12" y2="12" />
            <line x1="12" y1="8" x2="12.01" y2="8" />
          </svg>
        </button>
        <ThemeToggle
          on:toggle={handleThemeToggle}
          theme={theme}
        />
      </div>
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
      {#key currentPath}
        <Content {htmlContent} on:open={handleSelectEntry} on:external={handleExternal} />
      {/key}

  {#if externalUrl}
    <ExternalLinkModal
      url={externalUrl}
      on:confirm={confirmExternal}
      on:cancel={() => (externalUrl = null)}
    />
  {/if}

  {#if infoOpen}
    <DocInfoModal
      title={pageTitle}
      path={currentPath ?? ''}
      size={docSize}
      modified={docModified}
      words={docWords}
      chars={docChars}
      on:close={() => (infoOpen = false)}
    />
  {/if}

  {#if searchOpen}
    <SearchModal
      client={client}
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
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text);
    font-size: 1rem;
  }
  .expand-btn:hover { background: var(--surface-hover); color: var(--accent); border-color: var(--accent); }
  .expand-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
  }
  .expand-label { font-size: 0.85rem; }
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
  .toolbar-right { display: flex; align-items: center; gap: 0.5rem; }
  .nav-btn .nav-label { font-size: 0.78rem; }
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
