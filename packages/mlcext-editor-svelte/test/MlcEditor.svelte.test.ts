import { flushSync, mount, unmount } from 'svelte';
import { describe, expect, it } from 'vitest';
import { EditorView } from '@codemirror/view';
import { EditorState } from '@codemirror/state';
import { forEachDiagnostic } from '@codemirror/lint';
import { CompletionContext, type CompletionResult, type CompletionSource } from '@codemirror/autocomplete';
import type { MlcAnalysis, MlcDiagnostic } from 'mlcext-editor';
import { MlcEditor } from '../src/index.js';

const wait = (ms = 30) => new Promise((r) => setTimeout(r, ms));

type Api = {
  getView(): EditorView | null;
  refresh(): void;
  insertText(t: string): void;
  replaceRange(l: number, c: number, el: number, ec: number, t: string): void;
  wrapSelection(p: string, s?: string): void;
};

function setup(props: Record<string, unknown>) {
  const target = document.createElement('div');
  document.body.append(target);
  const state = $state(props);
  const c = mount(MlcEditor, { target, props: state }) as unknown as Api;
  flushSync();
  return { target, c, props: state, view: () => c.getView()!, done: () => unmount(c) };
}

function messages(view: EditorView): string[] {
  const out: string[] = [];
  forEachDiagnostic(view.state, (d) => out.push(d.message));
  return out.sort();
}

const analysis = (md: string): MlcAnalysis => ({
  detected: true,
  diagnostics: [{ line: 1, col: 1, severity: 'warning', code: 'own', message: 'own' }],
  fields: md.includes('[@#ort')
    ? [{ line: 1, col: md.indexOf('[@') + 1, endLine: 1, endCol: md.indexOf(']') + 2, name: 'ort', named: true, kind: 'text', pattern: '', prefill: false }]
    : [],
  hints: [],
  lists: { eventKinds: [], prefill: [], frontMatterKeys: [] },
});

describe('<MlcEditor>', () => {
  it('shows the text and reports edits through onchange', async () => {
    let changed = '';
    const { view, done } = setup({ value: '# Titel', onchange: (v: string) => (changed = v) });
    expect(view().state.doc.toString()).toBe('# Titel');
    view().dispatch({ changes: { from: 7, insert: '!' } });
    expect(changed).toBe('# Titel!');
    done();
  });

  it('highlights fields from the source and analyses once on start', async () => {
    let calls = 0;
    let seen: MlcAnalysis | null = null;
    const { target, done } = setup({
      value: 'Ort: [@#ort ___]',
      source: { analyze: async (md: string) => (calls++, analysis(md)) },
      onAnalysis: (a: MlcAnalysis) => (seen = a),
    });
    await wait();
    expect(calls).toBe(1);
    expect(seen).not.toBeNull();
    expect(target.querySelector('.cm-mlc-field')?.textContent).toBe('[@#ort ___]');
    done();
  });

  it('shows the host findings next to the analysis', async () => {
    const diagnostics: MlcDiagnostic[] = [{ line: 1, col: 1, severity: 'error', code: 'srv', message: 'vom Server', source: 'Werkstatt' }];
    const { view, done } = setup({ value: 'Zeile', source: { analyze: async (md: string) => analysis(md) }, diagnostics });
    await wait();
    expect(messages(view())).toEqual(['own', 'vom Server']);
    done();
  });

  it('shows host findings without a source', async () => {
    const { view, done } = setup({ value: 'Zeile', diagnostics: [{ line: 1, col: 1, severity: 'info', code: 'x', message: 'nur Host' }] });
    await wait();
    expect(messages(view())).toEqual(['nur Host']);
    done();
  });

  it('passes the host completions through', async () => {
    const { view, done } = setup({ value: '{{st', completions: () => [{ label: '{{steckbrief.name}}' }] });
    await wait();
    const v = view();
    // markdown() brings completion sources of its own; ask them all.
    const labels: string[] = [];
    for (const source of v.state.languageDataAt<CompletionSource>('autocomplete', 4)) {
      const result = (await source(new CompletionContext(v.state, 4, false))) as CompletionResult | null;
      labels.push(...(result?.options.map((o) => o.label) ?? []));
    }
    expect(labels).toContain('{{steckbrief.name}}');
    done();
  });

  it('field box and double click call onEditField', async () => {
    let edited: [number, number] | null = null;
    const { target, view, done } = setup({
      value: 'Ort: [@#ort ___]',
      locale: 'de',
      source: { analyze: async (md: string) => analysis(md) },
      onEditField: (l: number, c: number) => (edited = [l, c]),
    });
    await wait();
    view().dispatch({ selection: { anchor: 8 } });
    const box = target.querySelector('.cm-mlc-field-hint') as HTMLElement;
    expect(box.textContent).toBe('✎ Feld');
    box.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }));
    box.click();
    expect(edited).toEqual([1, 6]);
    done();
  });

  it('host extensions, theme and read-only', async () => {
    const marker = EditorState.languageData.of(() => [{ commentTokens: { line: '%%' } }]);
    const { view, done } = setup({ value: 'x', theme: 'dark', readOnly: true, extensions: [marker] });
    expect(view().state.facet(EditorView.darkTheme)).toBe(true);
    expect(view().state.readOnly).toBe(true);
    expect(view().state.languageDataAt<{ line: string }>('commentTokens', 0)).toContainEqual({ line: '%%' });
    done();
  });

  it('follows prop changes: value, findings, theme', async () => {
    const { props, view, done } = setup({ value: 'eins', theme: 'light', diagnostics: [] });
    props.value = 'zwei';
    flushSync();
    expect(view().state.doc.toString()).toBe('zwei');
    props.diagnostics = [{ line: 1, col: 1, severity: 'warning', code: 'x', message: 'neu' }];
    flushSync();
    await wait();
    expect(messages(view())).toEqual(['neu']);
    props.diagnostics = [];
    flushSync();
    await wait();
    expect(messages(view())).toEqual([]);
    props.theme = 'dark';
    flushSync();
    expect(view().state.facet(EditorView.darkTheme)).toBe(true);
    done();
  });

  it('a new source analyses again, other prop changes do not', async () => {
    let first = 0;
    let second = 0;
    const { props, done } = setup({ value: 'a', source: { analyze: async (md: string) => (first++, analysis(md)) } });
    await wait();
    props.theme = 'dark';
    props.placeholder = 'Text …';
    flushSync();
    await wait();
    expect(first).toBe(1);
    props.source = { analyze: async (md: string) => (second++, analysis(md)) };
    flushSync();
    await wait();
    expect(second).toBe(1);
    done();
  });

  it('editing helpers work in code points like the analysis', async () => {
    const { c, view, done } = setup({ value: 'Grüße [@#ort ___] Ende' });
    c.replaceRange(1, 7, 1, 18, '[@#ort Raum]');
    expect(view().state.doc.toString()).toBe('Grüße [@#ort Raum] Ende');
    view().dispatch({ selection: { anchor: 0, head: 5 } });
    c.wrapSelection('**');
    expect(view().state.doc.toString()).toBe('**Grüße** [@#ort Raum] Ende');
    done();
  });
});
