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
  MLCRemoveSignature,
  MLCFitTable,
  OpenTemplatesFolder,
  TemplatesFolder,
  PrintPDF,
  GetPDFWatermark,
  SetPDFWatermark,
  GetPDFPreviewNoImages,
  SetPDFPreviewNoImages,
  GetSensorMonitorStatus,
  SetSensorMonitorURL,
  SensorMonitorStartLogin,
  SensorMonitorCancelLogin,
  SensorMonitorLogout,
  SensorMonitorLookup,
  SensorMonitorSource,
  SensorMonitorCompare,
  SensorMonitorSaveDraft,
  SetUnsavedChanges,
  CloseApp,
  MLCFormSchema,
  MLCFillForm,
  MLCFormRefOf,
  SaveFilledForm,
  LoadFormValues,
  MLCIsFormRecord
} from '../../bindings/marksafe/app'
import type {
  Analysis as MLCAnalysis,
  Diagnostic as MLCDiagnostic,
  TemplateInfo as MLCTemplate,
  TemplateInput as MLCTemplateField,
  RenderResult as MLCRenderResult,
  FormField as MLCFormField,
  FormRef as MLCFormRef,
} from '../../bindings/github.com/mlechner911/mdviewer/mlcext/models';
export type { MLCFormField, MLCFormRef };
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
/** Print through the PDF: rendered like the export, opened in the system's PDF viewer. */
export async function printPDF(md: string, docPath: string, watermark = ''): Promise<PDFResult> {
  if (!isWailsReady()) return pdfFailure('Wails not ready');
  try {
    return (await PrintPDF(md, docPath, watermark)) as PDFResult;
  } catch (err) {
    return pdfFailure(err);
  }
}

/** Watermark for PDFs from MarkSafe (a user setting; "" for none). */
export async function getPDFWatermark(): Promise<string> {
  if (!isWailsReady()) return '';
  try { return (await GetPDFWatermark()) ?? ''; } catch { return ''; }
}

export async function setPDFWatermark(text: string): Promise<void> {
  if (!isWailsReady()) return;
  try { await SetPDFWatermark(text); } catch (err) { console.error('setPDFWatermark failed:', err); }
}

