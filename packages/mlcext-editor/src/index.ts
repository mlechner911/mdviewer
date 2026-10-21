// CodeMirror 6 support for the MLC Markdown Extension.
//
// The MLC Markdown Extension is NOT standard Markdown: form fields "[@…]",
// field tooltips "<!--* … -->", fill hints "<!--! … -->", document header
// lines and the upload front matter of the MLC Sensor Monitor
// (docs/MLC-EXTENSIONS.md).
//
// This package implements no rules. Everything it shows comes from the Go
// checker mlcext (Analyze), handed in by the host through an MlcSource:
// MarkSafe calls it over Wails, a web frontend through its server. So there
// is exactly one implementation of the rules, and this package only renders
// its answer: field and hint highlighting, diagnostics, completion and
// snippets for the shared lists.

import { StateEffect, StateField, EditorState, type Extension, type Text } from '@codemirror/state';
import { Decoration, EditorView, ViewPlugin, type DecorationSet, type ViewUpdate } from '@codemirror/view';
import { lintGutter, setDiagnostics, type Diagnostic as CMDiagnostic } from '@codemirror/lint';
import { snippet, type Completion, type CompletionContext, type CompletionResult } from '@codemirror/autocomplete';

// ---------- The analysis (JSON of mlcext.Analysis) ----------

export type Severity = 'error' | 'warning' | 'info' | string;

export interface MlcDiagnostic {
  line: number;
  col: number;
  severity: Severity;
  code: string;
  message: string;
}

export interface MlcFieldSpan {
  line: number;
  col: number;
  endLine: number;
  endCol: number;
  name: string;
  named: boolean;
  kind: 'check' | 'text' | 'multiline' | string;
  pattern: string;
  prefill: boolean;
}

export interface MlcHintSpan {
  line: number;
  col: number;
  endLine: number;
  endCol: number;
  kind: 'tooltip' | 'fill-hint' | string;
}

export interface MlcPrefillField {
  name: string;
  marker: string;
  label: string;
}

export interface MlcLists {
  eventKinds: string[] | null;
  prefill: MlcPrefillField[] | null;
  frontMatterKeys: string[] | null;
}

/** mlcext.Analysis. Positions: 1-based, columns in code points, ends exclusive. */
export interface MlcAnalysis {
  detected: boolean;
  diagnostics: MlcDiagnostic[] | null;
  fields: MlcFieldSpan[] | null;
  hints: MlcHintSpan[] | null;
  lists: MlcLists;
}

/** Where the analysis comes from: the host calls mlcext.Analyze. */
export interface MlcSource {
  analyze(markdown: string): Promise<MlcAnalysis | null>;
  /**
   * Quick fixes, also from the host (mlcext): the label for a diagnostic
   * code, or null when there is none, and the fixed document.
   */
  fixLabel?(code: string): string | null;
  /** diagnostic: the finding (line, col) the fix is for. */
  fix?(code: string, markdown: string, diagnostic: MlcDiagnostic): Promise<string | null>;
}

export interface MlcOptions {
  /** Delay after the last change before analysing (ms). Default 300. */
  debounce?: number;
  /** Called with every fresh analysis (status bar, menus …). */
  onAnalysis?: (analysis: MlcAnalysis) => void;
}

// ---------- Positions ----------

/** Document offset of a 1-based line and a column counted in code points. */
export function toPos(doc: Text, line: number, col: number): number {
  if (line > doc.lines) return doc.length;
  const l = doc.line(Math.max(1, line));
  let offset = 0;
  let points = 1;
  for (const ch of l.text) {
    if (points >= col) break;
    offset += ch.length;
    points++;
  }
  return l.from + offset;
}

// ---------- State ----------

const setAnalysis = StateEffect.define<MlcAnalysis>();

interface MlcState {
  analysis: MlcAnalysis | null;
  decorations: DecorationSet;
}

const KIND_LABEL: Record<string, string> = {
  check: 'Ankreuzfeld',
  text: 'Textfeld',
  multiline: 'mehrzeiliges Textfeld',
};

function buildDecorations(doc: Text, a: MlcAnalysis): DecorationSet {
  if (!a.detected) return Decoration.none;
  const ranges = [];
  for (const f of a.fields ?? []) {
    const from = toPos(doc, f.line, f.col);
    const to = toPos(doc, f.endLine, f.endCol);
    if (to <= from) continue;
    const title =
      `MLC-Erweiterung: ${KIND_LABEL[f.kind] ?? f.kind} „${f.name}“` +
      (f.pattern ? ` (${f.pattern})` : '') +
      (f.prefill ? ' – wird aus einem Ereignis vorausgefüllt' : '');
    ranges.push(
      Decoration.mark({
        class: `cm-mlc-field cm-mlc-field-${f.kind}${f.prefill ? ' cm-mlc-field-prefill' : ''}`,
        attributes: { title },
      }).range(from, to),
    );
  }
  for (const h of a.hints ?? []) {
    const from = toPos(doc, h.line, h.col);
    const to = toPos(doc, h.endLine, h.endCol);
    if (to <= from) continue;
    const title =
      h.kind === 'tooltip'
        ? 'MLC-Erweiterung: Tooltip am Feld (am Bildschirm sichtbar, nicht gedruckt)'
        : 'MLC-Erweiterung: Ausfülltipp (nur in der App, nicht im PDF)';
    ranges.push(Decoration.mark({ class: `cm-mlc-hint cm-mlc-${h.kind}`, attributes: { title } }).range(from, to));
  }
  return Decoration.set(ranges, true);
}

