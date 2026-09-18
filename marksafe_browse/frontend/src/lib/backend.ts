// API client for the marksafe_browse Go backend

interface TOCEntry {
  path: string;
  title: string;
  children: TOCEntry[];
  isDir: boolean;
}

interface RenderResponse {
  title: string;
  html: string;
  path: string;
}

export async function fetchTree(): Promise<TOCEntry[]> {
  try {
    const res = await fetch('/tree');
    if (!res.ok) throw new Error('Failed to fetch tree');
    return await res.json();
  } catch (err) {
    console.error('fetchTree failed:', err);
    return [];
  }
}

export async function renderMarkdown(path: string): Promise<RenderResponse | null> {
  try {
    const res = await fetch(`/render?path=${encodeURIComponent(path)}`);
    if (!res.ok) throw new Error('Failed to render');
    return await res.json();
  } catch (err) {
    console.error('renderMarkdown failed:', err);
    return null;
  }
}

export async function fetchMarkdown(path: string): Promise<string | null> {
  try {
    const res = await fetch(`/md/${encodeURIComponent(path)}`);
    if (!res.ok) throw new Error('Failed to fetch markdown');
    return await res.text();
  } catch (err) {
    console.error('fetchMarkdown failed:', err);
    return null;
  }
}

export function isBackendReady(): boolean {
  if (typeof window === 'undefined') return false;
  const w = window as any;
  return Boolean(w.chrome?.webview?.postMessage || w.webkit?.messageHandlers?.external);
}

/**
 * Toggle between dark and light theme.
 * Stores preference in localStorage.
 */
export function toggleTheme(): void {
  const body = document.body;
  const current = body.className;
  const next = current === 'dark' ? 'light' : 'dark';
  body.className = next;
  localStorage.setItem('marksafe-theme', next);
}

/**
 * Get stored theme from localStorage. Returns null if not set.
 */
export function getStoredTheme(): string | null {
  return localStorage.getItem('marksafe-theme');
}

/**
 * Get system theme preference from prefers-color-scheme.
 */
export function getSystemTheme(): string {
  if (typeof window === 'undefined') return 'light';
  if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark';
  return 'light';
}

/**
 * Apply a theme to the document body.
 * If theme is 'auto', detects system preference.
 */
export function applyTheme(theme: string): void {
  const resolved = theme === 'auto' ? getSystemTheme() : theme;
  document.body.className = resolved;
  if (theme !== 'auto') {
    localStorage.setItem('marksafe-theme', theme);
  } else {
    localStorage.removeItem('marksafe-theme');
  }
}

/**
 * Listen for system theme changes (auto mode).
 * Returns a cleanup function.
 */
export function onSystemThemeChange(callback: (theme: string) => void): () => void {
  if (typeof window === 'undefined') return () => {};
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handler = (e: MediaQueryListEvent) => {
    callback(e.matches ? 'dark' : 'light');
  };
  mediaQuery.addEventListener('change', handler);
  return () => mediaQuery.removeEventListener('change', handler);
}
