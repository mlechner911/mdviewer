import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import { createRenderer, frontMatterEntries, renderMarkdown } from '../src/render';

const html = (md: string, opts = {}) => renderMarkdown(md, opts).html;
const doc = (h: string) => {
  const d = document.createElement('div');
  d.innerHTML = h;
  return d;
};

describe('form fields (display)', () => {
  it('check box, line, pattern, preset, tooltip', () => {
    const d = doc(html('Ort: [@#ort ____Kühlraum A____] <!--* z. B. Labor -->\n\nDatum: [@#d TT.MM.JJJJ] ok: [@]\n'));
    const ort = d.querySelector('[data-field="ort"]')!;
    expect(ort.className).toContain('mlc-md-field-line');
    expect(ort.textContent).toBe('Kühlraum A');
    expect(ort.getAttribute('title')).toBe('z. B. Labor');
    expect(d.querySelector('.mlc-md-field-hint')!.textContent).toBe('z. B. Labor');
    expect(d.querySelector('[data-field="d"] .mlc-md-field-pattern')!.textContent).toBe('TT.MM.JJJJ');
    expect(d.querySelector('[data-field="feld_1"]')!.className).toContain('mlc-md-field-check');
  });

  it('underscores are no emphasis, multi-line field is no <hr>', () => {
    const h = html('[@#vorfall __Zeile 1__\n______\n______]\n\nDanach.');
    expect(h).not.toContain('<hr');
    expect(h).not.toContain('<strong>');
    const f = doc(h).querySelector('[data-field="vorfall"]')!;
    expect(f.className).toContain('mlc-md-field-multiline');
    expect(f.innerHTML).toBe('Zeile 1');
  });

  it('values are escaped, malformed fields stay text', () => {
    const h = html('A [@#s __<b>\\]x__] B and [@#offen ____\n\nC');
    expect(h).not.toContain('<b>');
    expect(doc(h).querySelector('[data-field="s"]')!.textContent).toBe('<b>]x');
    expect(h).toContain('[@#offen ____');
  });

  it('fields in code stay code', () => {
    const h = html('`[@#x ___]`\n\n```\n[@#y]\n```\n');
    expect(h).not.toContain('data-field');
  });

  it('fill tip becomes a collapsible block with Markdown inside', () => {
    const d = doc(html('<!--! **So** arbeiten Sie -->\n\nText', { locale: 'de' }));
    const tip = d.querySelector('details.mlc-md-tip')!;
    expect(tip.querySelector('summary')!.textContent).toBe('Ausfülltipp');
    expect(tip.querySelector('strong')!.textContent).toBe('So');
  });
});

