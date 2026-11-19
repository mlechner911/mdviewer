import { flushSync, mount, unmount, type Component } from 'svelte';
import { describe, expect, it } from 'vitest';
import { MlcFieldDialog, MlcImageDialog, imageMarkdown, type MlcFieldSpec } from '../src/index.js';

const wait = (ms = 10) => new Promise((r) => setTimeout(r, ms));

function setup(component: Component<any>, props: Record<string, unknown>) {
  const target = document.createElement('div');
  document.body.append(target);
  const c = mount(component, { target, props });
  flushSync();
  const q = <T extends Element = HTMLElement>(sel: string) => target.querySelector(sel) as T;
  return {
    target,
    q,
    submit: () => q<HTMLButtonElement>('button[type=submit]'),
    input(sel: string, value: string) {
      const el = q<HTMLInputElement | HTMLSelectElement>(sel);
      el.value = value;
      el.dispatchEvent(new Event(el instanceof HTMLSelectElement ? 'change' : 'input', { bubbles: true }));
      flushSync();
    },
    done: () => {
      unmount(c);
      target.remove();
    },
  };
}

/** A stand-in for mlcext.FormatField. */
async function format(s: MlcFieldSpec) {
  if (s.kind === 'pattern' && !s.pattern) return { text: '', error: 'Muster fehlt' };
  const name = s.name ? `#${s.name} ` : '';
  return { text: s.kind === 'check' ? `[@${name.trim()}]` : `[@${name}${'_'.repeat(s.chars || 4)}]`, error: '' };
}

describe('<MlcFieldDialog>', () => {
  it('writes the field through format and inserts it', async () => {
    let applied = '';
    const asked: MlcFieldSpec[] = [];
    const d = setup(MlcFieldDialog, { format: (s: MlcFieldSpec) => (asked.push(s), format(s)), locale: 'de', onApply: (t: string) => (applied = t), onClose() {} });
    d.input('#mlc-fd-name', 'ort');
    d.input('#mlc-fd-chars', '6');
    await wait();
    flushSync();
    expect(asked.at(-1)).toMatchObject({ kind: 'text', name: 'ort', chars: 6 });
    expect(d.q('[data-testid=field-syntax]').textContent).toBe('[@#ort ______]');
    expect(d.q('h3').textContent).toBe('Formularfeld einfügen');
    d.submit().click();
    expect(applied).toBe('[@#ort ______]');
    d.done();
  });

  it('edits a field: starts from its spec, warns about a taken name', async () => {
    const d = setup(MlcFieldDialog, {
      initial: { kind: 'pattern', name: 'datum', pattern: 'TT.MM.JJJJ' },
      usedNames: ['datum', 'ort'],
      format,
      locale: 'en',
      onApply() {},
      onClose() {},
    });
    expect(d.q<HTMLSelectElement>('#mlc-fd-kind').value).toBe('date');
    expect(d.q('h3').textContent).toBe('Edit form field');
    expect(d.target.textContent).not.toContain('already used');
    d.input('#mlc-fd-name', 'ort');
    expect(d.target.textContent).toContain('This name is already used in the document.');
    d.done();
  });

  it('shows why there is no field and blocks applying', async () => {
    const d = setup(MlcFieldDialog, { format, locale: 'de', onApply() {}, onClose() {} });
    d.input('#mlc-fd-kind', 'pattern');
    await wait();
    flushSync();
    expect(d.q('[data-testid=field-error]').textContent).toBe('Muster fehlt');
    expect(d.submit().disabled).toBe(true);
    d.done();
  });

  it('keeps only the answer to the latest input', async () => {
    const slow: ((v: { text: string; error: string }) => void)[] = [];
    const d = setup(MlcFieldDialog, { format: () => new Promise((r) => slow.push(r)), onApply() {}, onClose() {} });
    d.input('#mlc-fd-name', 'neu');
    slow[slow.length - 1]({ text: '[@#neu ___]', error: '' });
    slow[0]({ text: '[@ alt]', error: '' });
    await wait();
    flushSync();
    expect(d.q('[data-testid=field-syntax]').textContent).toBe('[@#neu ___]');
    d.done();
  });

  it('host texts replace the built-in ones', async () => {
    const d = setup(MlcFieldDialog, { format, locale: 'fr', labels: { fieldDialogInsert: 'Insérer un champ', cancel: 'Annuler' }, onApply() {}, onClose() {} });
    expect(d.q('h3').textContent).toBe('Insérer un champ');
    expect(d.target.textContent).toContain('Annuler');
    expect(d.target.textContent).toContain('Insert');
    d.done();
  });
});