const mlcField = StateField.define<MlcState>({
  create: () => ({ analysis: null, decorations: Decoration.none }),
  update(value, tr) {
    let { analysis, decorations } = value;
    if (tr.docChanged) decorations = decorations.map(tr.changes);
    for (const e of tr.effects) {
      if (e.is(setAnalysis)) {
        analysis = e.value;
        decorations = buildDecorations(tr.state.doc, e.value);
      }
    }
    return { analysis, decorations };
  },
  provide: (f) => EditorView.decorations.from(f, (v) => v.decorations),
});

/** The latest analysis, e.g. for a host menu. */
export function currentAnalysis(state: EditorState): MlcAnalysis | null {
  return state.field(mlcField, false)?.analysis ?? null;
}

function toCMDiagnostics(doc: Text, a: MlcAnalysis, source?: MlcSource): CMDiagnostic[] {
  if (!a.detected) return [];
  return (a.diagnostics ?? []).map((d) => {
    const label = source?.fix && d.code ? source.fixLabel?.(d.code) : null;
    const actions = label
      ? [
          {
            name: label,
            apply(view: EditorView) {
              const text = view.state.doc.toString();
              void source!.fix!(d.code, text, d).then((fixed) => {
                if (fixed == null || view.state.doc.toString() !== text) return;
                view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: fixed } });
              });
            },
          },
        ]
      : undefined;
    const from = toPos(doc, d.line, d.col);
    const field = (a.fields ?? []).find((f) => f.line === d.line && f.col === d.col);
    const to = field ? toPos(doc, field.endLine, field.endCol) : doc.lineAt(from).to;
    return {
      from,
      to: Math.max(to, from),
      severity: d.severity === 'error' ? 'error' : d.severity === 'warning' ? 'warning' : 'info',
      source: 'MLC-Erweiterung',
      message: d.message,
      actions,
    };
  });
}

function analyzer(source: MlcSource, opt: MlcOptions) {
  return ViewPlugin.fromClass(
    class {
      timer: ReturnType<typeof setTimeout> | undefined;
      constructor(readonly view: EditorView) {
        this.schedule(0);
      }
      update(u: ViewUpdate) {
        if (u.docChanged) this.schedule(opt.debounce ?? 300);
      }
      schedule(delay: number) {
        clearTimeout(this.timer);
        this.timer = setTimeout(() => void this.run(), delay);
      }
      async run() {
        const text = this.view.state.doc.toString();
        const a = await source.analyze(text);
        // A stale answer (the text changed meanwhile) is dropped; the
        // pending run brings the fresh one.
        if (!a || this.view.state.doc.toString() !== text) return;
        this.view.dispatch(setDiagnostics(this.view.state, toCMDiagnostics(this.view.state.doc, a, source)));
        this.view.dispatch({ effects: setAnalysis.of(a) });
        opt.onAnalysis?.(a);
      }
      destroy() {
        clearTimeout(this.timer);
      }
    },
  );
}

// ---------- Snippets ----------

export interface MlcSnippet {
  label: string;
  detail: string;
  /** CodeMirror snippet template: ${name} fields, ${} cursor. */
  template: string;
}

const LINE = '______________________________';
const AREA = '_'.repeat(70);

/** Building blocks of the extension, for completion and host menus. */
export function mlcSnippets(lists?: MlcLists | null): MlcSnippet[] {
  const base: MlcSnippet[] = [
    { label: 'Textfeld', detail: '[@#name ____]', template: `[@#\${name} ${LINE}]` },
    { label: 'Ankreuzfeld', detail: '[@#name]', template: '[@#${name}]' },
    { label: 'Datum', detail: 'TT.MM.JJJJ', template: '[@#${datum} TT.MM.JJJJ]' },
    { label: 'Uhrzeit', detail: 'HH:MM', template: '[@#${uhrzeit} HH:MM]' },
    {
      label: 'Mehrzeiliges Feld',
      detail: 'drei Zeilen',
      template: `[@#\${name} ${AREA}\n${AREA}\n${AREA}]`,
    },
    { label: 'Feld bis Seitenende', detail: 'wächst bis zum Seitenende', template: `[@#\${name} ${AREA}\n…]` },
    { label: 'Tooltip am Feld', detail: '<!--* … -->', template: '<!--* ${Hinweis} -->' },
    { label: 'Ausfülltipp', detail: '<!--! … --> (nur in der App)', template: '<!--! ${So arbeiten Sie mit diesem Blatt.} -->' },
    {
      label: 'Dokumentkopf',
      detail: 'Titel, Dokumenten-ID, Version',
      template: '# ${Titel}\n\n**Dokumenten-ID:** ${SOP-XX-001} | **Version:** ${1.0}\n\n---\n\n${}',
    },
  ];
  const prefill = (lists?.prefill ?? []).map((p) => ({
    label: `Vorausgefüllt: ${p.name}`,
    detail: p.label,
    template: p.marker.replace(/[{}]/g, (c) => '\\' + c), // literal braces in snippet syntax
  }));
  return [...base, ...prefill];
}

