<script lang="ts">
  /**
   * Main Application component for MD Viewer.
   * Refactored for Svelte 5 Runes, CodeMirror 6 markdown editor, dark/light themes, and crisp print output.
   */
  import { onMount, tick, untrack } from 'svelte';
  import { get } from 'svelte/store';
  import { EventsOn, EventsOff, OnFileDrop, OnFileDropOff, BrowserOpenURL } from './lib/wails';
  import * as backend from './lib/backend';
  
  // Components
  import Editor from './components/Editor.svelte';
  import Preview from './components/Preview.svelte';
  import BrowseView from './shared/browse/BrowseView.svelte';
  import mlcMarkdownCSS from '../../packages/mlc-markdown/styles/mlc-markdown.css?raw';
  import { mlcSnippets } from '../../packages/mlcext-editor/src/index';
  import { createWailsClient, toBrowseRel } from './lib/wailsBrowseClient';
  import WhitelistModal from './components/WhitelistModal.svelte';
  import ConfirmModal from './components/ConfirmModal.svelte';
  import AboutModal from './components/AboutModal.svelte';
  import HamburgerMenu from './components/HamburgerMenu.svelte';
  import TabsBar from './components/TabsBar.svelte';
  import StatusBar from './components/StatusBar.svelte';
  import PdfPreview from './components/PdfPreview.svelte';
  import FormFillPanel from './components/FormFillPanel.svelte';
  import NewFromTemplateModal from './components/NewFromTemplateModal.svelte';
  import SettingsModal from './components/SettingsModal.svelte';
  import FieldDialog from './components/FieldDialog.svelte';
  // Optional Sensor Monitor module: the real one in the Sensor Monitor
  // variant, null in the standard build (vite.config.ts).
  import { Extension as SensorMonitor, SettingsSection as SensorMonitorSettings, texts as sensorMonitorTexts, license as sensorMonitorLicense } from '$sensormonitor';

  // State & Config
  import { getTheme } from './themes';
  import { t, locale, translations, addTranslations } from './i18n';
  import { APP_THEME, STYLE, DEFAULTS } from './lib/constants';
  import { c_welcomeMarkdown } from './lib/devdefmd.js';
  import { 
    appTheme, effectiveAppTheme, splitWidth, isFocusMode, 
    isEditorHidden, dropMessage, showToast,
    appVersion, toastType
  } from './lib/stores';

  // CONFIGURATION: Show HTML toolbar in Vite development mode, hide in production.
  const SHOW_HTML_TOOLBAR = import.meta.env.DEV;

  interface Tab {
    id: string;
    title: string;
    path: string | null;
    content: string;
    isDirty: boolean;
    /** Being filled in: the form it comes from (mlcext.FormRef). Stays
     *  when the mask is closed (hidden), so saving still writes a record
     *  and never replaces the form. */
    form?: { ref: backend.MLCFormRef; hidden?: boolean };
  }

  // --- Svelte 5 Runes: State ---
  let tabs = $state<Tab[]>([]);
  let activeTabIndex = $state(0);
  let isReady = $state(false);
  let fontSize = $state(DEFAULTS.fontSize);
  
  let showSecurityModal = $state(false);
  let securityType = $state<'path' | 'url'>('path');
  let securityResource = $state("");

  let isSyncScroll = $state(true);
  let editorComponent: any = $state();
  let previewComponent: any = $state();
  let scrollLock = false;

  // Scroll sync by source line (data-line in the preview, line blocks in
  // the editor), so both sides show the same passage.
  function handleEditorScroll(line: number) {
    if (!isSyncScroll || scrollLock || !previewComponent) return;
    scrollLock = true;
    previewComponent.scrollToLine(line);
    setTimeout(() => { scrollLock = false; }, 80);
  }

  function handlePreviewScroll(line: number) {
    if (!isSyncScroll || scrollLock || !editorComponent) return;
    scrollLock = true;
    editorComponent.scrollToLine(line);
    setTimeout(() => { scrollLock = false; }, 80);
  }

  // --- Svelte 5 Runes: Derived ---
  const markdown = $derived(tabs[activeTabIndex]?.content || "");
  const activeTab = $derived(tabs[activeTabIndex] || null);
  const currentPreviewTheme = $derived(getTheme($effectiveAppTheme));

  const wordCount = $derived(markdown ? (markdown.trim().split(/\s+/).filter(Boolean).length) : 0);
  const charCount = $derived(markdown ? markdown.length : 0);
  const readingTime = $derived(Math.ceil(wordCount / 225));

  const toolbarClass = $derived(STYLE.toolbar[$effectiveAppTheme]);
  const editorClass = $derived(STYLE.editor[$effectiveAppTheme]);
  const buttonClass = $derived(STYLE.button[$effectiveAppTheme]);
  const dividerClass = $derived(STYLE.divider[$effectiveAppTheme]);
  const focusButtonClass = $derived(STYLE.focusButton[$effectiveAppTheme]);

  const defaultMarkdown = () => get(t)('welcomeTitle') + (c_welcomeMarkdown[$locale as 'en'|'de'|'es'|'fr'] ?? c_welcomeMarkdown.en);

  // Re-render welcome tabs when the app language changes (only tabs that
  // still show the welcome text, not e.g. a document from a template).
  let prevLocale = $locale;
  let prevWelcome = defaultMarkdown();
  $effect(() => {
    if ($locale === prevLocale) return;
    prevLocale = $locale;
    const oldContent = prevWelcome;
    const newContent = prevWelcome = defaultMarkdown();
    tabs = tabs.map(tab => (!tab.path && !tab.isDirty && tab.content === oldContent) ? { ...tab, content: newContent } : tab);
  });

  function createNewTab(title = $t('untitled'), content = "", path: string | null = null): Tab {
    return {
      id: Math.random().toString(36).substring(2, 11),
      title,
      path,
      content,
      isDirty: false
    };
  }

  // New document of the MLC Markdown Extension from a template.
  let showTemplateModal = $state(false);
  let showSettings = $state(false);
  // MLC Markdown Extension analysis of the active document (status bar,
  // modules); set in debouncedUpdate.
  let mlcAnalysis = $state<backend.MLCAnalysis | null>(null);

  // Built-in modules (so far only the optional Sensor Monitor): their
  // texts, settings sections, status bar entries and the watermark
  // suggestion before a PDF.
  addTranslations(sensorMonitorTexts);
  const settingsSections = [SensorMonitorSettings].filter((c): c is NonNullable<typeof c> => !!c);
  let sensorMonitor = $state<any>(null);
  let moduleStatus = $state<{ text: string; tone?: 'ok' | 'warn'; title?: string }[]>([]);
  function openModuleTab(title: string, text: string) {
    tabs = [...tabs, createNewTab(title, text)];
    activeTabIndex = tabs.length - 1;
  }
  // Field dialog (MLC Markdown Extension): a new field at the cursor, or
  // the field at line, col replaced (range from mlcext.FieldAt).
  let fieldDialog = $state<{ initial: backend.MLCFieldSpec | null; range: { line: number; col: number; endLine: number; endCol: number } | null } | null>(null);
  function insertField() {
    fieldDialog = { initial: null, range: null };
  }
  async function editField(line: number, col: number) {
    const e = await backend.mlcFieldAt(markdown, line, col);
    if (!e) return;
    fieldDialog = { initial: e.spec, range: { line: e.line, col: e.col, endLine: e.endLine, endCol: e.endCol } };
  }
  function applyField(text: string) {
    const d = fieldDialog;
    fieldDialog = null;
    if (!d) return;
    if (d.range) editorComponent?.replaceRange(d.range.line, d.range.col, d.range.endLine, d.range.endCol, text);
    else editorComponent?.insertText(text);
  }

  async function suggestedWatermark(): Promise<string> {
    return (await sensorMonitor?.suggestWatermark?.(pdfWatermark)) ?? '';
  }
  function createFromTemplate(content: string, fileName: string) {
    const newTab = createNewTab(fileName || $t('untitled'), content);
    newTab.isDirty = true; // not saved yet: closing asks
    tabs = [...tabs, newTab];
    activeTabIndex = tabs.length - 1;
    showTemplateModal = false;
  }

  // --- Filling in a form (P-SENS-20261028-01 Phase 2) ---
  // "Fill in" opens the form in a new tab without a file, so saving can
  // never replace the form; a record that is already saved is filled on in
  // its own tab. Saving writes the record's meta block (backend).
  const readStore = (k: string) => { try { return localStorage.getItem(k) ?? ''; } catch { return ''; } };
  let filledBy = $state(readStore('marksafe_filled_by'));
  function setFilledBy(name: string) {
    filledBy = name;
    try { localStorage.setItem('marksafe_filled_by', name); } catch { /* private mode */ }
  }
  const canFill = $derived(!!mlcAnalysis?.fields?.length && (!tabs[activeTabIndex]?.form || !!tabs[activeTabIndex]?.form?.hidden));
  const filling = $derived(!!tabs[activeTabIndex]?.form && !tabs[activeTabIndex]?.form?.hidden);

  async function startFill() {
    const tab = tabs[activeTabIndex];
    if (!tab) return;
    previewMode = 'html';
    if (tab.form) {
      tab.form.hidden = false; // the mask was closed: open it again
      return;
    }
    const ref = await backend.mlcFormRef(tab.content);
    // The tab, not the index: the user may have switched tabs meanwhile.
    if (tab.path && (await backend.mlcIsFormRecord(tab.content))) {
      tab.form = { ref };
    } else {
      const newTab = createNewTab(`${(tab.title || $t('untitled')).replace(/\.md$/, '')} – ${$t('formFillTab')}`, tab.content);
      newTab.form = { ref };
      tabs = [...tabs, newTab];
      activeTabIndex = tabs.length - 1;
    }
  }

  function fillChanged(md: string) {
    const tab = tabs[activeTabIndex];
    if (!tab || tab.content === md) return;
    tab.content = md;
    tab.isDirty = true;
  }

  async function loadFillValues() {
    const tab = tabs[activeTabIndex];
    if (!tab) return;
    try {
      const values = await backend.loadFormValues();
      if (!values) return;
      fillChanged(await backend.mlcFillForm(tab.content, values));
      showToast($t('formFillLoaded', String(Object.keys(values).length)), 2500, 'info');
    } catch (err) {
      showToast(String(err), 6000, 'error');
    }
  }

  // asNew (Save As): a new record file, the tab keeps its path until then.
  async function saveFilledForm(asNew = false) {
    const index = activeTabIndex;
    const tab = tabs[index];
    if (!tab?.form || isSaving) return;
    if (!filledBy.trim()) {
      showToast($t('formFillByMissing'), 4000, 'error');
      return;
    }
    try {
      isSaving = true;
      const res = await backend.saveFilledForm(tab.content, tab.form.ref, filledBy.trim(), asNew ? '' : tab.path ?? '');
      if (!res) return; // cancelled
      tabs[index].path = res.path;
      tabs[index].content = res.content;
      tabs[index].title = await backend.getFileTitle(res.path);
      tabs[index].isDirty = false;
      await backend.addPathToWhitelist(await backend.getParentDir(res.path));
      showToast($t('formFillSaved', res.path), 3000, 'info');
      updateNativeMenu(get(locale));
    } catch (err) {
      showToast(String(err), 7000, 'error');
    } finally {
      isSaving = false;
    }
  }

  function addNewTab() {
    const newTab = createNewTab($t('untitled'), defaultMarkdown());
    tabs = [...tabs, newTab];
    activeTabIndex = tabs.length - 1;
  }

  async function handleCloseTab(index: number, event?: MouseEvent) {
    if (event) event.stopPropagation();
    const tab = tabs[index];
    if (tab?.isDirty) {
      const a = await askUnsaved([tab.title]);
      if (a === 'cancel') return;
      if (a === 'confirm' && !(await saveTab(index))) return;
      index = tabs.indexOf(tab);
      if (index < 0) return;
    }
    if (tabs.length === 1) {
      tabs = [createNewTab($t('untitled'), defaultMarkdown())];
      activeTabIndex = 0;
      return;
    }
    const wasActive = index === activeTabIndex;
    tabs = tabs.filter((_, i) => i !== index);
    if (wasActive) {
      activeTabIndex = Math.max(0, index - 1);
    } else if (index < activeTabIndex) {
      activeTabIndex--;
    }
  }

  function onContentInput() {
    if (tabs[activeTabIndex]) {
      tabs[activeTabIndex].isDirty = true;
    }
  }

  // Formatting Helpers for native menu and toolbar shortcuts
  function wrapSelection(prefix: string, suffix: string) {
    if (editorComponent?.wrapSelection) {
      editorComponent.wrapSelection(prefix, suffix);
      onContentInput();
    }
  }

  function prefixSelection(prefix: string) {
    if (editorComponent?.prefixSelection) {
      editorComponent.prefixSelection(prefix);
      onContentInput();
    }
  }

  async function handleSecurityRequest(detail: { type: 'path' | 'url', resource: string }) {
    securityType = detail.type;
    securityResource = detail.resource;
    showSecurityModal = true;
  }

  async function confirmSecurityRequest() {
    if (securityType === 'path') {
      await backend.addPathToWhitelist(securityResource);
    } else {
      await backend.addURLToWhitelist(securityResource);
    }
    showSecurityModal = false;
    debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
  }

  // Files into tabs (open dialog, recent files, drop, links): a file that
  // is already open comes to the front, a lone untouched welcome tab is
  // replaced, the folders are whitelisted. Returns how many were added.
  async function showFiles(files: { path: string; content: string }[]): Promise<number> {
    const added: Tab[] = [];
    for (const f of files) {
      const open = tabs.findIndex(t => t.path === f.path);
      if (open >= 0) {
        activeTabIndex = open;
        continue;
      }
      added.push(createNewTab(await backend.getFileTitle(f.path), f.content, f.path));
      await backend.addPathToWhitelist(await backend.getParentDir(f.path));
    }
    if (!added.length) return 0;
    if (tabs.length === 1 && !tabs[0].path && !tabs[0].isDirty && tabs[0].content === defaultMarkdown()) {
      tabs = added;
      activeTabIndex = 0;
    } else {
      tabs = [...tabs, ...added];
      activeTabIndex = tabs.length - 1;
    }
    updateNativeMenu(get(locale));
    return added.length;
  }

  // A file path (recent files, link in the preview) into a tab.
  async function showPath(path: string) {
    const content = tabs.some(t => t.path === path) ? '' : await backend.readFile(path);
    if (content !== undefined && content !== null) await showFiles([{ path, content }]);
  }

  async function handleOpenExternalMD(detail: { path: string }) {
    try {
      await showPath(detail.path);
    } catch (err) {
      console.error("Failed to open linked markdown:", err);
    }
  }

  const checkWailsReady = backend.isWailsReady;

  function updateEffectiveTheme() {
    if ($appTheme === 'auto') {
      const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      effectiveAppTheme.set(isDark ? 'dark' : 'light');
    } else {
      effectiveAppTheme.set($appTheme as 'dark' | 'light');
    }
  }

  function toggleAppTheme() {
    appTheme.update(current => {
      if (current === APP_THEME.DARK) return APP_THEME.LIGHT;
      if (current === APP_THEME.LIGHT) return APP_THEME.AUTO;
      return APP_THEME.DARK;
    });
  }

  // Native Menu Translation Sync
  async function updateNativeMenu(currentLocale: string) {
    if (!isReady || !checkWailsReady()) return;
    const tMap = translations[currentLocale];
    if (!tMap) return;
    const menuTranslations = {
      currentLocale,
      // every menu text, also those of built-in modules (menuSensor…)
      ...Object.fromEntries(Object.entries(tMap).filter(([k]) => k.startsWith('menu'))),
      aboutTitle: tMap.aboutTitle,
      aboutBody: tMap.aboutBody.replace('%s', get(appVersion))
    };
    await backend.updateMenu(menuTranslations);
  }

  let isResizing = $state(false);
  function startResizing() { isResizing = true; }
  function stopResizing() { isResizing = false; }
  function onMouseMove(event: MouseEvent) {
    if (!isResizing || $isEditorHidden) return;
    splitWidth.set((event.clientX / window.innerWidth) * 100);
    if ($splitWidth < 10) splitWidth.set(10);
    if ($splitWidth > 90) splitWidth.set(90);
  }

  // MLC Markdown Extension (non-standard) analysis of the active document:
  // status bar, editor menu. The editor gets its own via mlcSource.
  const mlcSource = {
    analyze: (md: string) => backend.analyzeMLC(md),
    // Quick fix for the legacy header line (mlcext.MoveHeaderToFrontMatter).
    fixLabel: (code: string) =>
      code === 'header-in-body' ? $t('fixMoveHeader') : code === 'table-column-narrow' ? $t('tableFitContent') : null,
    fix: (code: string, md: string, d: { line: number }) =>
      code === 'header-in-body'
        ? backend.mlcMoveHeader(md)
        : code === 'table-column-narrow'
          ? backend.mlcFitTable(md, d.line, 'content')
          : Promise.resolve(null),
  };
  // Table at the editor's cursor (separator row, 0: none): table menu.
  let cursorTable = $state(0);
  const TABLE_TEMPLATE = '| Spalte 1 | Spalte 2 | Spalte 3 |\n| --- | --- | --- |\n|  |  |  |\n';
  // Symbols that print in the PDF, from mlcext (same list for every tool).
  const pdfSymbols = $derived(mlcAnalysis?.lists?.symbols ?? []);
  // Alert boxes for the toolbar, colors as in mlc-markdown.css.
  const ALERTS = [
    { kind: 'NOTE', label: 'alertNote', color: '#0969da' },
    { kind: 'TIP', label: 'alertTip', color: '#1a7f37' },
    { kind: 'IMPORTANT', label: 'alertImportant', color: '#8250df' },
    { kind: 'WARNING', label: 'alertWarning', color: '#9a6700' },
    { kind: 'CAUTION', label: 'alertCaution', color: '#cf222e' },
  ] as const;
  const mlcMenu = $derived(mlcAnalysis?.detected ? mlcSnippets(mlcAnalysis.lists) : []);

  // PDF preview rendered by mlcdoc, offered for MLC documents when mlcdoc
  // is installed. The HTML preview stays mounted (export, scroll sync).
  let mlcdocVersion = $state('');
  let previewMode = $state<'html' | 'pdf'>('html');
  // PDF via mlcdoc (shipped with MarkSafe) for every document.
  const canPdfPreview = $derived(!!mlcdocVersion);
  // Formulas are not typeset by mlcdoc yet; they appear as plain text.
  let hasMath = $state(false);

  function showPdfError(res: { error: string; errorLine?: number }) {
    showToast($t('pdfExportFailed') + ' ' + (res.errorLine ? `${$t('line')} ${res.errorLine}: ` : '') + res.error, 6000, 'error');
  }

  async function handleExportPdf() {
    const tab = tabs[activeTabIndex];
    if (!tab || !mlcdocVersion) return;
    const res = await backend.exportPDF(tab.content, tab.path ?? '', await suggestedWatermark());
    if (res.error) {
      showPdfError(res);
      return;
    }
    if (!res.path) return; // cancelled
    const notes = [
      ...(res.ignored.length ? [$t('pdfIgnoredKeys', res.ignored.join(', '))] : []),
      ...(hasMath ? [$t('pdfMathAsText')] : []),
      ...res.warnings,
    ];
    showToast($t('pdfExported', res.path) + (notes.length ? ' · ' + notes.join(' · ') : ''), notes.length ? 7000 : 3000, 'info');
  }
  const showPdfPreview = $derived(canPdfPreview && previewMode === 'pdf');

  let timeout: ReturnType<typeof setTimeout>;
  function debouncedUpdate(value: string, themeStyle: string) {
    if (!isReady || !checkWailsReady()) return;
    clearTimeout(timeout);
    timeout = setTimeout(async () => {
      try {
        mlcAnalysis = await backend.analyzeMLC(value);
      } catch (err) {
        console.error("Failed to render markdown:", err);
      }
    }, 300);
  }

  async function handleOpen() {
    if (!isReady || !checkWailsReady()) return;
    try {
      const result = await backend.openFile();
      if (result?.path) await showFiles([result]);
    } catch (err) { console.error("Failed to open file:", err); showToast($t('openFailed'), 3000, 'error'); }
  }

  let isSaving = false;

  // A yes/no question as a promise (ConfirmModal below).
  type ConfirmAnswer = 'confirm' | 'alt' | 'cancel';
  let confirmAsk = $state<{ title: string; message: string; confirmLabel: string; cancelLabel?: string; altLabel?: string; resolve: (a: ConfirmAnswer) => void } | null>(null);
  function askConfirm(title: string, message: string, confirmLabel: string, cancelLabel?: string): Promise<boolean> {
    return new Promise((resolve) => (confirmAsk = { title, message, confirmLabel, cancelLabel, resolve: (a) => resolve(a === 'confirm') }));
  }
  function answerConfirm(a: ConfirmAnswer | boolean) {
    confirmAsk?.resolve(a === true ? 'confirm' : a === false ? 'cancel' : a);
    confirmAsk = null;
  }

  // Unsaved changes: Save / Don't save / Cancel.
  function askUnsaved(names: string[]): Promise<ConfirmAnswer> {
    const message = names.length === 1 ? $t('unsavedOne', names[0]) : $t('unsavedMany', names.length);
    return new Promise((resolve) => (confirmAsk = {
      title: $t('unsavedTitle'), message, confirmLabel: $t('save'), cancelLabel: $t('cancel'), altLabel: $t('unsavedDiscard'), resolve,
    }));
  }

  // Saves tab index (Save As for a new document); false when it stays
  // unsaved (cancelled or failed).
  async function saveTab(index: number): Promise<boolean> {
    activeTabIndex = index;
    await tick();
    await handleSave();
    return !tabs[index]?.isDirty;
  }

  // The window's X or quitting with unsaved tabs (backend held it back).
  let closeAsking = false;
  async function closeRequested() {
    if (closeAsking) return;
    closeAsking = true;
    try {
      const dirty = tabs.map((tab, i) => ({ tab, i })).filter((x) => x.tab.isDirty);
      if (dirty.length) {
        const a = await askUnsaved(dirty.map((x) => x.tab.title));
        if (a === 'cancel') return;
        if (a === 'confirm') {
          for (const x of dirty) {
            const i = tabs.indexOf(x.tab);
            if (i >= 0 && !(await saveTab(i))) return;
          }
        }
      }
      await backend.closeApp();
    } finally {
      closeAsking = false;
    }
  }

  // Tell the backend whether closing must ask.
  $effect(() => {
    const unsaved = tabs.some((tab) => tab.isDirty);
    untrack(() => void backend.setUnsavedChanges(unsaved));
  });

  // An edited signed document no longer matches its signature: saved as a
  // new version without the block, after asking (else it would look signed).
  // Returns the text to save, or null when the user cancels.
  async function contentToSave(index: number): Promise<string | null> {
    const content = tabs[index].content;
    const sig = await backend.mlcSignatureState(content);
    if (!sig.signed || sig.intact) return content;
    let at = sig.signedAt;
    const d = new Date(at);
    if (at && !isNaN(d.getTime())) at = d.toLocaleString($locale, { dateStyle: 'medium', timeStyle: 'short' });
    const ok = await askConfirm($t('signatureDropTitle'), $t('signatureDropBody', sig.signedBy || '?', at || '?'), $t('signatureDropConfirm'));
    if (!ok) return null;
    const stripped = await backend.mlcRemoveSignature(content);
    tabs[index].content = stripped;
    return stripped;
  }

  async function handleSave() {
    if (!isReady || !checkWailsReady() || isSaving) return;
    const tab = activeTab;
    if (!tab) return;
    if (tab.form) return saveFilledForm();

    // If document has no filename/path yet, route to Save As
    if (!tab.path) {
      await handleSaveAs();
      return;
    }

    try {
      isSaving = true;
      const content = await contentToSave(activeTabIndex);
      if (content === null) return;
      const savedPath = await backend.saveFile(tab.path, content);
      if (savedPath) {
        tabs[activeTabIndex].isDirty = false;
        showToast($t('fileSaved'), 2000, 'info');
        updateNativeMenu(get(locale));
      }
    } catch (err) {
      console.error("Failed to save file:", err);
      showToast($t('saveFailed'), 3000, 'error');
    } finally {
      isSaving = false;
    }
  }

  async function handleSaveAs() {
    if (!isReady || !checkWailsReady() || isSaving) return;
    const tab = activeTab;
    if (!tab) return;
    // A filled-in form is saved as a record, under a new name.
    if (tab.form) return saveFilledForm(true);

    try {
      isSaving = true;
      let defaultName = "document.md";
      if (tab.path) {
        defaultName = await backend.getFileTitle(tab.path);
      } else if (tab.title && tab.title !== $t('untitled')) {
        defaultName = tab.title.endsWith('.md') ? tab.title : `${tab.title}.md`;
      }

      const content = await contentToSave(activeTabIndex);
      if (content === null) return;
      const savedPath = await backend.saveFileAs(defaultName, content);
      if (savedPath) {
        tabs[activeTabIndex].path = savedPath;
        tabs[activeTabIndex].title = await backend.getFileTitle(savedPath);
        tabs[activeTabIndex].isDirty = false;
        const parentDir = await backend.getParentDir(savedPath);
        await backend.addPathToWhitelist(parentDir);
        showToast($t('fileSaved'), 2000, 'info');
        updateNativeMenu(get(locale));
      }
    } catch (err) {
      console.error("Failed to save file as:", err);
      showToast($t('saveFailed'), 3000, 'error');
    } finally {
      isSaving = false;
    }
  }

  async function handleExport() {
    if (!isReady || !checkWailsReady()) return;
    try {
      const theme = $effectiveAppTheme === 'dark' ? 'dark' : 'light';
      const pageCSS = `body { margin: 0; padding: 2rem; background: ${theme === 'dark' ? '#0d0e11' : '#ffffff'}; }
        .mlc-md { margin: 0 auto; }
        @media print { body { background: #ffffff; padding: 0; } }`;
      // The rendered preview (diagrams, formulas, embedded images), so the
      // file shows offline exactly what the preview shows.
      const article = (await previewComponent?.getExportHTML()) ?? '';
      if (!article) return;
      await backend.exportHTML(article, mlcMarkdownCSS + '\n' + pageCSS + '\n' + (previewComponent?.codeCSS() ?? ''), activeTab?.content ?? '');
    } catch (err) {
      console.error("Failed to export HTML:", err);
      showToast($t('exportFailed'), 3000, 'error');
    }
  }

  // Printing goes through the PDF (mlcdoc, opened in the system's PDF
  // viewer): it looks exactly like the export. window.print() did not work
  // in the Windows app and never looked like the PDF.
  async function handlePrint() {
    const tab = tabs[activeTabIndex];
    if (!tab) return;
    if (!mlcdocVersion) {
      showToast($t('printNeedsMlcdoc'), 5000, 'error');
      return;
    }
    const wm = await suggestedWatermark();
    showToast($t('printPreparing'), 2000, 'info');
    const res = await backend.printPDF(tab.content, tab.path ?? '', wm);
    if (res.error) {
      showPdfError(res);
      return;
    }
    showToast($t('printOpened'), 3000, 'info');
  }

  // Watermark for PDFs (setting; changed in the PDF preview bar).
  let pdfWatermark = $state('');
  // Placeholders instead of images in the PDF preview (setting).
  let pdfNoImages = $state(false);
  async function changeNoImages(v: boolean) {
    await backend.setPDFPreviewNoImages(v);
    pdfNoImages = v;
  }
  async function changeWatermark(text: string) {
    await backend.setPDFWatermark(text);
    pdfWatermark = text.trim();
  }

  function adjustFontSize(delta: number) {
    fontSize = Math.min(Math.max(fontSize + delta, 50), 200);
  }

  // Markdown files dropped on the window.
  async function handleFileDrop(paths: string[]) {
    const files: { path: string; content: string }[] = [];
    for (const path of paths ?? []) {
      if (!/\.(md|markdown|mdown|mkd|mdx)$/i.test(path)) continue;
      try {
        const content = tabs.some(t => t.path === path) ? '' : await backend.readFile(path);
        if (content !== undefined && content !== null) files.push({ path, content });
      } catch (err) { console.error("Failed to read dropped file:", err); }
    }
    const n = await showFiles(files);
    if (n > 0) showToast($t('filesLoaded', n));
  }

  async function handleOpenRecent(path: string) {
    if (!isReady || !checkWailsReady()) return;
    try {
      await showPath(path);
    } catch (err) {
      console.error("Failed to open recent file:", err);
      showToast($t('loadFailed'), 3000, 'error');
    }
  }

  // Native menu and backend events.
  const appEvents: Record<string, (data: any) => void> = {
    'menu-open-file': handleOpen,
    'menu-open-recent': handleOpenRecent,
    'menu-save-file': () => saveOnce('save'),
    'menu-save-file-as': () => saveOnce('saveAs'),
    'menu-new-tab': addNewTab,
    'menu-new-from-template': () => (showTemplateModal = true),
    'menu-export-pdf': handleExportPdf,
    'menu-export-html': handleExport,
    'menu-print': handlePrint,
    'format-bold': () => wrapSelection('**', '**'),
    'format-italic': () => wrapSelection('*', '*'),
    'format-h1': () => prefixSelection('# '),
    'format-h2': () => prefixSelection('## '),
    'format-h3': () => prefixSelection('### '),
    'format-code': () => wrapSelection('\n```\n', '\n```\n'),
    'set-locale': (l: string) => locale.set(l),
    'menu-about': () => openAbout(),
    'menu-settings': () => { if (settingsSections.length) showSettings = true; },
    'app-close-requested': () => void closeRequested(),
    'set-theme': (theme: string) => appTheme.set(theme as any),
  };

  onMount(() => {
    const init = async () => {
      const version = await backend.getVersion();
      appVersion.set(version);
      backend.mlcdocVersion().then((v) => (mlcdocVersion = v));
      backend.getPDFWatermark().then((w) => (pdfWatermark = w));
      backend.getPDFPreviewNoImages().then((v) => (pdfNoImages = v));
      const wailsReady = checkWailsReady();
      if (wailsReady || import.meta.env.DEV) {
        if (wailsReady) {
          isReady = true;
          for (const [name, fn] of Object.entries(appEvents)) EventsOn(name, fn);
          OnFileDrop((_x: number, _y: number, paths: string[]) => void handleFileDrop(paths), false);
        }

        const result = wailsReady ? await backend.getInitialContent() : null;
        if (result) {
          const title = await backend.getFileTitle(result.path);
          tabs = [createNewTab(title, result.content, result.path)];
          // Opened with a document: read it first, the editor comes on
          // request. Not for a new or empty file (then one wants to write),
          // and only for this start, not as the remembered setting.
          if (result.content.trim() !== '' && !get(isEditorHidden)) {
            startInPreview = true;
            isEditorHidden.set(true);
          }
          const parentDir = await backend.getParentDir(result.path);
          await backend.addPathToWhitelist(parentDir);
        } else {
          tabs = [createNewTab($t('untitled'), defaultMarkdown())];
        }
        activeTabIndex = 0;

        if (import.meta.env.DEV && !wailsReady) isReady = true;

        debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
        if (wailsReady) updateNativeMenu(get(locale));
      } else {
        setTimeout(init, 50);
      }
    };
    init();
    updateEffectiveTheme();
    // Restore editor visibility (last persisted state).
    try {
      if (localStorage.getItem('marksafe-editor-hidden') === '1') isEditorHidden.set(true);
    } catch { /* storage unavailable */ }
    isEditorHidden.subscribe((hidden) => {
      if (startInPreview) return; // hidden for this start only
      try { localStorage.setItem('marksafe-editor-hidden', hidden ? '1' : '0'); } catch { /* ignore */ }
    });
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => { if ($appTheme === 'auto') updateEffectiveTheme(); };
    mediaQuery.addEventListener('change', handler);
    return () => { 
      mediaQuery.removeEventListener('change', handler); 
      OnFileDropOff(); 
      EventsOff(...Object.keys(appEvents));
    };
  });

  // Browse-window mode (?browse=<root>&path=<rel>): this same App boots
  // as a standalone browser window instead of the editor.
  const browseParams: { root: string; path: string | null } | null = (() => {
    try {
      const q = new URLSearchParams(window.location.search);
      const root = q.get('browse');
      if (!root) return null;
      return { root, path: q.get('path') };
    } catch {
      return null;
    }
  })();
  const browseWailsClient = browseParams ? createWailsClient(browseParams.root) : null;

  let showAbout = $state(false);
  let aboutTitle = $state('');
  let aboutMessage = $state('');

  async function handleBrowseExternal(e: CustomEvent<{ url: string; reuse: boolean }>) {
    const { url } = e.detail;
    // Dead links surface here instead of a browser error page. A
    // failing checker must not block the user: fail open.
    let alive = true;
    try {
      alive = await backend.checkExternalURL(url);
    } catch {
      alive = true;
    }
    if (!alive) {
      showToast($t('externalDead'), 4000, 'error');
      return;
    }
    try {
      BrowserOpenURL(url);
    } catch (err) {
      console.error(`Failed to open ${url}:`, err);
    }
  }

  async function toggleBrowseMode() {
    const tab = tabs[activeTabIndex];
    const tabPath = tab?.path;
    if (!tabPath) {
      showToast($t('browseNeedFile'));
      return;
    }
    const rootDir = await backend.getParentDir(tabPath);
    if (!rootDir) {
      showToast($t('browseNeedFile'));
      return;
    }
    // Image/file delivery inside browse mode reuses the whitelist.
    await backend.addPathToWhitelist(rootDir);
    await backend.openBrowseWindow(rootDir, toBrowseRel(rootDir, tabPath) ?? 'index.md');
  }

  function openAbout() {
    const tMap = translations[$locale] || translations.de;
    aboutTitle = tMap.aboutTitle;
    aboutMessage = tMap.aboutBody.replace('%s', get(appVersion));
    showAbout = true;
  }

  // Ctrl+S arrives twice outside macOS: as the menu accelerator and, since
  // Wails passes the key on to the webview as well, as a keydown here.
  // Only the first of the two runs (two "Save As" dialogs otherwise).
  let lastSave = { kind: '', at: 0 };
  function saveOnce(kind: 'save' | 'saveAs') {
    const now = Date.now();
    if (lastSave.kind === kind && now - lastSave.at < 500) return;
    lastSave = { kind, at: now };
    if (kind === 'saveAs') handleSaveAs();
    else handleSave();
  }

  // Started with a document, the editor is hidden for this start only
  // (see init); the first toggle makes it a normal, remembered choice again.
  let startInPreview = false;
  function toggleEditor() {
    startInPreview = false;
    isEditorHidden.update(v => !v);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (browseParams) return; // browse window has its own shortcuts
    // Ctrl+P prints through the PDF; the webview's own print dialog stays
    // closed. On macOS the menu handles Cmd+P.
    if ((e.ctrlKey || e.metaKey) && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'p') {
      // No print dialog of the webview; without mlcdoc there is no printing.
      e.preventDefault();
      if (mlcdocVersion && !/mac/i.test(navigator.platform)) handlePrint();
      return;
    }
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && !e.altKey && e.key.toLowerCase() === 'f') {
      // Field dialog: a new form field at the cursor (MLC extension).
      e.preventDefault();
      insertField();
      return;
    }
    if ((e.ctrlKey || e.metaKey) && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'e') {
      e.preventDefault();
      if (!$isFocusMode) toggleEditor();
      return;
    }
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key.toLowerCase() === 'b') {
      e.preventDefault();
      toggleBrowseMode();
      return;
    }
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
      e.preventDefault();
      saveOnce(e.shiftKey ? 'saveAs' : 'save');
    }
  }

  // --- Svelte 5 Runes: Effects ---
  $effect(() => {
    if (browseParams) return; // BrowseView owns the title there
    if (activeTab) {
      const winTitle = `${activeTab.title}${activeTab.isDirty ? ' *' : ''} - MarkSafe`;
      document.title = winTitle;
      backend.setWindowTitle(winTitle);
    } else {
      document.title = 'MarkSafe';
      backend.setWindowTitle('MarkSafe');
    }
  });

  $effect(() => {
    if ($appTheme) untrack(() => updateEffectiveTheme());
  });

  $effect(() => {
    if ($locale) untrack(() => updateNativeMenu($locale));
  });

  $effect(() => {
    if (isReady && $effectiveAppTheme) {
      const theme = getTheme($effectiveAppTheme);
      debouncedUpdate(markdown, theme.chromaStyle);
    }
  });

  $effect(() => {
    if (isReady && markdown !== undefined) {
      debouncedUpdate(markdown, currentPreviewTheme.chromaStyle);
    }
  });