const PNG = 'data:image/png;base64,iVBORw0KGgo=';

describe('<MlcImageDialog>', () => {
  it('a pasted picture is embedded with description and size', async () => {
    let applied = '';
    const d = setup(MlcImageDialog, { dataUri: PNG, locale: 'de', onApply: (t: string) => (applied = t), onClose() {} });
    d.input('#mlc-img-alt', 'Logo');
    d.input('#mlc-img-size', '200');
    d.submit().click();
    await wait();
    expect(applied).toBe(`![Logo|200](${PNG})`);
    d.done();
  });

  it('a dropped file in the document folder starts as a reference', async () => {
    let request: unknown = null;
    const d = setup(MlcImageDialog, {
      chosen: Promise.resolve({ name: 'plan.png', size: 2048, embeddable: true, reference: 'bilder/plan.png', path: '/doc/bilder/plan.png' }),
      toMarkdown: async (r: unknown) => ((request = r), '![x](y)'),
      onApply() {},
      onClose() {},
    });
    await wait();
    flushSync();
    expect(d.q<HTMLInputElement>('#mlc-img-alt').value).toBe('plan');
    const [reference, embedded] = d.target.querySelectorAll<HTMLInputElement>('input[type=radio]');
    expect(reference.checked).toBe(true);
    expect(embedded.checked).toBe(false);
    expect(d.target.textContent).toContain('2 KB');
    d.submit().click();
    await wait();
    expect(request).toMatchObject({ alt: 'plan', embed: false, dataUri: null, choice: { path: '/doc/bilder/plan.png' } });
    d.done();
  });

  it('without a host: the chosen file is embedded as data: URL', async () => {
    let applied = '';
    const d = setup(MlcImageDialog, {
      choose: async () => ({ name: 'foto 1.jpg', size: 10, embeddable: true, dataUri: PNG }),
      onApply: (t: string) => (applied = t),
      onClose() {},
    });
    await wait();
    flushSync();
    expect(d.target.textContent).toContain('Embedded in the document');
    expect(d.target.textContent).toContain('outside the document folder');
    d.submit().click();
    await wait();
    expect(applied).toBe(`![foto 1](${PNG})`);
    d.done();
  });

  it('editing keeps the file and changes only the alt text', async () => {
    let applied = '';
    const text = '![Plan|50%](bilder/plan.png)';
    const d = setup(MlcImageDialog, {
      existing: { text, alt: 'Plan|50%', target: 'bilder/plan.png', altFrom: 2, altTo: 10 },
      locale: 'de',
      onApply: (t: string) => (applied = t),
      onClose() {},
    });
    expect(d.q<HTMLInputElement>('#mlc-img-size').value).toBe('50%');
    expect(d.q('h3').textContent).toBe('Bild bearbeiten');
    d.input('#mlc-img-size', '120x80');
    d.submit().click();
    await wait();
    expect(applied).toBe('![Plan|120x80](bilder/plan.png)');
    d.done();
  });

  it('a wrong size blocks applying, errors of the host are shown', async () => {
    const d = setup(MlcImageDialog, {
      dataUri: PNG,
      toMarkdown: async () => {
        throw new Error('Error: Datei nicht lesbar');
      },
      onApply() {},
      onClose() {},
    });
    d.input('#mlc-img-size', '500%');
    expect(d.submit().disabled).toBe(true);
    expect(d.q('#mlc-img-size').getAttribute('aria-invalid')).toBe('true');
    d.input('#mlc-img-size', '');
    d.submit().click();
    await wait();
    flushSync();
    expect(d.q('[role=alert]').textContent).toBe('Datei nicht lesbar');
    d.done();
  });

  it('Escape closes', () => {
    let closed = false;
    const d = setup(MlcImageDialog, { dataUri: PNG, onApply() {}, onClose: () => (closed = true) });
    d.q('[role=dialog]').dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    expect(closed).toBe(true);
    d.done();
  });
});

describe('imageMarkdown', () => {
  it('escapes brackets and wraps targets with spaces', () => {
    expect(imageMarkdown('a [b]', 'bilder/mein plan.png')).toBe('![a \\[b\\]](<bilder/mein plan.png>)');
    expect(imageMarkdown('x', PNG)).toBe(`![x](${PNG})`);
  });
});
