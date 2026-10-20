import { flushSync, mount, tick, unmount } from 'svelte';
import { describe, expect, it, vi } from 'vitest';
import { MlcMarkdown } from '../src/index.js';

const wait = (ms = 0) => new Promise((r) => setTimeout(r, ms));

function setup(props: Record<string, unknown>) {
  const target = document.createElement('div');
  document.body.append(target);
  const c = mount(MlcMarkdown, { target, props });
  flushSync();
  return { target, c, done: () => unmount(c) };
}

describe('<MlcMarkdown>', () => {
  it('renders fields, alerts and hides front matter', async () => {
    const { target, done } = setup({ source: '---\ntitle: X\n---\n# A\n\n> [!TIP]\n> t\n\nOrt: [@#ort ___]\n', locale: 'de' });
    await tick();
    expect(target.querySelector('h1')!.textContent).toBe('A');
    expect(target.querySelector('.mlc-md-alert-title')!.textContent).toBe('Tipp');
    expect(target.querySelector('[data-field="ort"]')).not.toBeNull();
    expect(target.textContent).not.toContain('title: X');
    done();
  });

  it('front matter as a box', async () => {
    const { target, done } = setup({ source: '---\ntitle: X\nversion: 1\n---\nText', frontMatter: 'box' });
    await tick();
    expect([...target.querySelectorAll('.mlc-md-frontmatter td')].map((t) => t.textContent)).toEqual(['title', 'X', 'version', '1']);
    done();
  });

  it('onNavigate gets anchors, links and refs; true prevents the default', async () => {
    const seen: string[] = [];
    const { target, done } = setup({
      source: '[a](#x) [b](doc.md) [c](https://e.org) EV-2026-0001',
      linkify: (t: string) => t.split(/(EV-\d{4}-\d{4})/).map((p) => (p.startsWith('EV-') ? { text: p, href: `#event=${p}` } : p)),
      onNavigate: (n: { href: string; kind: string }) => (seen.push(`${n.kind}:${n.href}`), n.kind !== 'external'),
    });
    await tick();
    for (const a of target.querySelectorAll('a')) a.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
    expect(seen).toEqual(['anchor:#x', 'link:doc.md', 'external:https://e.org', 'ref:#event=EV-2026-0001']);
    done();
  });

  it('debounces source changes', async () => {
    const onRendered = vi.fn();
    const target = document.createElement('div');
    const props = $state({ source: '# Eins', debounce: 30, onRendered });
    const c = mount(MlcMarkdown, { target, props });
    flushSync();
    await tick();
    props.source = '# Zwei';
    flushSync();
    await tick();
    expect(target.querySelector('h1')!.textContent).toBe('Eins');
    await wait(60);
    expect(target.querySelector('h1')!.textContent).toBe('Zwei');
    unmount(c);
  });

  it('header snippet and data-line', async () => {
    const { target, done } = setup({ source: '# A\n\nB\n' });
    await tick();
    expect([...target.querySelectorAll('[data-line]')].map((e) => e.getAttribute('data-line'))).toEqual(['1', '3']);
    done();
  });

  it('lazy KaTeX for formulas', async () => {
    const loadKatex = vi.fn(async () => ({ renderToString: (t: string) => `<b class="k">${t}</b>` }));
    const { target, done } = setup({ source: 'Formel $a+b$', loadKatex });
    await tick();
    await wait(10);
    expect(loadKatex).toHaveBeenCalledOnce();
    expect(target.querySelector('.mlc-md-math b.k')!.textContent).toBe('a+b');
    done();
  });
});

describe('<MlcMarkdown> robustness', () => {
  it('a failing loader does not stop formulas or onRendered', async () => {
    const onRendered = vi.fn();
    const loadKatex = vi.fn(async () => ({ renderToString: (t: string) => `<b class="k">${t}</b>` }));
    const loadMermaid = vi.fn(async () => {
      throw new Error('Unsupported color format');
    });
    const err = vi.spyOn(console, 'error').mockImplementation(() => {});
    const { target, done } = setup({ source: '$a$\n\n```mermaid\ngraph LR; A-->B;\n```\n', loadKatex, loadMermaid, onRendered });
    await tick();
    await wait(10);
    expect(target.querySelector('.mlc-md-math b.k')).not.toBeNull();
    expect(onRendered).toHaveBeenCalled();
    err.mockRestore();
    done();
  });
});
