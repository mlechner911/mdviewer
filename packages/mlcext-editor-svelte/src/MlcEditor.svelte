<script lang="ts">
  // <MlcEditor> — a CodeMirror 6 Markdown editor with the MLC Markdown
  // Extension (non-standard), built on mlcext-editor. It implements no
  // rules: the analysis comes from the host's `source` (mlcext.Analyze),
  // findings and completions of the host come in through `diagnostics` and
  // `completions`. What only one app needs (menus, dialogs, its own
  // CodeMirror extensions) stays in the app and comes in via `extensions`.
  import { onMount, untrack } from 'svelte';
  import { basicSetup } from 'codemirror';
  import { Compartment, EditorState, type Extension } from '@codemirror/state';
  import { EditorView, placeholder as placeholderExt } from '@codemirror/view';
  import { syntaxHighlighting } from '@codemirror/language';
  import { markdown } from '@codemirror/lang-markdown';
  import {
    fieldHint,
    insertMlcSnippet,
    mlcExtension,
    refreshAnalysis,
    setMlcDiagnostics,
    toPos,
    type MlcAnalysis,
    type MlcCompletionContext,
    type MlcCompletionItem,
    type MlcDiagnostic,
    type MlcSnippet,
    type MlcSource,
  } from 'mlcext-editor';
  import { themes } from './theme.js';

  type Theme = 'light' | 'dark';

  let {
    value = $bindable(''),
    theme = 'light',
    locale = 'en',
    source,
    completions,
    diagnostics,
    extraDiagnostics,
    onAnalysis,
    onEditField,
    fieldHintLabel,
    fieldHintTitle,
    placeholder = '',
    readOnly = false,
    lineWrapping = true,
    extensions = [],
    onchange,
    onScrollLine,
    onReady,
    class: className = '',
  }: {
    /** The Markdown text; bind:value keeps it in sync both ways. */
    value?: string;
    theme?: Theme;
    /** Language of schema descriptions and of the field box. */
    locale?: string;
    /** mlcext.Analyze of the host; without it the editor still completes and shows `diagnostics`. */
    source?: MlcSource;
    completions?: (context: MlcCompletionContext) => MlcCompletionItem[] | null | Promise<MlcCompletionItem[] | null>;
    /** Findings of the host, shown next to the extension's; replaced whenever the array changes. */
    diagnostics?: MlcDiagnostic[];
    /** Findings pulled on every analysis (see mlcext-editor); `refresh()` shows changes. */
    extraDiagnostics?: (markdown: string) => MlcDiagnostic[];
    onAnalysis?: (analysis: MlcAnalysis) => void;
    /** Field dialog of the host: box at the field under the cursor and double click. */
    onEditField?: (line: number, col: number) => void;
    fieldHintLabel?: string;
    fieldHintTitle?: string;
    placeholder?: string;
    readOnly?: boolean;
    lineWrapping?: boolean;
    /** CodeMirror extensions of the host; reconfigured when the array changes. */
    extensions?: Extension[];
    onchange?: (value: string) => void;
    /** Top visible source line (1-based) while the user scrolls; for scroll sync with <MlcMarkdown>. */
    onScrollLine?: (line: number) => void;
    onReady?: (view: EditorView) => void;
    class?: string;
  } = $props();

  const NOOP_SOURCE: MlcSource = { analyze: async () => null };

  let container = $state<HTMLDivElement>();
  let view: EditorView | null = null;
  let fromProp = false;
  let scrollLock = false;

  const themeC = new Compartment();
  const mlcC = new Compartment();
  const hostC = new Compartment();
  const optionsC = new Compartment();

  const german = () => locale.toLowerCase().startsWith('de');

  function mlcPart(src: MlcSource | undefined, withFieldHint: boolean): Extension {
    return [
      mlcExtension(src ?? NOOP_SOURCE, {
        locale: () => locale,
        onAnalysis: (a) => onAnalysis?.(a),
        extraDiagnostics: (md) => extraDiagnostics?.(md) ?? [],
        // Read on every request: the host may swap its function.
        completions: (ctx) => completions?.(ctx) ?? null,
      }),
      withFieldHint
        ? fieldHint({
            label: fieldHintLabel ?? (german() ? '✎ Feld' : '✎ Field'),
            title: fieldHintTitle ?? (german() ? 'Feld bearbeiten (Doppelklick)' : 'Edit field (double click)'),
            onEdit: (line, col) => onEditField?.(line, col),
          })
        : [],
    ];
  }

  function themePart(mode: Theme): Extension {
    const t = themes[mode === 'dark' ? 'dark' : 'light'];
    return [t.editor, syntaxHighlighting(t.highlight)];
  }

  function optionsPart(ph: string, ro: boolean, wrap: boolean): Extension {
    return [
      ph ? placeholderExt(ph) : [],
      EditorState.readOnly.of(ro),
      EditorView.editable.of(!ro),
      wrap ? EditorView.lineWrapping : [],
    ];
  }

  /** The CodeMirror view, e.g. for the host's own commands. */
  export function getView(): EditorView | null {
    return view;
  }

  export function focus() {
    view?.focus();
  }

  /** The first (1-based) source line visible at the top. */
  export function getTopLine(): number {
    if (!view) return 1;
    const block = view.lineBlockAtHeight(Math.max(0, view.scrollDOM.scrollTop - (view.documentPadding?.top ?? 0)));
    return view.state.doc.lineAt(block.from).number;
  }

  /** Puts the 1-based source line at the top; does not echo onScrollLine. */
  export function scrollToLine(line: number) {
    if (!view) return;
    const n = Math.min(Math.max(1, Math.round(line)), view.state.doc.lines);
    const block = view.lineBlockAt(view.state.doc.line(n).from);
    scrollLock = true;
    view.scrollDOM.scrollTop = block.top + (view.documentPadding?.top ?? 0);
    setTimeout(() => (scrollLock = false), 50);
  }

  /** Puts text at the cursor, replacing a selection. */
  export function insertText(text: string) {
    if (!view) return;
    const { from, to } = view.state.selection.main;
    view.dispatch({ changes: { from, to, insert: text }, selection: { anchor: from + text.length } });
    view.focus();
  }

  /** Replaces a range: 1-based lines, columns in code points, end exclusive (as in the analysis). */
  export function replaceRange(line: number, col: number, endLine: number, endCol: number, text: string) {
    if (!view) return;
    const from = toPos(view.state.doc, line, col);
    const to = toPos(view.state.doc, endLine, endCol);
    replaceOffsets(from, to, text);
  }

  /** Replaces from..to (offsets in the document). */
  export function replaceOffsets(from: number, to: number, text: string) {
    if (!view) return;
    view.dispatch({ changes: { from, to, insert: text }, selection: { anchor: from + text.length } });
    view.focus();
  }

  /** Wraps the selection, e.g. "**" for bold. */
  export function wrapSelection(prefix: string, suffix = prefix) {
    if (!view) return;
    const { from, to } = view.state.selection.main;
    const text = view.state.sliceDoc(from, to);
    view.dispatch({
      changes: { from, to, insert: prefix + text + suffix },
      selection: { anchor: from + prefix.length, head: to + prefix.length },
    });
    view.focus();
  }

  /** Puts prefix at the start of the cursor's line, e.g. "## ". */
  export function prefixSelection(prefix: string) {
    if (!view) return;
    const { from } = view.state.selection.main;
    const line = view.state.doc.lineAt(from);
    view.dispatch({ changes: { from: line.from, insert: prefix }, selection: { anchor: from + prefix.length } });
    view.focus();
  }

  /** Inserts a building block of the extension (mlcSnippets) at the selection. */
  export function insertSnippet(snippet: MlcSnippet) {
    if (view) insertMlcSnippet(view, snippet);
  }

  /** Analyse again now, e.g. when extraDiagnostics changed. */
  export function refresh() {
    view?.dispatch({ effects: refreshAnalysis.of(null) });
  }

  function handleScroll() {
    if (!scrollLock && view) onScrollLine?.(getTopLine());
  }

  onMount(() => {
    if (!container) return;
    view = new EditorView({
      parent: container,
      state: EditorState.create({
        doc: value,
        extensions: [
          basicSetup,
          markdown(),
          optionsC.of(optionsPart(placeholder, readOnly, lineWrapping)),
          mlcC.of(mlcPart(source, !!onEditField)),
          hostC.of(extensions),
          themeC.of(themePart(theme)),
          EditorView.updateListener.of((u) => {
            if (!u.docChanged || fromProp) return;
            const text = u.state.doc.toString();
            value = text;
            onchange?.(text);
          }),
        ],
      }),
    });
    mlcKey = keyOfMlc();
    const scroller = view.scrollDOM;
    scroller.addEventListener('scroll', handleScroll, { passive: true });
    if (diagnostics) setMlcDiagnostics(view, diagnostics);
    onReady?.(view);
    return () => {
      scroller.removeEventListener('scroll', handleScroll);
      view?.destroy();
      view = null;
    };
  });

  // Changes of the props reconfigure only their part of the editor.
  $effect(() => {
    const part = themePart(theme);
    untrack(() => view?.dispatch({ effects: themeC.reconfigure(part) }));
  });
  $effect(() => {
    const part = optionsPart(placeholder, readOnly, lineWrapping);
    untrack(() => view?.dispatch({ effects: optionsC.reconfigure(part) }));
  });
  // The MLC part is rebuilt only when the source, the field box or its texts
  // change: a rebuild starts a fresh analysis.
  let mlcKey: unknown[] = [];
  const keyOfMlc = () => [source, !!onEditField, fieldHintLabel, fieldHintTitle, german()];
  $effect(() => {
    const key = keyOfMlc();
    untrack(() => {
      if (!view || key.every((k, i) => k === mlcKey[i])) return;
      mlcKey = key;
      view.dispatch({ effects: mlcC.reconfigure(mlcPart(source, !!onEditField)) });
    });
  });
  $effect(() => {
    const ext = extensions;
    untrack(() => view?.dispatch({ effects: hostC.reconfigure(ext) }));
  });
  $effect(() => {
    const list = diagnostics;
    untrack(() => {
      if (view && list) setMlcDiagnostics(view, list);
    });
  });

  // Text set from outside (another file, a dialog) goes into the editor.
  $effect(() => {
    const next = value;
    untrack(() => {
      if (!view || next === undefined) return;
      const current = view.state.doc.toString();
      if (current === next) return;
      fromProp = true;
      view.dispatch({ changes: { from: 0, to: current.length, insert: next } });
      fromProp = false;
    });
  });
</script>

<div bind:this={container} class="mlc-editor {className}" data-theme={theme}></div>

<style>
  .mlc-editor {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    overflow: hidden;
  }
  .mlc-editor :global(.cm-editor) {
    height: 100%;
    outline: none;
  }
  .mlc-editor :global(.cm-scroller) {
    overflow: auto;
    font-family: inherit;
  }
</style>
