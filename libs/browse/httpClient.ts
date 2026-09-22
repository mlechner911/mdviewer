import type { BrowseClient, BrowseDoc, BrowseHit, TocEntry, BrowseTheme } from './browseClient';

// BrowseClient over the marksafe_browse HTTP server
// (GET /tree, /render?path=&theme=, /api/search?q=).
export function createHttpClient(): BrowseClient {
  return {
    async getTree(): Promise<TocEntry[]> {
      try {
        const resp = await fetch('/tree');
        if (!resp.ok) return [];
        return (await resp.json()) ?? [];
      } catch (err) {
        console.error('Failed to fetch tree:', err);
        return [];
      }
    },

    async renderDoc(path: string, theme: BrowseTheme, signal?: AbortSignal): Promise<BrowseDoc> {
      const resp = await fetch(`/render?path=${encodeURIComponent(path)}&theme=${theme}`, {
        signal,
      });
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const data = await resp.json();
      if (!data) throw new Error('Empty response');
      return {
        title: data.title ?? '',
        html: data.html ?? '',
        size: typeof data.size === 'number' ? data.size : 0,
        modified: typeof data.modified === 'string' ? data.modified : '',
      };
    },

    async search(q: string): Promise<BrowseHit[]> {
      const resp = await fetch(`/api/search?q=${encodeURIComponent(q)}`);
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      return (await resp.json()) ?? [];
    },
  };
}
