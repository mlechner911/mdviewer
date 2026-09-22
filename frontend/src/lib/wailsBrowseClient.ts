// BrowseClient over Wails bindings (no HTTP server): getTree/renderDoc
// hit the Go service from Phase 1 directly. Rendered HTML rewrites
// /raw?path= targets to the /local-resource asset handler (absolute
// paths, whitelist-checked like the main preview).
import { BrowseTree, BrowseRender, BrowseSearch } from '../../bindings/marksafe/app';
import type {
  BrowseClient,
  BrowseDoc,
  BrowseHit,
  BrowseTheme,
  TocEntry,
} from '../../../libs/browse/browseClient';

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

export function createWailsClient(rootDir: string): BrowseClient {
  return {
    async getTree(): Promise<TocEntry[]> {
      try {
        return (await BrowseTree(rootDir)) ?? [];
      } catch (err) {
        console.error('BrowseTree failed:', err);
        return [];
      }
    },

    async renderDoc(path: string, theme: BrowseTheme): Promise<BrowseDoc> {
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

    async search(q: string): Promise<BrowseHit[]> {
      return (await BrowseSearch(rootDir, q)) ?? [];
    },
  };
}
