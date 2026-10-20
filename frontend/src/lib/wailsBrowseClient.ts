// BrowseClient over Wails bindings (no HTTP server): getTree/renderDoc
// hit the Go service from Phase 1 directly. Rendered HTML rewrites
// /raw?path= targets to the /local-resource asset handler (absolute
// paths, whitelist-checked like the main preview).
import { BrowseTree, BrowseRender, BrowseSource, BrowseLinks, BrowseSearch, BrowseDocuments, GetVersion, GetStyleCSS } from '../../bindings/marksafe/app';
import type {
  BrowseClient,
  BrowseDoc,
  BrowseHit,
  BrowseTheme,
  DocSource,
  LinkTarget,
  MlcDocument,
  TocEntry,
} from '../shared/browse/browseClient';

function joinAbs(rootDir: string, rel: string): string {
  const root = rootDir.replace(/\\/g, '/').replace(/\/+$/, '');
  return `${root}/${rel.replace(/^\/+/, '')}`;
}

export function toBrowseRel(rootDir: string, absPath: string): string | null {
  const root = rootDir.replace(/\\/g, '/').replace(/\/+$/, '');
  const abs = absPath.replace(/\\/g, '/');
  if (abs === root) return null;
  if (abs.startsWith(root + '/')) return abs.slice(root.length + 1);
  return null;
}

// The browse window boots faster than the main view and fires its first
// binding call before the native bridge is attached. Checking for the
// webview object is NOT enough (it exists before the runtime answers),
// so we retry a real binding call exactly like the main app init does.
// Otherwise the tree fetch dies silently (empty sidebar, working
// document) on cold start.
async function ensureReady(): Promise<void> {
  for (let i = 0; i < 100; i++) {
    try {
      await GetVersion();
      return;
    } catch {
      await new Promise((r) => setTimeout(r, 50));
    }
  }
}

export function createWailsClient(rootDir: string): BrowseClient {
  return {
    async getTree(): Promise<TocEntry[]> {
      await ensureReady();
      try {
        return (await BrowseTree(rootDir)) ?? [];
      } catch (err) {
        console.error('BrowseTree failed:', err);
        return [];
      }
    },

    async renderDoc(path: string, theme: BrowseTheme): Promise<BrowseDoc> {
      await ensureReady();
      const doc = await BrowseRender(
        rootDir,
        path,
        theme === 'light' ? 'github' : 'github-dark',
      );
      if (!doc) throw new Error('Empty response');
      // /raw?path=<rel> has no server here: point at the local-resource
      // handler with absolute paths instead (same as file links below).
      const html = (doc.html ?? '').replace(/\/raw\?path=([^"'\s>]+)/g, (_m, q) => {
        try {
          const rel = decodeURIComponent(String(q).replace(/\+/g, ' '));
          return '/local-resource?path=' + encodeURIComponent(joinAbs(rootDir, rel));
        } catch {
          return _m;
        }
      });
      return {
        title: doc.title ?? '',
        html,
        size: doc.size ?? 0,
        modified: doc.modified ?? '',
      };
    },

    async getSource(path: string): Promise<DocSource | null> {
      await ensureReady();
      const doc = await BrowseSource(rootDir, path);
      return doc ? (doc as DocSource) : null;
    },

    async resolveLinks(docPath: string, hrefs: string[]): Promise<Record<string, LinkTarget>> {
      if (!hrefs.length) return {};
      await ensureReady();
      return ((await BrowseLinks(rootDir, docPath, hrefs)) ?? {}) as Record<string, LinkTarget>;
    },

    // No /raw server here: the local-resource handler with absolute
    // paths (whitelist-checked like the main preview).
    rawUrl(path: string): string {
      return '/local-resource?path=' + encodeURIComponent(joinAbs(rootDir, path));
    },

    async search(q: string): Promise<BrowseHit[]> {
      await ensureReady();
      return (await BrowseSearch(rootDir, q)) ?? [];
    },

    async documents(): Promise<MlcDocument[]> {
      try {
        await ensureReady();
        return ((await BrowseDocuments(rootDir)) ?? []) as MlcDocument[];
      } catch {
        return [];
      }
    },

    async chromaCss(theme: BrowseTheme): Promise<string> {
      await ensureReady();
      return (await GetStyleCSS(theme === 'light' ? 'github' : 'github-dark')) ?? '';
    },
  };
}