describe('Markdown features', () => {
  it('footnotes come after the text, not before it', () => {
    const d = doc(html('# Titel\n\nText[^1].\n\n[^1]: Fußnote.\n\nEnde.\n'));
    const kids = Array.from(d.children).map((e) => e.tagName.toLowerCase());
    expect(kids[0]).toBe('h1');
    expect(d.lastElementChild!.matches('section[data-footnotes]')).toBe(true);
    expect(d.querySelector('[data-footnotes]')!.textContent).toContain('Fußnote.');
  });

  it('alerts with the title in the document language', () => {
    const d = doc(html('> [!WARNING]\n> Bei mehr als −70 °C umlagern.\n', { locale: 'de' }));
    const a = d.querySelector('.mlc-md-alert-warning')!;
    expect(a.querySelector('.mlc-md-alert-title')!.textContent).toBe('Warnung');
    expect(a.textContent).toContain('umlagern');
    expect(doc(html('> [!NOTE]\n> x', { locale: 'en' })).querySelector('.mlc-md-alert-title')!.textContent).toBe('Note');
    expect(html('> nur ein Zitat')).toContain('<blockquote');
  });

  it('heading anchors like mlcmd2pdf: slug, duplicates, explicit ids first', () => {
    const d = doc(html('# Bericht\n\n## Übersicht\n\n## Übersicht\n\n## Explizit {#fest}\n\n## fest\n'));
    expect([...d.querySelectorAll('h1,h2')].map((h) => h.id)).toEqual(['bericht', 'übersicht', 'übersicht-1', 'fest', 'fest-1']);
    expect(d.querySelector('#fest')!.textContent).toBe('Explizit');
  });

  it('page break and anchor target pass, other raw HTML does not', () => {
    const h = html('<div class="page-break"></div>\n\nText <a id="ziel"></a> und <span onclick="x()">s</span>\n\n<script>alert(1)</script>\n');
    expect(h).toContain('mlc-md-page-break');
    expect(h).toContain('id="ziel"');
    expect(h).not.toMatch(/<script|onclick|<span>s/);
  });

  it('rawHtml sanitize keeps harmless HTML, removes the dangerous parts', () => {
    const d = doc(html('<details><summary>Mehr</summary>\n\nInhalt\n\n</details>\n\nText <a href="x.md" onclick="y()">Link</a> <kbd>Strg</kbd> und <a id="ziel"></a> danach\n\n<script>alert(1)</script>\n', { rawHtml: 'sanitize' }));
    expect(d.querySelector('details summary')!.textContent).toBe('Mehr');
    const a = d.querySelector('a[href="x.md"]')!;
    expect(a.textContent).toBe('Link');
    expect(a.hasAttribute('onclick')).toBe(false);
    expect(d.querySelector('kbd')!.textContent).toBe('Strg');
    expect(d.querySelector('#ziel')).not.toBeNull();
    expect(d.innerHTML).not.toMatch(/<script|alert/);
  });

  it('rawHtml sanitize: URLs in embedded HTML go through resolveUrl', () => {
    const seen: string[] = [];
    const resolveUrl = (u: string, k: string) => (seen.push(`${k}:${u}`), `/r/${u}`);
    const d = doc(html('<img src="a.png" srcset="b.png 2x"> <a href=\'x.md?a=1&amp;b=2\'>L</a>\n\n![m](c.png)\n', { rawHtml: 'sanitize', resolveUrl }));
    expect(seen).toEqual(expect.arrayContaining(['image:a.png', 'link:x.md?a=1&b=2', 'image:c.png']));
    expect(d.querySelector('img')!.getAttribute('src')).toBe('/r/a.png');
    expect(d.querySelector('img')!.hasAttribute('srcset')).toBe(false);
    expect(d.querySelector('a')!.getAttribute('href')).toBe('/r/x.md?a=1&b=2');
  });

  it('tables: wide separator row gives column widths', () => {
    const wide = `| A | B |\n| ${'-'.repeat(60)} | ${'-'.repeat(20)} |\n| 1 | 2 |\n`;
    const d = doc(html(wide));
    expect(d.querySelector('table')!.className).toBe('mlc-md-table-full');
    expect([...d.querySelectorAll('col')].map((c) => (c as HTMLElement).style.width)).toEqual(['75%', '25%']);
    expect(doc(html('| A | B |\n|---|---|\n| 1 | 2 |\n')).querySelector('col')).toBeNull();
  });

  it('task lists, footnotes, mermaid, math placeholders', () => {
    const d = doc(html('- [x] erledigt\n- [ ] offen\n\nText[^1]\n\n[^1]: Fußnote\n\n```mermaid\ngraph LR; A-->B;\n```\n\nInline $E=mc^2$ und\n\n$$x^2$$\n'));
    expect(d.querySelectorAll('input[type=checkbox]').length).toBe(2);
    expect(d.textContent).toContain('Fußnote');
    expect(d.querySelector('pre.mlc-md-mermaid')!.textContent).toContain('A-->B');
    expect([...d.querySelectorAll('.mlc-md-math')].map((m) => (m as HTMLElement).dataset.tex)).toEqual(['E=mc^2', 'x^2']);
  });

  it('highlight, resolveUrl and linkify hooks', () => {
    const d = doc(
      html('```go\nfunc x(){}\n```\n\n![b](bild.png) [l](doc.md) siehe EV-2026-0001.', {
        highlight: (code: string, lang: string) => `<span class="hl-${lang}">${code.trim()}</span>`,
        resolveUrl: (u: string, kind: string) => `/res/${kind}/${u}`,
        linkify: (t: string) => t.split(/(EV-\d{4}-\d{4})/).map((p) => (/^EV-/.test(p) ? { text: p, href: `#event=${p}` } : p)),
      }),
    );
    expect(d.querySelector('code .hl-go')!.textContent).toBe('func x(){}');
    expect(d.querySelector('img')!.getAttribute('src')).toBe('/res/image/bild.png');
    expect(d.querySelector('a[href="/res/link/doc.md"]')).not.toBeNull();
    expect(d.querySelector('a.mlc-md-ref')!.getAttribute('href')).toBe('#event=EV-2026-0001');
  });
});

describe('data-line for scroll sync', () => {
  it('blocks carry their line in the file, after front matter and multi-line fields', () => {
    const md = '---\ntitle: X\n---\n# Titel\n\nAbsatz\n\n[@#a ___\n___\n___]\n\n## Danach\n\n- Liste\n';
    const d = doc(renderMarkdown(md).html);
    const line = (sel: string) => Number(d.querySelector(sel)!.getAttribute('data-line'));
    expect(line('h1')).toBe(4);
    expect(line('p')).toBe(6);
    expect(line('h2')).toBe(12);
    expect(line('ul')).toBe(14);
  });
});

describe('front matter', () => {
  it('is split off and readable as entries', () => {
    const r = renderMarkdown('---\ntitle: "CAPA"\nevent_kinds: [DEVIATION]\n---\n# A\n');
    expect(r.frontMatter).toContain('event_kinds');
    expect(r.html).not.toContain('event_kinds');
    expect(frontMatterEntries(r.frontMatter)).toEqual([['title', 'CAPA'], ['event_kinds', 'DEVIATION']]);
    expect(frontMatterEntries('event_kinds: [OUT_OF_RANGE, "NO_DATA"]')).toEqual([['event_kinds', 'OUT_OF_RANGE, NO_DATA']]);
    expect(frontMatterEntries('event_kinds:\n  - OUT_OF_RANGE\n  - NO_DATA')).toEqual([['event_kinds', 'OUT_OF_RANGE, NO_DATA']]);
  });
});

describe('fixture of the MLC extension', () => {
  it('renders the MarkSafe test document', () => {
    const md = readFileSync(resolve(process.cwd(), '../../internal/markdown/testdata/mlc-extension.md'), 'utf8');
    const r = createRenderer({ locale: 'de' }).render(md);
    const d = doc(r.html);
    expect(d.querySelectorAll('.mlc-md-alert').length).toBe(5);
    expect(d.querySelector('#boxen')).not.toBeNull();
    expect(d.querySelector('.mlc-md-page-break')).not.toBeNull();
    expect(d.querySelectorAll('[data-field]').length).toBe(3);
    expect(d.querySelector('table')).not.toBeNull();
  });
});
