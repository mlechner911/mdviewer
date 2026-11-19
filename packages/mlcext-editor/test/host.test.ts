// Findings and completions handed in by the host (P-MARKS-20261117-01,
// phase 1): the editor shows them next to the extension's own.
import { describe, expect, it } from 'vitest';
import { EditorSelection, EditorState } from '@codemirror/state';
import { EditorView } from '@codemirror/view';
import { diagnosticCount, forEachDiagnostic, type Diagnostic } from '@codemirror/lint';
import { CompletionContext, type CompletionResult, type CompletionSource } from '@codemirror/autocomplete';
import {
  mlcExtension,
  setMlcDiagnostics,
  type MlcAnalysis,
  type MlcCompletionContext,
  type MlcCompletionItem,
  type MlcLists,
  type MlcOptions,
  type MlcSource,
} from '../src/index';

const wait = () => new Promise((r) => setTimeout(r, 20));
const lists: MlcLists = { eventKinds: ['DEVIATION', 'ALARM'], prefill: [], frontMatterKeys: ['title', 'event_kinds', 'zone'] };
const analysis: MlcAnalysis = {
  detected: true,
  diagnostics: [{ line: 1, col: 1, severity: 'warning', code: 'mlc-own', message: 'own' }],
  fields: [],
  hints: [],
  lists,
};

function editor(doc: string, source: MlcSource, opt: MlcOptions = {}) {
  return new EditorView({ parent: document.body, state: EditorState.create({ doc, extensions: mlcExtension(source, opt) }) });
}

function shown(view: EditorView): Diagnostic[] {
  const all: Diagnostic[] = [];
  forEachDiagnostic(view.state, (d) => all.push(d));
  return all;
}

describe('setMlcDiagnostics', () => {
  it('shows the host findings next to the analysis, without analysing again', async () => {
    let analyses = 0;
    const view = editor('Zeile eins\nZeile zwei', { analyze: async () => (analyses++, analysis) });
    await wait();
    setMlcDiagnostics(view, [{ line: 2, col: 7, severity: 'error', code: 'srv', message: 'Zone fehlt', source: 'Werkstatt' }]);
    await wait();
    expect(analyses).toBe(1);
    const got = shown(view);
    expect(got).toHaveLength(2);
    const host = got.find((d) => d.source === 'Werkstatt')!;
    expect(host.severity).toBe('error');
    expect(view.state.doc.sliceString(host.from, host.to)).toBe('zwei');
    view.destroy();
  });

  it('moves them with the text and keeps them through the next analysis', async () => {
    const view = editor('Zeile eins\nZeile zwei', { analyze: async () => analysis }, { debounce: 0 });
    await wait();
    setMlcDiagnostics(view, [{ line: 2, col: 7, severity: 'warning', code: 'srv', message: 'm' }]);
    await wait();
    view.dispatch({ changes: { from: 0, insert: 'Neu\n' } });
    await wait();
    const host = shown(view).find((d) => d.message === 'm')!;
    expect(view.state.doc.sliceString(host.from, host.to)).toBe('zwei');
    expect(diagnosticCount(view.state)).toBe(2);
    view.destroy();
  });

  it('replaces the previous set, an empty list removes them', async () => {
    const view = editor('a\nb', { analyze: async () => analysis });
    await wait();
    setMlcDiagnostics(view, [
      { line: 1, col: 1, severity: 'info', code: 'x', message: 'x' },
      { line: 2, col: 1, severity: 'info', code: 'y', message: 'y' },
    ]);
    await wait();
    // shown(), not diagnosticCount: CodeMirror counts findings on the same
    // range (here own and x on "a") once.
    expect(shown(view).map((d) => d.message).sort()).toEqual(['own', 'x', 'y']);
    setMlcDiagnostics(view, []);
    await wait();
    expect(shown(view).map((d) => d.message)).toEqual(['own']);
    view.destroy();
  });

  it('shows them without any analysis', async () => {
    const view = editor('nur Text', { analyze: async () => null });
    await wait();
    setMlcDiagnostics(view, [{ line: 1, col: 5, severity: 'warning', code: 'srv', message: 'm' }]);
    await wait();
    expect(diagnosticCount(view.state)).toBe(1);
    view.destroy();
  });
});

/** Ask the editor's completion source at the end of doc (or at |). */
async function complete(doc: string, opt: MlcOptions, explicit = false): Promise<{ result: CompletionResult | null; state: EditorState }> {
  const pos = doc.includes('|') ? doc.indexOf('|') : doc.length;
  const text = doc.replace('|', '');
  const view = editor(text, { analyze: async () => ({ ...analysis, diagnostics: [] }) }, opt);
  await wait();
  view.dispatch({ selection: EditorSelection.cursor(pos) });
  const source = view.state.languageDataAt<CompletionSource>('autocomplete', pos)[0];
  const result = (await source(new CompletionContext(view.state, pos, explicit))) as CompletionResult | null;
  const state = view.state;
  view.destroy();
  return { result, state };
}