</script>

<svelte:window onmousemove={onMouseMove} onmouseup={stopResizing} onkeydown={handleKeydown} />

<AboutModal
  show={showAbout}
  title={aboutTitle}
  message={aboutMessage}
  licenses={sensorMonitorLicense ? [sensorMonitorLicense] : []}
  onClose={() => showAbout = false}
  theme={$effectiveAppTheme}
/>

{#if fieldDialog}
  <FieldDialog
    initial={fieldDialog.initial}
    usedNames={(mlcAnalysis?.fields ?? []).filter((f) => f.named).map((f) => f.name)}
    prefill={(mlcAnalysis?.lists?.prefill ?? []).map((p) => ({ name: p.name, label: p.label }))}
    theme={$effectiveAppTheme}
    onApply={applyField}
    onClose={() => (fieldDialog = null)}
  />
{/if}
{#if showSettings}
  <SettingsModal sections={settingsSections} theme={$effectiveAppTheme} onClose={() => (showSettings = false)} />
{/if}
{#if SensorMonitor && isReady}
  <SensorMonitor
    bind:this={sensorMonitor}
    markdown={markdown}
    doc={mlcAnalysis?.document ?? null}
    hasErrors={(mlcAnalysis?.diagnostics ?? []).some((d) => d.severity === 'error')}
    filename={tabs[activeTabIndex]?.path?.split(/[\\/]/).pop() ?? ''}
    theme={$effectiveAppTheme}
    onOpenTab={openModuleTab}
    onStatus={(items: typeof moduleStatus) => (moduleStatus = items)}
    onMenuChanged={() => updateNativeMenu(get(locale))}
  />
{/if}

<NewFromTemplateModal
  show={showTemplateModal}
  theme={$effectiveAppTheme}
  docPath={tabs[activeTabIndex]?.path}
  onCreate={createFromTemplate}
  onCancel={() => (showTemplateModal = false)}
/>

{#if confirmAsk}
  <ConfirmModal
    title={confirmAsk.title}
    message={confirmAsk.message}
    confirmLabel={confirmAsk.confirmLabel}
    cancelLabel={confirmAsk.cancelLabel ?? $t('cancel')}
    altLabel={confirmAsk.altLabel ?? ''}
    onAlt={() => answerConfirm('alt')}
    confirmClass={confirmAsk.altLabel ? 'bg-blue-600 hover:bg-blue-500' : undefined}
    theme={$effectiveAppTheme}
    onConfirm={() => answerConfirm(true)}
    onCancel={() => answerConfirm(false)}
  />
{/if}

<WhitelistModal 
  show={showSecurityModal} 
  type={securityType} 
  resource={securityResource} 
  theme={$effectiveAppTheme}
  onConfirm={confirmSecurityRequest}
  onCancel={() => showSecurityModal = false}
/>

{#if browseParams && browseWailsClient}
  <BrowseView
    client={browseWailsClient}
    theme={$effectiveAppTheme}
    syncUrl={true}
    allowReuseWindow={false}
    initialPath={browseParams.path}
    on:toggleTheme={() => toggleAppTheme()}
    on:openExternal={handleBrowseExternal}
  />
{:else}
<main data-file-drop-target class="flex h-screen w-full overflow-hidden flex-col select-none {$effectiveAppTheme === 'dark' ? 'bg-slate-900' : 'bg-white'}">
  {#if SHOW_HTML_TOOLBAR}
    <HamburgerMenu 
      onFileNew={addNewTab}
      onFileNewTemplate={() => (showTemplateModal = true)}
      onFileOpenLocal={handleOpen}
      onFileOpenRecent={(p) => p && handleOpenRecent(p)}
      onFileSave={handleSave}
      onFileSaveAs={handleSaveAs}
      onEditCopyAsMarkdown={() => handleExport()}
      onEditCopyAsHtml={() => handleExport()}
      onEditCopyAsRTF={() => handleExport()}
      onEditCopyImage={() => handleExport()}
      onEditCopyImageText={() => handleExport()}
      onFormatToggle={toggleEditor}
      onFormatWordCount={() => console.log('wordCount')}
      onViewZoomIn={() => adjustFontSize(5)}
      onViewZoomOut={() => adjustFontSize(-5)}
      onViewResetZoom={() => fontSize = 100}
      onViewBrowse={() => toggleBrowseMode()}
      onHelpAbout={() => openAbout()}
      onOpenSettings={() => { if (settingsSections.length) showSettings = true; }}
      onAboutWails={() => alert('MarkSafe')}
      onOpenProductPage={() => window.open('https://mlcgo.eu/products/marksafe/', '_blank')}
      onHelpShowShortcuts={() => void 0}
    />
  {/if}
  
  <TabsBar tabs={tabs} bind:activeTabIndex={activeTabIndex} onCloseTab={handleCloseTab} />

  <div class="flex flex-1 overflow-hidden relative">
    {#if !$isEditorHidden && !$isFocusMode}
    <div class="flex flex-col min-w-0 border-r relative {editorClass}" style="width: {$splitWidth}%;">
      <div class="h-10 border-b flex items-center px-3 gap-1 shrink-0 {toolbarClass} overflow-x-auto select-none">
        <span class="text-xs font-bold uppercase tracking-wider opacity-50 mr-1 shrink-0">{$t('editor')}</span>
        <div class="h-4 w-px {dividerClass} shrink-0 mx-1"></div>
        <button 
          onclick={() => wrapSelection('**', '**')} 
          class="w-6 h-6 flex items-center justify-center rounded text-xs font-bold hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title="{$t('menuBold')} (Ctrl+B)"
        >B</button>
        <button 
          onclick={() => wrapSelection('*', '*')} 
          class="w-6 h-6 flex items-center justify-center rounded text-xs italic font-serif hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title="{$t('menuItalic')} (Ctrl+I)"
        >I</button>
        <div class="h-4 w-px {dividerClass} shrink-0 mx-0.5"></div>
        <button 
          onclick={() => prefixSelection('# ')} 
          class="px-1.5 h-6 flex items-center justify-center rounded text-[11px] font-bold hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title={$t('formatH1')}
        >H1</button>
        <button 
          onclick={() => prefixSelection('## ')} 
          class="px-1.5 h-6 flex items-center justify-center rounded text-[11px] font-bold hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title={$t('formatH2')}
        >H2</button>
        <button 
          onclick={() => prefixSelection('### ')} 
          class="px-1.5 h-6 flex items-center justify-center rounded text-[11px] font-bold hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title={$t('formatH3')}
        >H3</button>
        <div class="h-4 w-px {dividerClass} shrink-0 mx-0.5"></div>
        <button 
          onclick={() => wrapSelection('\n```\n', '\n```\n')} 
          class="px-1 h-6 flex items-center justify-center rounded text-[11px] font-mono hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title="{$t('menuCodeBlock')} (Ctrl+Shift+C)"
        >&lt;/&gt;</button>
        <button 
          onclick={() => wrapSelection('`', '`')} 
          class="w-6 h-6 flex items-center justify-center rounded text-xs font-mono hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title={$t('formatInlineCode')}
        >`c`</button>
        <div class="h-4 w-px {dividerClass} shrink-0 mx-0.5"></div>
        <button 
          onclick={() => prefixSelection('> ')} 
          class="w-6 h-6 flex items-center justify-center rounded text-xs font-serif font-bold hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title={$t('formatQuote')}
        >&ldquo;</button>
        <button 
          onclick={() => prefixSelection('- ')} 
          class="w-6 h-6 flex items-center justify-center rounded text-xs hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title={$t('formatList')}
        >•-</button>
        <button 
          onclick={() => prefixSelection('- [ ] ')} 
          class="w-6 h-6 flex items-center justify-center rounded text-xs hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title={$t('formatTaskList')}
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 11 12 14 22 4"></polyline><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"></path></svg>
        </button>
        <button 
          onclick={() => wrapSelection('[', '](https://)')} 
          class="w-6 h-6 flex items-center justify-center rounded text-xs hover:bg-black/10 dark:hover:bg-white/10 transition-colors shrink-0 {buttonClass}" 
          title={$t('formatLink')}
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path></svg>
        </button>
        <!-- Tables: insert, column widths for the PDF (also right click, and the box at the separator row) -->
        <details class="mlc-menu relative shrink-0">
          <summary
            class="list-none cursor-pointer px-1.5 h-6 flex items-center rounded text-[11px] font-bold hover:bg-black/10 dark:hover:bg-white/10 {buttonClass}"
            title={$t('tableToolbarTitle')}
          >⊞ ▾</summary>
          <div class="fixed z-50 mt-1 py-1 min-w-64 rounded border shadow-lg text-xs {toolbarClass}">
            <button
              class="w-full text-left px-3 py-1 hover:bg-black/10 dark:hover:bg-white/10"
              onclick={(e) => { editorComponent?.insertText(TABLE_TEMPLATE); (e.currentTarget as HTMLElement).closest('details')?.removeAttribute('open'); }}
            >{$t('tableInsert')}</button>
            <div class="px-3 pt-2 pb-0.5 opacity-60">{$t('tableMenuTitle')}{cursorTable ? '' : ' · ' + $t('tableNeedCursor')}</div>
            {#each [['auto', 'tableFitAuto'], ['content', 'tableFitContent']] as [mode, label]}
              <button
                class="w-full text-left px-3 py-1 hover:bg-black/10 dark:hover:bg-white/10 disabled:opacity-40 disabled:hover:bg-transparent"
                disabled={!cursorTable}
                onclick={(e) => { editorComponent?.fitTableAtCursor(mode as 'auto' | 'content'); (e.currentTarget as HTMLElement).closest('details')?.removeAttribute('open'); }}
              >{$t(label)}</button>
            {/each}
          </div>
        </details>
        <!-- Symbols that print in the PDF (mlcext.PDFSymbols: the fonts of mlcdoc) -->
        {#if pdfSymbols.length}
        <details class="mlc-menu relative shrink-0">
          <summary
            class="list-none cursor-pointer px-1.5 h-6 flex items-center rounded text-[11px] font-bold hover:bg-black/10 dark:hover:bg-white/10 {buttonClass}"
            title={$t('symbolInsertTitle')}
          >☺ ▾</summary>
          <div class="fixed z-50 mt-1 p-2 w-72 rounded border shadow-lg text-xs {toolbarClass}">
            {#each pdfSymbols as g}
              <div class="opacity-60 mt-1 mb-0.5">{g.name}</div>
              <div class="flex flex-wrap gap-0.5">
                {#each g.symbols as sym}
                  <button
                    class="w-7 h-7 rounded text-base hover:bg-black/10 dark:hover:bg-white/10"
                    title={sym}
                    onclick={(e) => { editorComponent?.insertText(sym); (e.currentTarget as HTMLElement).closest('details')?.removeAttribute('open'); }}
                  >{sym}</button>
                {/each}
              </div>
            {/each}
            <div class="opacity-50 mt-2 leading-snug">{$t('symbolInsertNote')}</div>
          </div>
        </details>
        {/if}
        <!-- Alert boxes (GitHub/GitLab "> [!NOTE]" …; also boxes in the PDF) -->
        <details class="mlc-menu relative shrink-0">
          <summary
            class="list-none cursor-pointer px-1.5 h-6 flex items-center rounded text-[11px] font-bold hover:bg-black/10 dark:hover:bg-white/10 {buttonClass}"
            title={$t('alertInsertTitle')}
          >! ▾</summary>
          <div class="fixed z-50 mt-1 py-1 min-w-44 rounded border shadow-lg text-xs {toolbarClass}">
            {#each ALERTS as a}
              <button
                class="w-full text-left px-3 py-1 flex items-center gap-2 hover:bg-black/10 dark:hover:bg-white/10"
                onclick={(e) => { editorComponent?.insertAlert(a.kind); (e.currentTarget as HTMLElement).closest('details')?.removeAttribute('open'); }}
              ><span class="inline-block w-1 h-3.5 rounded-sm" style="background: {a.color}"></span><span>{$t(a.label)}</span><span class="ml-auto opacity-50 font-mono">[!{a.kind}]</span></button>
            {/each}
          </div>
        </details>
        {#if mlcMenu.length > 0}
          <!-- MLC Markdown Extension (non-standard): only for MLC documents -->
          <div class="h-4 w-px {dividerClass} shrink-0 mx-0.5"></div>
          <details class="mlc-menu relative shrink-0">
            <summary
              class="list-none cursor-pointer px-1.5 h-6 flex items-center rounded text-[11px] font-bold hover:bg-black/10 dark:hover:bg-white/10 {buttonClass}"
              title={$t('mlcInsertTitle')}
            >MLC ▾</summary>
            <div class="fixed z-50 mt-1 py-1 min-w-56 rounded border shadow-lg text-xs {toolbarClass}">
              <button
                class="w-full text-left px-3 py-1 flex justify-between gap-4 font-semibold hover:bg-black/10 dark:hover:bg-white/10"
                onclick={(e) => { insertField(); (e.currentTarget as HTMLElement).closest('details')?.removeAttribute('open'); }}
              ><span>{$t('fieldInsert')}</span><span class="opacity-60 font-mono">Ctrl+Shift+F</span></button>
              <div class="my-1 border-t opacity-30"></div>
              {#each mlcMenu as item}
                <button
                  class="w-full text-left px-3 py-1 flex justify-between gap-4 hover:bg-black/10 dark:hover:bg-white/10"
                  onclick={(e) => { editorComponent?.insertSnippet(item); (e.currentTarget as HTMLElement).closest('details')?.removeAttribute('open'); }}
                ><span>{item.label}</span><span class="opacity-60 font-mono">{item.detail}</span></button>
              {/each}
            </div>
          </details>
        {/if}
      </div>
      {#if tabs[activeTabIndex]}
      <div class="flex-1 min-h-0 relative">
        <Editor
          bind:this={editorComponent}
          bind:value={tabs[activeTabIndex].content}
          theme={$effectiveAppTheme}
          placeholder={$t('placeholder')}
          onchange={onContentInput}
          onscroll={handleEditorScroll}
          {mlcSource}
          tableFit={backend.mlcFitTable}
          oncursortable={(l: number) => (cursorTable = l)}
          oneditfield={editField}
        />
      </div>
      {/if}
      <div 
        role="slider"
        aria-label={$t('resizeSplitter')}
        aria-valuenow={$splitWidth}
        aria-valuemin={10}
        aria-valuemax={90}
        tabindex="0"
        onmousedown={startResizing} 
        onkeydown={(e) => {
          if (e.key === 'ArrowLeft') splitWidth.set(Math.max(10, $splitWidth - 1));
          if (e.key === 'ArrowRight') splitWidth.set(Math.min(90, $splitWidth + 1));
        }}
        class="absolute top-0 right-0 w-1 h-full cursor-col-resize hover:bg-blue-500/50 transition-colors z-10"
      ></div>
    </div>
    {/if}

    <div class="flex-1 min-w-0 flex flex-col relative">
      {#if isResizing} <div class="absolute inset-0 z-50"></div> {/if}
      <div class="p-2 h-10 border-b flex flex-nowrap items-center px-4 gap-4 shrink-0 whitespace-nowrap overflow-hidden {toolbarClass}">
        <div class="flex items-center gap-2">
            <!-- Focus mode always hides the editor, so the editor toggle would do nothing there -->
            {#if !$isFocusMode}
            <!-- With the editor hidden it is labelled "Edit": the way back has to be easy to find -->
            <button
                onclick={toggleEditor}
                class="p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors flex items-center gap-1 {$isEditorHidden ? focusButtonClass : ''}"
                title={($isEditorHidden ? $t('showEditor') : $t('hideEditor')) + ' (' + (/mac/i.test(navigator.platform) ? '⌘' : 'Ctrl+') + 'E)'}
            >
                {#if $isEditorHidden}
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"></path><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path></svg>
                <span class="text-xs">{$t('editDocument')}</span>
                {:else}
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="9" y1="3" x2="9" y2="21"></line></svg>
                {/if}
            </button>
            {/if}
            <button 
                onclick={() => isFocusMode.update(v => !v)} 
                class="p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors {$isFocusMode ? focusButtonClass : ''}"
                title={$isFocusMode ? $t('exitFocusMode') : $t('focusMode')}
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"></path></svg>
            </button>
            <button 
                onclick={() => isSyncScroll = !isSyncScroll} 
                class="p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors {isSyncScroll ? focusButtonClass : ''}"
                title={$t('syncScroll')}
            >
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 15l5 5 5-5"/><path d="M7 9l5-5 5 5"/></svg>
            </button>
        </div>
        <div class="text-xs font-bold uppercase tracking-wider opacity-50">{$t('preview')}</div>
        <div class="flex-1"></div>
        <!-- Browsing needs a folder: only for a saved file. -->
        {#if tabs[activeTabIndex]?.path}
        <button onclick={() => toggleBrowseMode()} title={$t('browserModus')} aria-label={$t('browserModus')} class="p-1 rounded transition-colors {buttonClass}">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" /><line x1="9" y1="3" x2="9" y2="21" /></svg>
        </button>
        {/if}
        {#if canFill}
          <button onclick={startFill} title={$t('formFillStartHint')} class="px-1.5 h-6 rounded text-[11px] font-bold transition-colors {buttonClass}">{$t('formFillStart')}</button>
        {/if}
        {#if canPdfPreview}
          <button
            onclick={() => (previewMode = previewMode === 'pdf' ? 'html' : 'pdf')}
            title={$t('pdfPreviewToggle')}
            class="px-1.5 h-6 rounded text-[11px] font-bold transition-colors {buttonClass} {showPdfPreview ? 'ring-1 ring-current' : ''}"
          >PDF</button>
        {/if}
        <!-- Printing goes through the PDF: only with mlcdoc -->
        {#if mlcdocVersion}
        <button onclick={handlePrint} title={$t('print')} class="p-1 rounded transition-colors {buttonClass}">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 6 2 18 2 18 9"></polyline><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"></path><rect x="6" y="14" width="12" height="8"></rect></svg>
        </button>
        {/if}
        <div class="h-4 w-px {dividerClass}"></div>
        <button 
          onclick={toggleAppTheme} 
          title={$t('toggleTheme') + ' (' + ($appTheme === 'auto' ? $t('menuThemeAuto') : ($appTheme === 'dark' ? $t('menuThemeDark') : $t('menuThemeLight'))) + ')'}
          class="flex items-center gap-1.5 px-2 py-0.5 rounded text-xs transition-colors {buttonClass}"
        >
          {#if $appTheme === 'dark'}
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
            <span>{$t('menuThemeDark')}</span>
          {:else if $appTheme === 'light'}
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"></circle><line x1="12" y1="1" x2="12" y2="3"></line><line x1="12" y1="21" x2="12" y2="23"></line><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line><line x1="1" y1="12" x2="3" y2="12"></line><line x1="21" y1="12" x2="23" y2="12"></line><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line></svg>
            <span>{$t('menuThemeLight')}</span>
          {:else}
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
            <span>{$t('menuThemeAuto')}</span>
          {/if}
        </button>
        <!-- Zoom scales the HTML preview only; the PDF pages come as images. -->
        {#if !showPdfPreview}
        <div class="h-4 w-px {dividerClass}"></div>
        <div class="flex gap-1 items-center">
          <button onclick={() => adjustFontSize(-5)} class="w-5 h-5 flex items-center justify-center rounded text-xs font-bold {buttonClass}">-</button>
          <span class="text-[10px] opacity-60 w-8 text-center font-mono">{fontSize}%</span>
          <button onclick={() => adjustFontSize(5)} class="w-5 h-5 flex items-center justify-center rounded text-xs font-bold {buttonClass}">+</button>
        </div>
        {/if}
      </div>
      {#if showPdfPreview}
        <PdfPreview markdown={tabs[activeTabIndex]?.content ?? ''} path={tabs[activeTabIndex]?.path} version={mlcdocVersion} {hasMath} watermark={pdfWatermark} onwatermark={changeWatermark} noImages={pdfNoImages} onnoimages={changeNoImages} />
      {/if}
      {#if filling && !showPdfPreview}
        <FormFillPanel
          md={tabs[activeTabIndex].content}
          {filledBy}
          theme={$effectiveAppTheme}
          onchange={fillChanged}
          onfilledby={setFilledBy}
          onsave={() => saveFilledForm()}
          onload={loadFillValues}
          onclose={() => { const f = tabs[activeTabIndex].form; if (f) f.hidden = true; }}
        />
      {/if}
      <div class="flex flex-1 min-h-0 flex-col" class:hidden={showPdfPreview || filling}>
      <Preview 
        bind:this={previewComponent}
        source={markdown}
        theme={currentPreviewTheme} 
        fontSize={fontSize} 
        currentFilePath={tabs[activeTabIndex]?.path}
        analysis={mlcAnalysis}
        onsecurity_request={handleSecurityRequest}
        onopen_file={handleOpenExternalMD}
        onscrollline={handlePreviewScroll}
        onrendered={(el: HTMLElement) => (hasMath = !!el.querySelector('.mlc-md-math'))}
      />
      </div>
    </div>
  </div>

  <StatusBar {wordCount} {charCount} {readingTime} activeTab={tabs[activeTabIndex]} diagnostics={mlcAnalysis?.diagnostics ?? []} extraItems={moduleStatus} />

  {#if $dropMessage}
    <div class="drop-toast {$toastType === 'error' ? 'error' : ''}">{$dropMessage}</div>
  {/if}
</main>

{/if}

<style>
  :global(.browse-view) {
    flex: 1 1 auto;
    min-width: 0;
    overflow-y: auto;
  }

  :global(body) { margin: 0; }
  .no-scrollbar::-webkit-scrollbar { display: none; }
  .no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
  .drop-toast {
    position: absolute;
    bottom: 3.5rem;
    left: 50%;
    transform: translateX(-50%);
    background: #3b82f6;
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 9999px;
    font-size: 0.875rem;
    font-weight: 500;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
    z-index: 50;
    animation: slideUp 0.3s ease-out;
  }
  .drop-toast.error { background: #ef4444; }
  @keyframes slideUp { from { transform: translate(-50%, 100%); opacity: 0; } to { transform: translate(-50%, 0); opacity: 1; } }

</style>