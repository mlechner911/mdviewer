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
import { Decoration, EditorView, ViewPlugin, hoverTooltip, type DecorationSet, type ViewUpdate } from '@codemirror/view';
import { lintGutter, setDiagnostics, type Diagnostic as CMDiagnostic } from '@codemirror/lint';
import { snippet, type Completion, type CompletionContext, type CompletionResult } from '@codemirror/autocomplete';
import { fieldHelp, propertyInsertion, propertyOptions, type MlcSchema } from './schemaHelp';

// ---------- The analysis (JSON of mlcext.Analysis) ----------

export type Severity = 'error' | 'warning' | 'info' | string;

export interface MlcDiagnostic {
  line: number;
  col: number;
  severity: Severity;
  code: string;
  message: string;
  /** Who reported it (tooltip); default the MLC extension. */
  source?: string;
}

export interface MlcFieldSpan {
  line: number;
  col: number;
  endLine: number;
  endCol: number;
  name: string;
  named: boolean;
  kind: 'check' | 'text' | 'multiline' | 'signature' | string;
  pattern: string;
  prefill: boolean;
  /** Role of a signature field, who signs ("QMB"); absent for other kinds. */
  role?: string;
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
  /** The schema the document uses, if it is valid and known. */
  schema?: MlcSchema | null;
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
  /** Language of schema descriptions, read each time they are shown. */
  locale?: () => string;
  /**
   * Findings from elsewhere (e.g. a folder check after saving), shown with
   * the extension's own; read on every analysis. Not passed to onAnalysis.
   */
  extraDiagnostics?: (markdown: string) => MlcDiagnostic[];
  /**
   * Completions of the host (placeholders, event kinds, zones, field names
   * it knows), next to the extension's own. Called with where the cursor
   * is; may answer asynchronously. See MlcCompletionContext.
   */
  completions?: (context: MlcCompletionContext) => MlcCompletionItem[] | null | Promise<MlcCompletionItem[] | null>;
}

/**
 * Where completion was asked for. The host's items replace the token that
 * is being typed:
 * - `placeholder`: `{{…` in the text; token is `{{` plus what follows, so
 *   a label is written in full ("{{sensor.name}}").
 * - `field-name`: after `[@#`; token is the name typed so far.
 * - `event-kind`: in the front matter list `event_kinds: [`.
 * - `front-matter-key`: a key at the start of a front matter line; an item
 *   without `apply` inserts "label: ".
 * - `front-matter-value`: the value of `key: ` in the front matter (`key`
 *   names it), e.g. a zone.
 */
export type MlcCompletionKind = 'placeholder' | 'field-name' | 'event-kind' | 'front-matter-key' | 'front-matter-value';

export interface MlcCompletionContext {
  kind: MlcCompletionKind;
  /** The token typed so far (for a placeholder without the braces). */
  prefix: string;
  /** front-matter-value: the key. */
  key?: string;
  /** Cursor, 1-based line and column in code points (like diagnostics). */
  line: number;
  col: number;
  /** Asked for explicitly (Ctrl+Space) rather than while typing. */
  explicit: boolean;
  /** The latest analysis, if any. */
  analysis: MlcAnalysis | null;
}

export interface MlcCompletionItem {
  label: string;
  detail?: string;
  /** Longer description shown next to the list. */
  info?: string;
  /** Inserted text; default the label (front matter keys: "label: "). */
  apply?: string;
  /** CodeMirror completion type for the icon, default by kind. */
  type?: string;
}

