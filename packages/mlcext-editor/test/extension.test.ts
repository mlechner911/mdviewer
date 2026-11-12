import { describe, expect, it } from 'vitest';
import { EditorState } from '@codemirror/state';
import { EditorView } from '@codemirror/view';
import { diagnosticCount } from '@codemirror/lint';
import { isBeingTyped, mlcExtension, refreshAnalysis, mlcSnippets, type MlcDiagnostic, type MlcLists, type MlcSchema } from '../src/index';

const source = { analyze: async () => null };
const diag = (code: string, col = 1): MlcDiagnostic => ({ line: 1, col, severity: 'error', code, message: '' });

describe('mlcSnippets', () => {
  const lists = { prefill: [{ name: 'ort', label: 'Ort', marker: '[@#ort ____]' }] } as unknown as MlcLists;
  const labels = (schema?: MlcSchema | null) => mlcSnippets(lists, schema).map((s) => s.label);

  it('offers the old prefill names without a schema', () => {
    expect(labels()).toContain('Vorausgefüllt: ort');
  });
  it('leaves them out with a schema, the fields stay', () => {
    const got = labels({ id: 'https://schemas.mlcgo.eu/forms/test/v1' } as MlcSchema);
    expect(got).not.toContain('Vorausgefüllt: ort');
    expect(got).toContain('Textfeld');
  });
  it('does not write the legacy document header', () => {
    expect(labels()).not.toContain('Dokumentkopf');
  });
});

describe('completion source', () => {
  // A new function per call makes CodeMirror restart the query for ever: the
  // list never opens and the status stays "pending".
  it('stays the same function', () => {
    const state = EditorState.create({ doc: '[@', extensions: mlcExtension(source) });
    const first = state.languageDataAt('autocomplete', 2);
    const second = state.languageDataAt('autocomplete', 2);
    expect(first.length).toBe(1);
    expect(first[0]).toBe(second[0]);
  });
});

describe('isBeingTyped', () => {
  const doc = EditorState.create({ doc: 'Text [@#] Rest\nNächste' }).doc;
  it('holds back an unfinished field with the cursor in its line', () => {
    expect(isBeingTyped(doc, diag('field-name', 6), 8)).toBe(true);
    expect(isBeingTyped(doc, diag('field-unclosed', 6), 8)).toBe(true);
  });
  it('reports it once the cursor has left', () => {
    expect(isBeingTyped(doc, diag('field-name', 6), 2)).toBe(false);
    expect(isBeingTyped(doc, diag('field-name', 6), doc.length)).toBe(false);
  });
  it('holds back nothing else', () => {
    expect(isBeingTyped(doc, diag('field-duplicate', 6), 8)).toBe(false);
  });
});

describe('extraDiagnostics', () => {
  const analysis = { detected: true, diagnostics: [diag('mlc-own')], fields: [], hints: [], lists: {} as MlcLists };
  const wait = () => new Promise((r) => setTimeout(r, 20));

  it('shows extra findings next to the own ones and re-runs on refreshAnalysis', async () => {
    let extra: MlcDiagnostic[] = [];
    let analyses = 0;
    const view = new EditorView({
      parent: document.body,
      state: EditorState.create({
        doc: 'Zeile eins\nZeile zwei',
        extensions: mlcExtension(
          { analyze: async () => (analyses++, analysis) },
          { extraDiagnostics: () => extra, onAnalysis: (a) => expect(a.diagnostics).toHaveLength(1) },
        ),
      }),
    });
    await wait();
    expect(diagnosticCount(view.state)).toBe(1);

    extra = [{ line: 2, col: 1, severity: 'warning', code: 'x', message: 'from the folder', source: 'Ordner-Prüfung' }];
    view.dispatch({ effects: refreshAnalysis.of(null) });
    await wait();
    expect(analyses).toBe(2);
    expect(diagnosticCount(view.state)).toBe(2);
    view.destroy();
  });
});