/** Insert a snippet at the selection (for host menus and buttons). */
export function insertMlcSnippet(view: EditorView, s: MlcSnippet): void {
  const { from, to } = view.state.selection.main;
  snippet(s.template)(view, null, from, to);
  view.focus();
}

// ---------- Completion ----------

function inFrontMatter(state: EditorState, pos: number): boolean {
  const doc = state.doc;
  if (doc.lines < 2 || doc.line(1).text.trim() !== '---') return false;
  const here = doc.lineAt(pos).number;
  for (let n = 2; n <= doc.lines; n++) {
    const t = doc.line(n).text.trim();
    if (t === '---' || t === '...') return here > 1 && here < n;
  }
  return false;
}

function mlcCompletions(context: CompletionContext): CompletionResult | null {
  const lists = currentAnalysis(context.state)?.lists ?? null;

  if (inFrontMatter(context.state, context.pos)) {
    const line = context.state.doc.lineAt(context.pos);
    const before = line.text.slice(0, context.pos - line.from);
    if (/^event_kinds\s*:/.test(before)) {
      const word = context.matchBefore(/[A-Z_]*/);
      if (!word) return null;
      return {
        from: word.from,
        options: (lists?.eventKinds ?? []).map((k) => ({ label: k, type: 'enum', detail: 'Ereignisart' })),
        validFor: /^[A-Z_]*$/,
      };
    }
    const key = context.matchBefore(/^[a-z_]*/);
    if (key && (key.from < key.to || context.explicit)) {
      return {
        from: key.from,
        options: (lists?.frontMatterKeys ?? []).map((k) => ({
          label: k,
          type: 'property',
          apply: k === 'event_kinds' ? 'event_kinds: [' : `${k}: `,
        })),
      };
    }
    return null;
  }

  const variable = context.matchBefore(/\{\{\s*\w*/);
  if (variable) {
    return {
      from: variable.from,
      options: [
        { label: '{{doc_id}}', detail: 'Kennung aus dem Front Matter' },
        { label: '{{version}}', detail: 'Version aus dem Front Matter' },
        { label: '{{title}}', detail: 'Titel aus dem Front Matter' },
      ].map((o) => ({ ...o, type: 'variable' })),
      filter: false,
    };
  }

  const field = context.matchBefore(/\[@[#\w-]*/);
  const comment = context.matchBefore(/<!--?[-*!]?/);
  const match = field ?? comment;
  if (!match) return null;
  const options: Completion[] = mlcSnippets(lists)
    .filter((s) => (field ? s.template.startsWith('[@') : s.template.startsWith('<!--')))
    .map((s) => ({ label: s.label, detail: s.detail, type: 'text', apply: snippet(s.template) }));
  return { from: match.from, options, filter: false };
}

// ---------- Theme ----------

const mlcTheme = EditorView.baseTheme({
  '.cm-mlc-field': {
    backgroundColor: 'rgba(201, 76, 53, 0.10)',
    borderBottom: '1px solid rgba(201, 76, 53, 0.55)',
    borderRadius: '2px',
  },
  '.cm-mlc-field-prefill': { borderBottomStyle: 'dashed' },
  '.cm-mlc-hint': { color: '#8a6d00', fontStyle: 'italic' },
  '&dark .cm-mlc-field': {
    backgroundColor: 'rgba(240, 128, 104, 0.14)',
    borderBottomColor: 'rgba(240, 128, 104, 0.6)',
  },
  '&dark .cm-mlc-hint': { color: '#d6b656' },
});

// ---------- Entry point ----------

/**
 * Editor support for the MLC Markdown Extension. Plain Markdown is left
 * alone: without a detected MLC document there is no highlighting and no
 * diagnostic, only the "[@" completion.
 */
export function mlcExtension(source: MlcSource, opt: MlcOptions = {}): Extension {
  return [
    mlcField,
    analyzer(source, opt),
    lintGutter(),
    EditorState.languageData.of(() => [{ autocomplete: mlcCompletions }]),
    mlcTheme,
  ];
}