/** Dispatch to analyse again now, e.g. when extraDiagnostics changed. */
export const refreshAnalysis = StateEffect.define<null>();

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
  signature: 'Unterschriftsfeld',
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
      (f.role ? ` – ${f.role}` : '') +
      (f.prefill ? ' – wird aus einem Ereignis vorausgefüllt' : '');
    ranges.push(
      Decoration.mark({
        class: `cm-mlc-field cm-mlc-field-${f.kind}${f.prefill ? ' cm-mlc-field-prefill' : ''}`,
        // A schema's description comes as a tooltip (schemaHover) instead.
        attributes: f.named && fieldHelp(a.schema, f.name, '') ? {} : { title },
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

/** What the checker says about a field that is still being typed. */
const UNFINISHED_FIELD = ['field-unclosed', 'field-name'];

/**
 * A field the cursor is still in: "[@#" is no error yet, the author is typing.
 * It counts again as soon as the cursor leaves it.
 */
export function isBeingTyped(doc: Text, d: MlcDiagnostic, cursor: number): boolean {
  if (!UNFINISHED_FIELD.includes(d.code)) return false;
  const from = toPos(doc, d.line, d.col);
  return cursor > from && cursor <= doc.lineAt(from).to;
}

function toCMDiagnostics(doc: Text, a: MlcAnalysis, source: MlcSource | undefined, cursor: number): CMDiagnostic[] {
  // Outside the extension the host may still send notes (HTML in the PDF).
  return (a.diagnostics ?? []).filter((d) => !isBeingTyped(doc, d, cursor)).map((d) => {
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
      source: d.source ?? 'MLC-Erweiterung',
      message: d.message,
      actions,
    };
  });
}

// ---------- Findings pushed by the host ----------

interface HostDiagnostic {
  d: MlcDiagnostic;
  from: number;
  to: number;
}

const setHostDiagnostics = StateEffect.define<MlcDiagnostic[]>();

/** The host's findings, anchored to the text and moved with every edit. */
const hostDiagnosticsField = StateField.define<HostDiagnostic[]>({
  create: () => [],
  update(value, tr) {
    for (const e of tr.effects) {
      if (e.is(setHostDiagnostics)) {
        const doc = tr.state.doc;
        return e.value.map((d) => {
          const from = toPos(doc, d.line, d.col);
          return { d, from, to: Math.max(from, doc.lineAt(from).to) };
        });
      }
    }
    if (!tr.docChanged || !value.length) return value;
    return value.map((h) => {
      const from = tr.changes.mapPos(h.from, 1);
      return { ...h, from, to: Math.max(from, tr.changes.mapPos(h.to, -1)) };
    });
  },
});

/**
 * Show findings of the host, e.g. its server's rules, next to those of the
 * extension. Same form as mlcext diagnostics (1-based line, column in code
 * points); `source` names the checker in the tooltip. They move with the
 * text while the author types and replace the previous set on every call;
 * an empty list removes them. They show without an MlcSource analysis too.
 */
export function setMlcDiagnostics(view: EditorView, diagnostics: MlcDiagnostic[]): void {
  view.dispatch({ effects: setHostDiagnostics.of(diagnostics) });
}

function hostToCM(state: EditorState): CMDiagnostic[] {
  const len = state.doc.length;
  return (state.field(hostDiagnosticsField, false) ?? []).map(({ d, from, to }) => ({
    from: Math.min(from, len),
    to: Math.min(Math.max(to, from), len),
    severity: d.severity === 'error' ? 'error' : d.severity === 'warning' ? 'warning' : 'info',
    source: d.source ?? 'Host',
    message: d.message,
  }));
}

function analyzer(source: MlcSource, opt: MlcOptions) {
  return ViewPlugin.fromClass(
    class {
      timer: ReturnType<typeof setTimeout> | undefined;
      /** A finding is held back while the cursor is in its field. */
      heldBack = false;
      /** The extension's findings of the latest analysis, with the extra ones. */
      own: CMDiagnostic[] = [];
      constructor(readonly view: EditorView) {
        this.schedule(0);
      }
      update(u: ViewUpdate) {
        if (u.transactions.some((tr) => tr.effects.some((e) => e.is(setHostDiagnostics)))) {
          // Shown at once, without asking the analysis again.
          queueMicrotask(() => this.show());
        }
        if (u.transactions.some((tr) => tr.effects.some((e) => e.is(refreshAnalysis)))) this.schedule(0);
        else if (u.docChanged || (u.selectionSet && this.heldBack)) this.schedule(opt.debounce ?? 300);
        if (u.docChanged) {
          this.own = this.own.map((d) => {
            const from = u.changes.mapPos(d.from, 1);
            return { ...d, from, to: Math.max(from, u.changes.mapPos(d.to, -1)) };
          });
        }
      }
      show() {
        const own = this.own.filter((d) => d.to <= this.view.state.doc.length);
        this.view.dispatch(setDiagnostics(this.view.state, [...own, ...hostToCM(this.view.state)]));
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
        const { doc, selection } = this.view.state;
        const cursor = selection.main.head;
        this.heldBack = !!a.detected && (a.diagnostics ?? []).some((d) => isBeingTyped(doc, d, cursor));
        const extra = opt.extraDiagnostics?.(text) ?? [];
        const shown = extra.length ? { ...a, diagnostics: [...(a.diagnostics ?? []), ...extra] } : a;
        this.own = toCMDiagnostics(doc, shown, source, cursor);
        this.show();
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

/**
 * Building blocks of the extension, for completion and host menus. With a
 * schema, new fields take its names, so the old prefill names are left out.
 */
export function mlcSnippets(lists?: MlcLists | null, schema?: MlcSchema | null): MlcSnippet[] {
  const base: MlcSnippet[] = [
    { label: 'Textfeld', detail: '[@#name ____]', template: `[@#\${name} ${LINE}]` },
    { label: 'Ankreuzfeld', detail: '[@#name]', template: '[@#${name}]' },
    { label: 'Datum', detail: 'TT.MM.JJJJ', template: '[@#${datum} TT.MM.JJJJ]' },
    { label: 'Uhrzeit', detail: 'HH:MM', template: '[@#${uhrzeit} HH:MM]' },
    { label: 'Unterschrift', detail: 'wer unterschreibt', template: '[@#${unterschrift} Unterschrift: ${Rolle}]' },
    {
      label: 'Mehrzeiliges Feld',
      detail: 'drei Zeilen',
      template: `[@#\${name} ${AREA}\n${AREA}\n${AREA}]`,
    },
    { label: 'Feld bis Seitenende', detail: 'wächst bis zum Seitenende', template: `[@#\${name} ${AREA}\n…]` },
    { label: 'Tooltip am Feld', detail: '<!--* … -->', template: '<!--* ${Hinweis} -->' },
    { label: 'Ausfülltipp', detail: '<!--! … --> (nur in der App)', template: '<!--! ${So arbeiten Sie mit diesem Blatt.} -->' },
  ];
  if (schema) return base;
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

/** Completion of a schema's properties after "[@#". */
function schemaCompletions(context: CompletionContext, a: MlcAnalysis | null, locale: string): CompletionResult | null {
  const typed = context.matchBefore(/\[@#[a-z0-9_.]*/);
  if (!typed || !a?.schema) return null;
  const used = (a.fields ?? []).filter((f) => f.named).map((f) => f.name);
  const options = propertyOptions(a.schema, used, typed.text.slice(3), locale).map(
    (o): Completion => ({
      label: o.name,
      detail: o.detail,
      type: 'property',
      apply(view, _completion, from, to) {
        const insert = propertyInsertion(o.name, o.type, view.state.doc.lineAt(to).text.slice(to - view.state.doc.lineAt(to).from));
        view.dispatch({ changes: { from, to, insert }, selection: { anchor: from + insert.length } });
      },
    }),
  );
  return { from: typed.from + 3, options, validFor: /^[a-z0-9_.]*$/ };
}

/** Where the host may add items: the kind and the token being typed. */
interface HostSlot {
  kind: MlcCompletionKind;
  from: number;
  prefix: string;
  key?: string;
}

interface Asked {
  result: CompletionResult | null;
  slot: HostSlot | null;
}

function mlcCompletions(context: CompletionContext, locale: string): Asked {
  const analysis = currentAnalysis(context.state);
  const lists = analysis?.lists ?? null;

  if (inFrontMatter(context.state, context.pos)) {
    const line = context.state.doc.lineAt(context.pos);
    const before = line.text.slice(0, context.pos - line.from);
    if (/^event_kinds\s*:/.test(before)) {
      const word = context.matchBefore(/[A-Z_]*/);
      if (!word) return { result: null, slot: null };
      return {
        result: {
          from: word.from,
          options: (lists?.eventKinds ?? []).map((k) => ({ label: k, type: 'enum', detail: 'Ereignisart' })),
          validFor: /^[A-Z_]*$/,
        },
        slot: { kind: 'event-kind', from: word.from, prefix: word.text },
      };
    }
    const value = /^([A-Za-z_][\w-]*)\s*:\s*/.exec(before);
    if (value) {
      const word = context.matchBefore(/[^\s,[\]"']*/);
      if (!word || (word.from === word.to && !context.explicit)) return { result: null, slot: null };
      return { result: null, slot: { kind: 'front-matter-value', from: word.from, prefix: word.text, key: value[1] } };
    }
    const key = context.matchBefore(/^[a-z_]*/);
    if (key && (key.from < key.to || context.explicit)) {
      return {
        result: {
          from: key.from,
          options: (lists?.frontMatterKeys ?? []).map((k) => ({
            label: k,
            type: 'property',
            apply: k === 'event_kinds' ? 'event_kinds: [' : `${k}: `,
          })),
        },
        slot: { kind: 'front-matter-key', from: key.from, prefix: key.text },
      };
    }
    return { result: null, slot: null };
  }

  const variable = context.matchBefore(/\{\{\s*[\w.]*/);
  if (variable) {
    return {
      result: {
        from: variable.from,
        options: [
          { label: '{{doc_id}}', detail: 'Kennung aus dem Front Matter' },
          { label: '{{version}}', detail: 'Version aus dem Front Matter' },
          { label: '{{title}}', detail: 'Titel aus dem Front Matter' },
        ].map((o) => ({ ...o, type: 'variable' })),
        filter: false,
      },
      slot: { kind: 'placeholder', from: variable.from, prefix: variable.text.slice(2).trim() },
    };
  }

  const named = context.matchBefore(/\[@#[a-z0-9_.]*/);
  const slot: HostSlot | null = named ? { kind: 'field-name', from: named.from + 3, prefix: named.text.slice(3) } : null;

  const property = schemaCompletions(context, analysis, locale);
  if (property) return { result: property, slot };

  const field = context.matchBefore(/\[@[#\w-]*/);
  const comment = context.matchBefore(/<!--?[-*!]?/);
  const match = field ?? comment;
  if (!match) return { result: null, slot };
  const options: Completion[] = mlcSnippets(lists, analysis?.schema)
    .filter((s) => (field ? s.template.startsWith('[@') : s.template.startsWith('<!--')))
    .map((s) => ({
      label: s.label,
      detail: s.detail,
      type: 'text',
      apply(view, completion, from, to) {
        // The editor has closed the bracket itself; the snippet brings its own.
        const closed = field && view.state.doc.sliceString(to, to + 1) === ']';
        snippet(s.template)(view, completion, from, closed ? to + 1 : to);
      },
    }));
  return { result: { from: match.from, options, filter: false }, slot };
}

const HOST_TYPE: Record<MlcCompletionKind, string> = {
  placeholder: 'variable',
  'field-name': 'property',
  'event-kind': 'enum',
  'front-matter-key': 'property',
  'front-matter-value': 'constant',
};

/** A host item as a CodeMirror completion; it replaces exactly the slot's token. */
function hostCompletion(slot: HostSlot, item: MlcCompletionItem): Completion {
  const insert = item.apply ?? (slot.kind === 'front-matter-key' ? `${item.label}: ` : item.label);
  return {
    label: item.label,
    detail: item.detail,
    info: item.info,
    type: item.type ?? HOST_TYPE[slot.kind],
    apply(view, _completion, _from, to) {
      view.dispatch({ changes: { from: slot.from, to, insert }, selection: { anchor: slot.from + insert.length } });
    },
  };
}

/** The extension's completions, with the host's items added. */
async function withHost(context: CompletionContext, asked: Asked, opt: MlcOptions): Promise<CompletionResult | null> {
  const { result, slot } = asked;
  if (!slot || !opt.completions) return result;
  const line = context.state.doc.lineAt(context.pos);
  let items: MlcCompletionItem[] | null = null;
  try {
    items = await opt.completions({
      kind: slot.kind,
      prefix: slot.prefix,
      key: slot.key,
      line: line.number,
      col: [...line.text.slice(0, context.pos - line.from)].length + 1,
      explicit: context.explicit,
      analysis: currentAnalysis(context.state),
    });
  } catch {
    // A failing host must not take the extension's own list with it.
  }
  if (!items?.length) return result;
  const known = new Set(result?.options.map((o) => o.label));
  const extra = items.filter((i) => !known.has(i.label)).map((i) => hostCompletion(slot, i));
  if (!result) return { from: slot.from, options: extra };
  return { ...result, options: [...result.options, ...extra] };
}

// ---------- Schema descriptions ----------

function schemaHover(locale: () => string) {
  return hoverTooltip((view, pos) => {
    const a = currentAnalysis(view.state);
    if (!a?.schema) return null;
    for (const f of a.fields ?? []) {
      const from = toPos(view.state.doc, f.line, f.col);
      const to = toPos(view.state.doc, f.endLine, f.endCol);
      if (pos < from || pos > to || !f.named) continue;
      const help = fieldHelp(a.schema, f.name, locale());
      if (!help) return null;
      return {
        pos: from,
        end: to,
        above: true,
        create() {
          const dom = document.createElement('div');
          dom.className = 'cm-mlc-schema-help';
          const name = dom.appendChild(document.createElement('strong'));
          name.textContent = f.name;
          for (const line of help) dom.appendChild(document.createElement('div')).textContent = line;
          return { dom };
        },
      };
    }
    return null;
  });
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
  '.cm-mlc-schema-help': { padding: '4px 8px', maxWidth: '28rem' },
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
  const locale = opt.locale ?? (() => 'en');
  // One function for the life of the editor: CodeMirror compares sources by
  // identity, and a new one per call restarts the query for ever (no list).
  const complete = (context: CompletionContext) => withHost(context, mlcCompletions(context, locale()), opt);
  return [
    schemaHover(locale),
    mlcField,
    hostDiagnosticsField,
    analyzer(source, opt),
    lintGutter(),
    EditorState.languageData.of(() => [{ autocomplete: complete }]),
    mlcTheme,
  ];
}

export { exampleValues, fieldHelp, localizedText, propertyOptions, type MlcSchema, type MlcSchemaField } from './schemaHelp';

// Field dialog support: box at the field under the cursor, double click.
export { fieldHint, fieldAt, fieldAtCursor, type FieldHintOptions } from './fieldHint';
