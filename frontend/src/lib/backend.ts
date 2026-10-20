// Helper wrappers for Wails runtime calls and graceful fallbacks for dev (Vite) mode
import { 
  GetStyleCSS, 
  RenderMarkdown, 
  OpenFile, 
  SaveFile, 
  SaveFileAs, 
  GetInitialContent, 
  ReadFile, 
  ExportHTML, 
  GetFileTitle,
  IsPathAllowed,
  IsURLAllowed,
  AddPathToWhitelist,
  AddURLToWhitelist,
  GetParentDir,
  ResolveRelativePath,
  UpdateMenu,
  GetVersion,
  SetWindowTitle,
  OpenBrowseWindow,
  CheckExternalURL,
  AnalyzeMLC,
  MLCDocVersion,
  PreviewPDF,
  ExportPDF,
  MLCTemplates,
  RenderMLCTemplate,
  SuggestMLCRefNo,
  MLCMoveHeader,
  MLCSignatureState,
  MLCRemoveSignature
} from '../../bindings/marksafe/app'
import type {
  Analysis as MLCAnalysis,
  Diagnostic as MLCDiagnostic,
  TemplateInfo as MLCTemplate,
  TemplateInput as MLCTemplateField,
  RenderResult as MLCRenderResult,
} from '../../bindings/github.com/mlechner911/mdviewer/mlcext/models';
import { get } from 'svelte/store';
import { t } from '../i18n';

export interface FileResult {
  path: string;
  content: string;
}

