<script lang="ts">
  /**
   * MarkSafe's Markdown editor: the shared <MlcEditor> (packages/
   * mlcext-editor-svelte: CodeMirror, MLC extension, themes, scroll sync)
   * plus what only MarkSafe has — context menus for fields, images and
   * tables, the table and image boxes, folded data: URLs, pasted pictures,
   * CodeMirror texts in the interface language.
   */
  import { untrack } from 'svelte';
  import { EditorView } from '@codemirror/view';
  import { MlcEditor } from 'mlcext-editor-svelte';
  import { fieldAt, type MlcSource, type MlcAnalysis, type MlcDiagnostic } from 'mlcext-editor';
  import type { EffectiveTheme_t } from '../lib/constants';
  import { editorPhrases } from '../lib/editor/phrases';
  import { dataUriFolding } from '../lib/editor/dataUri';
  import { tableHint, tableAtCursor } from '../lib/editor/tableHint';
  import { imageHint, imageAt } from '../lib/editor/imageHint';
  import type { ImageSpan } from 'mlc-markdown';
  import { locale, t as tr } from '../i18n';

  let {
    value = $bindable(''),
    theme = 'dark',
    placeholder = '',
    onchange,
    onscroll,
    mlcSource,
    onmlcanalysis,
    tableFit,
    oncursortable,
    oneditfield,
    oneditimage,
    onpasteimage,
    extraDiagnostics,
    diagnosticsRevision = 0,
  }: {
    value: string;
    theme: EffectiveTheme_t;
    placeholder?: string;
    onchange?: (val: string) => void;
    /** Top visible source line (1-based) while scrolling, for scroll sync. */
    onscroll?: (line: number) => void;
    /** MLC Markdown Extension (non-standard) support: analysis from mlcext. */
    mlcSource?: MlcSource;
    onmlcanalysis?: (analysis: MlcAnalysis) => void;
    /** Column widths of the table at line (mlcext.FitTableColumns); null: no table. */
    tableFit?: (md: string, line: number, mode: 'auto' | 'content') => Promise<string | null>;
    /** Separator row of the table at the cursor (0: none), for the toolbar. */
    oncursortable?: (line: number) => void;
    /** Open the field dialog for the form field at line, col (MLC extension). */
    oneditfield?: (line: number, col: number) => void;
    /** Open the image dialog for an image ![alt](target) (offsets in the document). */
    oneditimage?: (img: ImageSpan) => void;
    /** A picture was pasted (data: URL); text on the clipboard goes first. */
    onpasteimage?: (dataUri: string) => void;
    /** Findings from outside the extension (folder check), shown inline. */
    extraDiagnostics?: (md: string) => MlcDiagnostic[];
    /** Bump to show changed extraDiagnostics at once. */
    diagnosticsRevision?: number;
  } = $props();

  type EditorApi = {
    getView(): EditorView | null;
    focus(): void;
    getTopLine(): number;
    scrollToLine(line: number): void;
    insertText(text: string): void;
    replaceRange(line: number, col: number, endLine: number, endCol: number, text: string): void;
    replaceOffsets(from: number, to: number, text: string): void;
    wrapSelection(prefix: string, suffix?: string): void;
    prefixSelection(prefix: string): void;
    refresh(): void;
  };
  let editor = $state<EditorApi>();
  const view = () => editor?.getView() ?? null;

  $effect(() => {
    if (diagnosticsRevision === 0) return;
    untrack(() => editor?.refresh());
  });

  // Right-click on a form field: edit it in the field dialog.
  let fieldMenu = $state<{ x: number; y: number; line: number; col: number } | null>(null);

  // Right-click on an image: edit it in the image dialog.
  let imageMenu = $state<{ x: number; y: number; image: ImageSpan } | null>(null);

  // Right-click in a table: column widths for the PDF.
  let tableMenu = $state<{ x: number; y: number; line: number } | null>(null);
  async function fitTable(mode: 'auto' | 'content') {
    const menu = tableMenu;
    tableMenu = null;
    const v = view();
    if (!menu || !v || !tableFit) return;
    const text = v.state.doc.toString();
    const fixed = await tableFit(text, menu.line, mode);
    if (fixed == null || v.state.doc.toString() !== text) return;
    v.dispatch({ changes: { from: 0, to: text.length, insert: fixed } });
    v.focus();
  }

  /** getTopLine returns the first (1-based) source line visible at the top. */
  export function getTopLine(): number {
    return editor?.getTopLine() ?? 1;
  }

  /** scrollToLine puts the 1-based source line at the top. */
  export function scrollToLine(line: number) {
    editor?.scrollToLine(line);
  }

  /** wrapSelection wraps selected text with prefix/suffix (bold, italic, code). */
  export function wrapSelection(prefix: string, suffix: string) {
    editor?.wrapSelection(prefix, suffix);
  }

  /** prefixSelection prefixes the current line with markdown token (# , ## ). */
  export function prefixSelection(prefix: string) {
    editor?.prefixSelection(prefix);
  }

  /** Column widths of the table at the cursor (toolbar menu). */
  export function fitTableAtCursor(mode: 'auto' | 'content') {
    const v = view();
    if (!v) return;
    const line = tableAtCursor(v.state);
    if (!line) return;
    tableMenu = { x: 0, y: 0, line };
    void fitTable(mode);
  }

  /** Replaces the range (1-based lines, columns in characters, end exclusive), e.g. a field from the field dialog. */
  export function replaceRange(line: number, col: number, endLine: number, endCol: number, text: string) {
    editor?.replaceRange(line, col, endLine, endCol, text);
  }

  /** replaceOffsets replaces from..to (offsets in the document) with text. */
  export function replaceOffsets(from: number, to: number, text: string) {
    editor?.replaceOffsets(from, to, text);
  }

  /** insertText puts text at the cursor (replacing a selection). */
  export function insertText(text: string) {
    editor?.insertText(text);
  }

  /**
   * insertAlert turns the selected lines (or the current line) into an alert
   * box "> [!WARNING]" (GitHub/GitLab; mlcdoc sets it as a box too).
   */
  export function insertAlert(kind: 'NOTE' | 'TIP' | 'IMPORTANT' | 'WARNING' | 'CAUTION') {
    const v = view();
    if (!v) return;
    const state = v.state;
    const { from, to } = state.selection.main;
    const first = state.doc.lineAt(from);
    const last = state.doc.lineAt(to);
    const lines: string[] = [];
    for (let n = first.number; n <= last.number; n++) {
      lines.push('> ' + state.doc.line(n).text.replace(/^>\s?/, ''));
    }
    const insert = `> [!${kind}]\n` + lines.join('\n');
    v.dispatch({ changes: { from: first.from, to: last.to, insert }, selection: { anchor: first.from + insert.length } });
    v.focus();
  }

  /** Focus the CodeMirror editor. */
  export function focus() {
    editor?.focus();
  }

  // Pasted pictures and the context menus (field, image, table).
  const events = EditorView.domEventHandlers({
    paste(e) {
      const data = e.clipboardData;
      if (!onpasteimage || !data || data.getData('text/plain')) return false;
      const file = Array.from(data.files).find((f) => f.type.startsWith('image/'));
      if (!file) return false;
      e.preventDefault();
      const reader = new FileReader();
      reader.onload = () => onpasteimage?.(String(reader.result));
      reader.readAsDataURL(file);
      return true;
    },
    contextmenu(e, v) {
      const pos = v.posAtCoords({ x: e.clientX, y: e.clientY });
      if (pos == null) return false;
      const field = oneditfield ? fieldAt(v.state, pos) : null;
      if (field) {
        e.preventDefault();
        fieldMenu = { x: e.clientX, y: e.clientY, line: field.line, col: field.col };
        return true;
      }
      const image = oneditimage ? imageAt(v.state, pos) : null;
      if (image) {
        e.preventDefault();
        imageMenu = { x: e.clientX, y: e.clientY, image };
        return true;
      }
      if (!tableFit) return false;
      const line = v.state.doc.lineAt(pos);
      if (!line.text.includes('|')) return false;
      e.preventDefault();
      tableMenu = { x: e.clientX, y: e.clientY, line: line.number };
      return true;
    },
  });

  // MarkSafe's own extensions; CodeMirror's texts and the boxes follow the
  // interface language, so they are rebuilt when it changes.
  const extensions = $derived.by(() => {
    const loc = $locale;
    return [
      events,
      ...(tableFit
        ? [tableHint({
            label: '⇔ ' + $tr('tableHintLabel'),
            title: $tr('tableMenuTitle'),
            open: (x, y, line) => (tableMenu = { x, y, line }),
            onCursorTable: (line) => oncursortable?.(line),
          })]
        : []),
      ...(oneditimage
        ? [imageHint({ label: '✎ ' + $tr('imageHintLabel'), title: $tr('imageHintTitle'), onEdit: (img) => oneditimage?.(img) })]
        : []),
      editorPhrases(loc),
      dataUriFolding({ image: $tr('dataUriImage'), expandHint: $tr('dataUriExpand'), collapseHint: $tr('dataUriCollapse'), locale: loc }),
    ];
  });