function host(items: MlcCompletionItem[]) {
  const asked: MlcCompletionContext[] = [];
  return { asked, opt: { completions: (c: MlcCompletionContext) => (asked.push(c), items) } satisfies MlcOptions };
}

function applyOption(doc: string, result: CompletionResult, label: string): string {
  const pos = doc.includes('|') ? doc.indexOf('|') : doc.length;
  const view = new EditorView({ parent: document.body, state: EditorState.create({ doc: doc.replace('|', ''), selection: { anchor: pos } }) });
  const option = result.options.find((o) => o.label === label)!;
  if (typeof option.apply === 'function') option.apply(view, option, result.from, pos);
  else view.dispatch({ changes: { from: result.from, to: pos, insert: option.apply ?? option.label } });
  const out = view.state.doc.toString();
  view.destroy();
  return out;
}

describe('completions from the host', () => {
  it('placeholders: added to the built-in ones, with prefix and position', async () => {
    const { asked, opt } = host([{ label: '{{sensor.name}}', detail: 'Steckbrief' }]);
    const doc = 'Sensor {{sen';
    const { result } = await complete(doc, opt);
    expect(asked[0]).toMatchObject({ kind: 'placeholder', prefix: 'sen', line: 1, col: 13, explicit: false });
    const labels = result!.options.map((o) => o.label);
    expect(labels).toContain('{{doc_id}}');
    expect(labels).toContain('{{sensor.name}}');
    expect(applyOption(doc, result!, '{{sensor.name}}')).toBe('Sensor {{sensor.name}}');
  });

  it('field names after [@#, also without a schema', async () => {
    const { asked, opt } = host([{ label: 'zone_temp', detail: 'Zone' }]);
    const doc = 'Wert: [@#zo|]';
    const { result } = await complete(doc, opt);
    expect(asked[0]).toMatchObject({ kind: 'field-name', prefix: 'zo' });
    expect(applyOption(doc, result!, 'zone_temp')).toBe('Wert: [@#zone_temp]');
  });

  it('event kinds of the house next to the shared list, duplicates once', async () => {
    const { asked, opt } = host([{ label: 'DOOR_OPEN' }, { label: 'ALARM' }]);
    const { result } = await complete('---\nevent_kinds: [DO|\n---\n# T', opt);
    expect(asked[0].kind).toBe('event-kind');
    expect(result!.options.map((o) => o.label)).toEqual(['DEVIATION', 'ALARM', 'DOOR_OPEN']);
  });

  it('front matter keys insert "key: "', async () => {
    const { asked, opt } = host([{ label: 'sensor_profile' }]);
    const doc = '---\nsen|\n---\n# T';
    const { result } = await complete(doc, opt);
    expect(asked[0]).toMatchObject({ kind: 'front-matter-key', prefix: 'sen' });
    expect(applyOption(doc, result!, 'sensor_profile')).toBe('---\nsensor_profile: \n---\n# T');
  });

  it('front matter values name their key (zones)', async () => {
    const { asked, opt } = host([{ label: 'Kühlraum 2', apply: 'kuehlraum-2' }]);
    const doc = '---\nzone: kü|\n---\n# T';
    const { result } = await complete(doc, opt);
    expect(asked[0]).toMatchObject({ kind: 'front-matter-value', key: 'zone', prefix: 'kü' });
    expect(applyOption(doc, result!, 'Kühlraum 2')).toBe('---\nzone: kuehlraum-2\n---\n# T');
  });

  it('an empty value asks only on Ctrl+Space', async () => {
    const { asked, opt } = host([{ label: 'A' }]);
    await complete('---\nzone: |\n---\n# T', opt);
    expect(asked).toHaveLength(0);
    await complete('---\nzone: |\n---\n# T', opt, true);
    expect(asked).toHaveLength(1);
  });

  it('answers asynchronously', async () => {
    const opt: MlcOptions = { completions: async () => [{ label: '{{zone}}' }] };
    const { result } = await complete('{{', opt);
    expect(result!.options.map((o) => o.label)).toContain('{{zone}}');
  });

  it('a failing host keeps the built-in list', async () => {
    const opt: MlcOptions = {
      completions: () => {
        throw new Error('offline');
      },
    };
    const { result } = await complete('{{', opt);
    expect(result!.options.map((o) => o.label)).toEqual(['{{doc_id}}', '{{version}}', '{{title}}']);
  });
});