// True when running inside the Wails webview. `window._wails` is not a usable
// signal in v3 (the runtime creates it on import, even in a plain browser), so
// check for the native message bridge instead: WebView2 on Windows, WebKit's
// `external` handler on macOS/Linux.
export function isWailsReady(): boolean {
  if (typeof window === 'undefined') return false;
  const w = window as any;
  return Boolean(w.chrome?.webview?.postMessage || w.webkit?.messageHandlers?.external);
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

export type { MLCAnalysis, MLCDiagnostic, MLCTemplate, MLCTemplateField, MLCRenderResult };

/**
 * Analysis of the MLC Markdown Extension (non-standard): diagnostics, field
 * and hint positions, shared lists. Null outside Wails.
 */
export async function analyzeMLC(value: string): Promise<MLCAnalysis | null> {
  if (!isWailsReady()) return null;
  try {
    return await AnalyzeMLC(value);
  } catch (err) {
    console.error('analyzeMLC failed:', err);
    return null;
  }
}

/** Installed mlcdoc version, "" when there is none (PDF preview hidden). */
export async function mlcdocVersion(): Promise<string> {
  if (!isWailsReady()) return '';
  try {
    return (await MLCDocVersion()) ?? '';
  } catch {
    return '';
  }
}

export type PDFResult = {
  pages: string[]; path: string; warnings: string[]; ignored: string[]; error: string; errorLine: number;
};

function pdfFailure(err: any): PDFResult {
  return { pages: [], path: '', warnings: [], ignored: [], error: String(err?.message ?? err), errorLine: 0 };
}

/** Page images rendered by mlcdoc (as in the exported PDF), or why it cannot render. */
export async function previewPDF(md: string, docPath: string): Promise<PDFResult | null> {
  if (!isWailsReady()) return null;
  try {
    return (await PreviewPDF(md, docPath)) as PDFResult;
  } catch (err) {
    return pdfFailure(err);
  }
}

/** Render with mlcdoc and save via a dialog; path "" when cancelled. */
export async function exportPDF(md: string, docPath: string): Promise<PDFResult> {
  if (!isWailsReady()) return pdfFailure('Wails not ready');
  try {
    return (await ExportPDF(md, docPath)) as PDFResult;
  } catch (err) {
    return pdfFailure(err);
  }
}

/** Templates of the MLC Markdown Extension (built-in + document folder). */
export async function mlcTemplates(): Promise<MLCTemplate[]> {
  if (!isWailsReady()) return [];
  try {
    return ((await MLCTemplates()) ?? []).filter((x): x is MLCTemplate => !!x);
  } catch {
    return [];
  }
}

/** Fill a template; field errors and diagnostics of the new document. */
export async function renderMLCTemplate(id: string, values: Record<string, any>): Promise<MLCRenderResult> {
  const empty = { content: '', fileName: '', fieldErrors: {}, diagnostics: [] } as unknown as MLCRenderResult;
  if (!isWailsReady()) return { ...empty, fieldErrors: { '': 'Wails not ready' } };
  try {
    return await RenderMLCTemplate(id, values);
  } catch (err: any) {
    return { ...empty, fieldErrors: { '': String(err?.message ?? err) } };
  }
}

/** RefNo proposal from the document folder (only a proposal). */
export async function suggestMLCRefNo(prefix: string, area: string, docPath: string): Promise<string> {
  if (!isWailsReady()) return '';
  try {
    return (await SuggestMLCRefNo(prefix, area, docPath)) ?? '';
  } catch {
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
    return `<pre style="white-space:pre-wrap;">${get(t)('previewUnavailable')}</pre>`;
  } catch (err) {
    console.error('renderMarkdown failed:', err);
    return `<pre style="white-space:pre-wrap;">${get(t)('previewError')}</pre>`;
  }
}

export async function openFile(): Promise<FileResult | undefined> {
  if (!isWailsReady()) return undefined;
  try { return (await OpenFile()) ?? undefined; } catch (err) { console.error('openFile failed:', err); return undefined; }
}

export async function getFileTitle(path: string): Promise<string> {
  if (!isWailsReady()) return path.split(/[\\\\/]/).pop() || "Untitled";
  try { return await GetFileTitle(path); } catch (err) { console.error('getFileTitle failed:', err); return path; }
}

export async function saveFile(path: string, content: string): Promise<string | undefined> {
  if (!isWailsReady()) return undefined;
  try { return await SaveFile(path, content); } catch (err) { console.error('saveFile failed:', err); return undefined; }
}

export async function saveFileAs(defaultFilename: string, content: string): Promise<string | undefined> {
  if (!isWailsReady()) return undefined;
  try { return await SaveFileAs(defaultFilename, content); } catch (err) { console.error('saveFileAs failed:', err); return undefined; }
}

/** md (the source) only names the file after its doc_id. */
export async function exportHTML(html: string, css: string, md = ''): Promise<void> {
  if (!isWailsReady()) return;
  try { await ExportHTML(html, css, md); } catch (err) { console.error('exportHTML failed:', err); }
}

/** Quick fix: legacy header lines into the front matter; null when there is nothing to move. */
export async function mlcMoveHeader(md: string): Promise<string | null> {
  if (!isWailsReady()) return null;
  try { return (await MLCMoveHeader(md)) || null; } catch (err) { console.error('mlcMoveHeader failed:', err); return null; }
}

/** Signed, and does the content still match the signed checksum? */
export async function mlcSignatureState(md: string): Promise<{ signed: boolean; intact: boolean; signedBy: string; signedAt: string }> {
  const none = { signed: false, intact: false, signedBy: '', signedAt: '' };
  if (!isWailsReady()) return none;
  try { return (await MLCSignatureState(md)) ?? none; } catch (err) { console.error('mlcSignatureState failed:', err); return none; }
}

export async function mlcRemoveSignature(md: string): Promise<string> {
  if (!isWailsReady()) return md;
  try { return await MLCRemoveSignature(md); } catch (err) { console.error('mlcRemoveSignature failed:', err); return md; }
}

export async function getInitialContent(): Promise<FileResult | undefined> {
  if (!isWailsReady()) return undefined;
  try { return (await GetInitialContent()) ?? undefined; } catch (err) { console.error('getInitialContent failed:', err); return undefined; }
}

export async function readFile(path: string): Promise<string | undefined> {
  if (!isWailsReady()) return undefined;
  try { return await ReadFile(path); } catch (err) { console.error('readFile failed:', err); return undefined; }
}

// Security Whitelist Bindings
export async function isPathAllowed(path: string): Promise<boolean> {
  if (!isWailsReady()) return true;
  return await IsPathAllowed(path);
}

export async function isURLAllowed(url: string): Promise<boolean> {
  if (!isWailsReady()) return true;
  return await IsURLAllowed(url);
}

export async function addPathToWhitelist(path: string): Promise<void> {
  if (!isWailsReady()) return;
  await AddPathToWhitelist(path);
}

export async function openBrowseWindow(rootDir: string, relPath: string): Promise<void> {
  if (!isWailsReady()) return;
  await OpenBrowseWindow(rootDir, relPath);
}

export async function checkExternalURL(url: string): Promise<boolean> {
  if (!isWailsReady()) return true;
  try {
    return await CheckExternalURL(url);
  } catch (err) {
    console.error('checkExternalURL failed:', err);
    return true;
  }
}

export async function addURLToWhitelist(url: string): Promise<void> {
  if (!isWailsReady()) return;
  await AddURLToWhitelist(url);
}

export async function getParentDir(path: string): Promise<string> {
  if (!isWailsReady()) return "";
  return await GetParentDir(path);
}

export async function resolveRelativePath(baseDir: string, relPath: string): Promise<string> {
  if (!isWailsReady()) return relPath;
  return await ResolveRelativePath(baseDir, relPath);
}

// Menu Translation Binding
export async function updateMenu(translations: Record<string, string>): Promise<void> {
  if (!isWailsReady()) return;
  await UpdateMenu(translations);
}

// Version Binding
export async function getVersion(): Promise<string> {
  if (isWailsReady()) {
    try {
      const v = await GetVersion();
      if (v) return v;
    } catch (err) {
      console.error('getVersion failed:', err);
    }
  }
  return '1.5.1';
}

// Window Title Binding
export async function setWindowTitle(title: string): Promise<void> {
  if (isWailsReady()) {
    try {
      await SetWindowTitle(title);
    } catch (err) {
      console.error('setWindowTitle failed:', err);
    }
  }
}