export async function exportPDF(md: string, docPath: string, watermark = ''): Promise<PDFResult> {
  if (!isWailsReady()) return pdfFailure('Wails not ready');
  try {
    return (await ExportPDF(md, docPath, watermark)) as PDFResult;
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

/** Show the user's template folder (JSON + Markdown, editable without a release). */
export async function openTemplatesFolder(): Promise<void> {
  if (!isWailsReady()) return;
  try { await OpenTemplatesFolder(); } catch (err) { console.error('openTemplatesFolder failed:', err); }
}

export async function templatesFolder(): Promise<string> {
  if (!isWailsReady()) return '';
  try { return (await TemplatesFolder()) ?? ''; } catch { return ''; }
}

/** Placeholders instead of images in the PDF preview (export and print keep them). */
export async function getPDFPreviewNoImages(): Promise<boolean> {
  if (!isWailsReady()) return false;
  try { return !!(await GetPDFPreviewNoImages()); } catch { return false; }
}

export async function setPDFPreviewNoImages(v: boolean): Promise<void> {
  if (!isWailsReady()) return;
  try { await SetPDFPreviewNoImages(v); } catch (err) { console.error('setPDFPreviewNoImages failed:', err); }
}

/** Column widths of the table at line: "auto" or "content"; null when there is no table. */
export async function mlcFitTable(md: string, line: number, mode: 'auto' | 'content'): Promise<string | null> {
  if (!isWailsReady()) return null;
  try { return (await MLCFitTable(md, line, mode)) || null; } catch (err) { console.error('mlcFitTable failed:', err); return null; }
}

// --- Sensor Monitor (off until an address is set; see sensor.go) ---

export type SensorMonitorStatus = { address: string; signedIn: boolean; email: string; expires: string };
export type SensorLoginCode = { userCode: string; url: string; expiresIn: number };

const SENSOR_OFF: SensorMonitorStatus = { address: '', signedIn: false, email: '', expires: '' };

export async function sensorMonitorStatus(): Promise<SensorMonitorStatus> {
  if (!isWailsReady()) return SENSOR_OFF;
  try { return ((await GetSensorMonitorStatus()) as SensorMonitorStatus) ?? SENSOR_OFF; } catch { return SENSOR_OFF; }
}

/** Stores the address ("" turns the connection off); returns it normalized. Throws with the reason. */
export async function setSensorMonitorURL(address: string): Promise<string> {
  if (!isWailsReady()) return '';
  return (await SetSensorMonitorURL(address)) ?? '';
}

/** Starts the device login; the result comes as the event "sensormonitor-login". Throws with the reason. */
export async function sensorMonitorStartLogin(): Promise<SensorLoginCode | null> {
  if (!isWailsReady()) return null;
  return ((await SensorMonitorStartLogin()) as SensorLoginCode) ?? null;
}

export async function sensorMonitorCancelLogin(): Promise<void> {
  if (!isWailsReady()) return;
  try { await SensorMonitorCancelLogin(); } catch { /* ignore */ }
}

export async function sensorMonitorLogout(): Promise<void> {
  if (!isWailsReady()) return;
  try { await SensorMonitorLogout(); } catch { /* ignore */ }
}

export type SensorError = { code: string; message: string; params: Record<string, string>; signIn: boolean };
export type SensorDocument = {
  id: string; doc_no: string; ref_no: string; version: string; title: string; filename: string; status: string;
  approved_at: string; approved_by_label: string; uploaded_by_label: string; valid_until: string; created_at: string;
};
export type SensorStand = { ref_no: string; approved: SensorDocument | null; draft: SensorDocument | null };
export type SensorComparison = SensorStand & { matches_approved: boolean; matches_draft: boolean };
export type SensorFinding = { line: number; col: number; severity: string; code: string; message: string };
export type SensorSavedDraft = SensorDocument & { warnings: SensorFinding[] | null };

function sensorFailure(err: any): SensorError {
  return { code: '', message: String(err?.message ?? err), params: {}, signIn: false };
}

export async function sensorMonitorLookup(ref: string): Promise<{ stand: SensorStand | null; error: SensorError | null }> {
  if (!isWailsReady()) return { stand: null, error: sensorFailure('Wails not ready') };
  try { return (await SensorMonitorLookup(ref)) as any; } catch (err) { return { stand: null, error: sensorFailure(err) }; }
}

export async function sensorMonitorSource(id: string): Promise<{ text: string; error: SensorError | null }> {
  if (!isWailsReady()) return { text: '', error: sensorFailure('Wails not ready') };
  try { return (await SensorMonitorSource(id)) as any; } catch (err) { return { text: '', error: sensorFailure(err) }; }
}

export async function sensorMonitorCompare(ref: string, md: string): Promise<{ comparison: SensorComparison | null; error: SensorError | null }> {
  if (!isWailsReady()) return { comparison: null, error: sensorFailure('Wails not ready') };
  try { return (await SensorMonitorCompare(ref, md)) as any; } catch (err) { return { comparison: null, error: sensorFailure(err) }; }
}

export async function sensorMonitorSaveDraft(md: string, filename: string, reason: string, pin: string): Promise<{ draft: SensorSavedDraft | null; error: SensorError | null }> {
  if (!isWailsReady()) return { draft: null, error: sensorFailure('Wails not ready') };
  try { return (await SensorMonitorSaveDraft(md, filename, reason, pin)) as any; } catch (err) { return { draft: null, error: sensorFailure(err) }; }
}

/** Whether a tab has unsaved changes: then closing the window asks first. */
export async function setUnsavedChanges(unsaved: boolean): Promise<void> {
  if (!isWailsReady()) return;
  try { await SetUnsavedChanges(unsaved); } catch { /* ignore */ }
}

/** Quit after the user answered the unsaved-changes question. */
export async function closeApp(): Promise<void> {
  if (!isWailsReady()) return;
  try { await CloseApp(); } catch (err) { console.error('closeApp failed:', err); }
}

// --- Filling in a form (mlcext.FormSchema / FillFields, P-SENS-20261028-01) ---

/** The fields of md for the input mask. */
export async function mlcFormSchema(md: string): Promise<MLCFormField[]> {
  if (!isWailsReady()) return [];
  try { return ((await MLCFormSchema(md)) ?? []).filter((f): f is MLCFormField => !!f); } catch (err) { console.error('mlcFormSchema failed:', err); return []; }
}

/** md with values (field name → value) set into its fields. */
export async function mlcFillForm(md: string, values: Record<string, string>): Promise<string> {
  if (!isWailsReady()) return md;
  try { return await MLCFillForm(md, values); } catch (err) { console.error('mlcFillForm failed:', err); return md; }
}

/** The form md is filled from (doc_id, version, SHA-256). */
export async function mlcFormRef(md: string): Promise<MLCFormRef> {
  const none = { docId: '', version: '', docNo: '', sha256: '' };
  if (!isWailsReady()) return none;
  try { return (await MLCFormRefOf(md)) ?? none; } catch (err) { console.error('mlcFormRef failed:', err); return none; }
}

/** md is a filled-in form (meta.schema mlc-form-record/v1). */
export async function mlcIsFormRecord(md: string): Promise<boolean> {
  if (!isWailsReady()) return false;
  try { return await MLCIsFormRecord(md); } catch { return false; }
}

/** Save md as a filled-in form; null when cancelled. Errors are thrown. */
export async function saveFilledForm(md: string, form: MLCFormRef, filledBy: string, path: string): Promise<FileResult | null> {
  if (!isWailsReady()) return null;
  return (await SaveFilledForm(md, form, filledBy, path)) ?? null;
}

/** Field values from a JSON file; null when cancelled. Errors are thrown. */
export async function loadFormValues(): Promise<Record<string, string> | null> {
  if (!isWailsReady()) return null;
  const v = await LoadFormValues();
  return v ? (Object.fromEntries(Object.entries(v).filter(([, x]) => x !== undefined)) as Record<string, string>) : null;
}
