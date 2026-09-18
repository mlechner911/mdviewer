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

export function toggleTheme(): void {
  const body = document.body;
  const current = body.className;
  const next = current === 'dark' ? 'light' : 'dark';
  body.className = next;
  localStorage.setItem('marksafe-theme', next);
}

export function getStoredTheme(): string {
  const stored = localStorage.getItem('marksafe-theme');
  if (stored) return stored;
  if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark';
  return 'light';
}

export function applyTheme(theme: string): void {
  document.body.className = theme;
  localStorage.setItem('marksafe-theme', theme);
}
