// What the app adds to rendered Markdown, for every place that shows a
// document: images and links pass the whitelist, code is highlighted, and
// diagrams and formulas load on first use.
import hljs from 'highlight.js/lib/common';
import * as backend from './backend';
import { BrowserOpenURL } from './wails';

// Images are not requested before the whitelist check: the renderer gets a
// transparent placeholder carrying the real URL, checked after rendering.
export const IMAGE_PLACEHOLDER = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7';
const MARK = '#mlc-img=';

export function resolveUrl(url: string, kind: 'link' | 'image'): string {
  if (kind === 'image' && url && !url.startsWith('data:')) return IMAGE_PLACEHOLDER + MARK + encodeURIComponent(url);
  return url;
}

export function highlight(code: string, lang: string): string | undefined {
  if (!hljs.getLanguage(lang)) return undefined;
  return hljs.highlight(code, { language: lang, ignoreIllegals: true }).value;
}

export const loadKatex = async () => (await import('katex')).default;

export function createMdHooks(opt: {
  /** The document's file: relative images and links start from its folder. */
  filePath: () => string | null | undefined;
  mermaid: () => { theme: string; themeVariables?: Record<string, unknown> };
  onsecurity?: (detail: { type: 'path' | 'url'; resource: string }) => void;
  onopenfile?: (detail: { path: string }) => void;
}) {
  async function baseDir(): Promise<string> {
    const path = opt.filePath();
    return path ? await backend.getParentDir(path) : '';
  }

  async function checkImages(root: HTMLElement) {
    for (const img of Array.from(root.querySelectorAll('img'))) {
      const src = img.getAttribute('src') ?? '';
      const at = src.indexOf(MARK);
      if (!src.startsWith(IMAGE_PLACEHOLDER) || at < 0) continue;
      const url = decodeURIComponent(src.slice(at + MARK.length));
      img.dataset.src = url;
      if (/^https?:\/\//i.test(url)) {
        try {
          const domain = new URL(url).hostname;
          if (await backend.isURLAllowed(domain)) img.src = url;
          else {
            img.style.display = 'none';
            opt.onsecurity?.({ type: 'url', resource: domain });
          }
        } catch {
          img.style.display = 'none';
        }
        continue;
      }
      const abs = await backend.resolveRelativePath(await baseDir(), url);
      if (await backend.isPathAllowed(abs)) {
        img.src = '/local-resource?path=' + encodeURIComponent(abs); // the handler checks again
      } else {
        img.style.display = 'none';
        opt.onsecurity?.({ type: 'path', resource: await backend.getParentDir(abs) });
      }
    }
  }

  function onNavigate({ href, kind }: { href: string; kind: string }): boolean {
    if (kind === 'anchor') return false; // the component jumps inside the document
    if (kind === 'external') {
      void (async () => {
        try {
          const domain = new URL(href).hostname;
          if (await backend.isURLAllowed(domain)) BrowserOpenURL(href);
          else opt.onsecurity?.({ type: 'url', resource: domain });
        } catch {
          console.error('Invalid URL clicked:', href);
        }
      })();
      return true;
    }
    if (/\.(md|markdown)(#.*)?$/i.test(href)) {
      void (async () => {
        const abs = await backend.resolveRelativePath(await baseDir(), href.replace(/#.*$/, ''));
        if (await backend.isPathAllowed(abs)) opt.onopenfile?.({ path: abs });
        else opt.onsecurity?.({ type: 'path', resource: await backend.getParentDir(abs) });
      })();
      return true;
    }
    console.warn('Direct file links are blocked for security. Use Markdown files or explicit whitelist.');
    return true;
  }

  async function loadMermaid() {
    const { default: mermaid } = await import('mermaid');
    mermaid.initialize({ startOnLoad: false, ...opt.mermaid(), fontFamily: 'inherit' } as Parameters<typeof mermaid.initialize>[0]);
    return mermaid;
  }

  return { checkImages, onNavigate, loadMermaid };
}
