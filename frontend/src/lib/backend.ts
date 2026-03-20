// Helper wrappers for Wails runtime calls and graceful fallbacks for dev (Vite) mode
import { GetStyleCSS, RenderMarkdown, OpenFile, SaveFile, GetInitialContent, ReadFile, ExportHTML, GetFileTitle } from '../../wailsjs/go/main/App.js'

export interface FileResult {
  path: string;
  content: string;
}

export function isWailsReady(): boolean {
  return typeof window !== 'undefined' && (window as any).go && (window as any).go.main && (window as any).go.main.App;
}

export async function getStyleCSS(style: string): Promise<string> {
  try {
    if (isWailsReady()) {
      return await GetStyleCSS(style);
    }
    console.warn('Wails runtime not available; returning empty highlighting CSS');
    return '';
  } catch (err) {
    console.error('getStyleCSS failed:', err);
    return '';
  }
}

export async function renderMarkdown(value: string, themeStyle: string): Promise<string> {
  try {
    if (isWailsReady()) {
      return await RenderMarkdown(value, themeStyle);
    }
    // Try client-side renderer if available (marked), otherwise provide placeholder
    if (typeof (window as any).marked === 'function') {
      return (window as any).marked(value);
    }
    return '<pre style="white-space:pre-wrap;">Preview unavailable (no backend runtime)</pre>';
  } catch (err) {
    console.error('renderMarkdown failed:', err);
    return '<pre style="white-space:pre-wrap;">Preview error</pre>';
  }
}

export async function openFile(): Promise<FileResult | undefined> {
  if (!isWailsReady()) return undefined;
  try { return await OpenFile(); } catch (err) { console.error('openFile failed:', err); return undefined; }
}

export async function getFileTitle(path: string): Promise<string> {
  if (!isWailsReady()) return path.split(/[\\\\/]/).pop() || "Untitled";
  try { return await GetFileTitle(path); } catch (err) { console.error('getFileTitle failed:', err); return path; }
}

export async function saveFile(content: string): Promise<string | undefined> {
  if (!isWailsReady()) return undefined;
  try { return await SaveFile(content); } catch (err) { console.error('saveFile failed:', err); return undefined; }
}

export async function exportHTML(html: string, css: string): Promise<void> {
  if (!isWailsReady()) return;
  try { await ExportHTML(html, css); } catch (err) { console.error('exportHTML failed:', err); }
}

export async function getInitialContent(): Promise<FileResult | undefined> {
  if (!isWailsReady()) return undefined;
  try { return await GetInitialContent(); } catch (err) { console.error('getInitialContent failed:', err); return undefined; }
}

export async function readFile(path: string): Promise<string | undefined> {
  if (!isWailsReady()) return undefined;
  try { return await ReadFile(path); } catch (err) { console.error('readFile failed:', err); return undefined; }
}