</script>

<div class="h-full w-full overflow-hidden select-text">
  <MlcEditor
    bind:this={editor}
    bind:value
    theme={theme === 'dark' ? 'dark' : 'light'}
    locale={$locale}
    {placeholder}
    source={mlcSource}
    {extraDiagnostics}
    onAnalysis={(a) => onmlcanalysis?.(a)}
    onEditField={mlcSource && oneditfield ? (l, c) => oneditfield?.(l, c) : undefined}
    fieldHintLabel={'✎ ' + $tr('fieldHintLabel')}
    fieldHintTitle={$tr('fieldHintTitle')}
    {extensions}
    {onchange}
    onScrollLine={(line) => onscroll?.(line)}
  />
</div>

{#if fieldMenu}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-[90]" onclick={() => (fieldMenu = null)} oncontextmenu={(e) => { e.preventDefault(); fieldMenu = null; }}></div>
  <div class="table-menu fixed z-[91] py-1 rounded border shadow-lg text-xs {theme === 'light' ? 'tm-light' : ''}" style="left: {fieldMenu.x}px; top: {fieldMenu.y}px">
    <button class="w-full text-left px-3 py-1" onclick={() => { const m = fieldMenu; fieldMenu = null; if (m) oneditfield?.(m.line, m.col); }}>{$tr('fieldEdit')}</button>
  </div>
{/if}

{#if imageMenu}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-[90]" onclick={() => (imageMenu = null)} oncontextmenu={(e) => { e.preventDefault(); imageMenu = null; }}></div>
  <div class="table-menu fixed z-[91] py-1 rounded border shadow-lg text-xs {theme === 'light' ? 'tm-light' : ''}" style="left: {imageMenu.x}px; top: {imageMenu.y}px">
    <button class="w-full text-left px-3 py-1" onclick={() => { const m = imageMenu; imageMenu = null; if (m) oneditimage?.(m.image); }}>{$tr('imageEdit')}</button>
  </div>
{/if}

{#if tableMenu}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-[90]" onclick={() => (tableMenu = null)} oncontextmenu={(e) => { e.preventDefault(); tableMenu = null; }}></div>
  <div class="table-menu fixed z-[91] py-1 rounded border shadow-lg text-xs {theme === 'light' ? 'tm-light' : ''}" style="left: {tableMenu.x}px; top: {tableMenu.y}px">
    <div class="px-3 py-1 opacity-60">{$tr('tableMenuTitle')}</div>
    <button class="w-full text-left px-3 py-1" onclick={() => fitTable('auto')}>{$tr('tableFitAuto')}</button>
    <button class="w-full text-left px-3 py-1" onclick={() => fitTable('content')}>{$tr('tableFitContent')}</button>
  </div>
{/if}

<style>
  .table-menu { background: var(--tm-bg, #1f2937); color: var(--tm-fg, #f3f4f6); border-color: rgba(127,127,127,.4); min-width: 16rem; }
  .table-menu button:hover { background: rgba(127,127,127,.2); }
  .table-menu.tm-light { --tm-bg: #ffffff; --tm-fg: #111827; }
</style>
