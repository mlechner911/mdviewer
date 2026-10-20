// Markdown → sanitized HTML with the MLC Markdown Extension (non-standard).
//
// Display only: fields, hints, alerts, anchors, page breaks look as close to
// the mlcdoc PDF as HTML allows, but nothing is checked here — malformed
// fields stay text, diagnostics come from the Go checker mlcext.
//
// Pipeline: front matter off → fields and hints become placeholders (so
// underscores are no emphasis) → marked, block by block with data-line →
// DOMPurify (only what the extension and Markdown produce) → placeholders
// become field HTML (escaped, built here) → linkify on text nodes.

import createDOMPurify from 'dompurify';
import { Marked, type Token, type Tokens } from 'marked';
import markedFootnote from 'marked-footnote';
import { headingSlug, scanFields, type Field } from './fields';
import { labelsFor, type Labels } from './labels';

export interface RenderOptions {
  /** Language of the document: alert titles, "Ausfülltipp" … (default de). */
  locale?: string;
  /** A single line break becomes <br> (default false, as in GFM files). */
  breaks?: boolean;
  /** Raw HTML other than page breaks and anchor targets: drop (default) or sanitize. */
  rawHtml?: 'drop' | 'sanitize';
  /** Highlight a code block; return HTML or undefined for plain text. */
  highlight?: (code: string, lang: string) => string | undefined;
  /** Render a formula; return HTML. Without it formulas get data-tex for enhance(). */
  math?: (tex: string, display: boolean) => string;
  /** Rewrite link and image targets (base paths, local resources). */
  resolveUrl?: (url: string, kind: 'link' | 'image') => string;
  /** Turn parts of plain text into links (event numbers, document IDs …). */
  linkify?: (text: string) => Array<string | { text: string; href: string }> | null;
  /** Window for DOMPurify outside a browser (e.g. jsdom). */
  window?: Window & typeof globalThis;
}

export interface Rendered {
  html: string;
  /** Raw front matter (without the --- lines), "" when there is none. */
  frontMatter: string;
  /** Number of lines the front matter occupied (data-line is relative to the file). */
  frontMatterLines: number;
}

const PH_START = '';
const PH_END = '';
const PH_RE = /(\d+)/g;

/** Split a leading front matter block off (--- … --- or ...). */
export function splitFrontMatter(md: string): { frontMatter: string; body: string; lines: number } {
  const text = md.replace(/^﻿/, '');
  const m = /^---[ \t]*\r?\n([\s\S]*?)\r?\n(?:---|\.\.\.)[ \t]*(?:\r?\n|$)/.exec(text);
  if (!m) return { frontMatter: '', body: md, lines: 0 };
  return { frontMatter: m[1], body: text.slice(m[0].length), lines: m[0].split('\n').length - (m[0].endsWith('\n') ? 1 : 0) };
}

/** Front matter as key/value pairs, one level (for a simple display). */
export function frontMatterEntries(raw: string): Array<[string, string]> {
  const out: Array<[string, string]> = [];
  for (const line of raw.split(/\r?\n/)) {
    const m = /^([A-Za-z0-9_-]+)\s*:\s*(.*)$/.exec(line);
    if (m) out.push([m[1], m[2].replace(/^["']|["']$/g, '')]);
    else if (/^\s+\S/.test(line) && out.length) out[out.length - 1][1] += (out[out.length - 1][1] ? ' ' : '') + line.trim();
  }
  return out;
}

export const escapeHtml = (s: string) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');

type Placeholder =
  | { type: 'field'; field: Field }
  | { type: 'tooltip'; text: string }
  | { type: 'tip'; markdown: string };

/** Replace fields and hints by placeholders, keeping the line count. */
function extractPlaceholders(body: string): { text: string; items: Placeholder[] } {
  const scan = scanFields(body);
  const cuts: Array<{ start: number; end: number; item: Placeholder }> = [];
  for (const f of scan.fields) cuts.push({ start: f.start, end: f.end, item: { type: 'field', field: f } });
  for (const h of scan.hints) {
    if (cuts.some((c) => h.start >= c.start && h.end <= c.end)) continue; // tooltip of a field
    cuts.push({ start: h.start, end: h.end, item: h.kind === 'fill-hint' ? { type: 'tip', markdown: h.text } : { type: 'tooltip', text: h.text } });
  }
  cuts.sort((a, b) => a.start - b.start);
  const items: Placeholder[] = [];
  let out = '';
  let pos = 0;
  for (const c of cuts) {
    const newlines = (body.slice(c.start, c.end).match(/\n/g) ?? []).length;
    out += body.slice(pos, c.start) + PH_START + items.length + PH_END + '\n'.repeat(newlines);
    items.push(c.item);
    pos = c.end;
  }
  return { text: out + body.slice(pos), items };
}

const ALERT_RE = /^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\][ \t]*\n?/i;

/** Column widths like mlcmd2pdf/Pandoc: a separator row over 72 characters spans the full width. */
function tableWidths(raw: string): number[] | null {
  const sep = raw.split('\n')[1]?.trim() ?? '';
  if (sep.length <= 72) return null;
  const dashes = sep.replace(/^\|/, '').replace(/\|$/, '').split('|').map((c) => (c.match(/-/g) ?? []).length);
  const total = dashes.reduce((a, b) => a + b, 0);
  return total > 0 ? dashes.map((d) => (d / total) * 100) : null;
}

const PAGE_BREAK_RE = /^\s*<div class="page-break">\s*<\/div>\s*$/;
const ANCHOR_RE = /^<a id="([\p{L}\p{N}_-]+)">\s*(?:<\/a>)?$/u;

/** A renderer with its own marked and DOMPurify instances (options apply to all calls). */
export function createRenderer(opts: RenderOptions = {}) {
  const labels = labelsFor(opts.locale);
  const win = opts.window ?? (typeof window !== 'undefined' ? window : undefined);
  if (!win) throw new Error('mlc-markdown: a DOM window is needed (browser or options.window)');
  const purify = createDOMPurify(win);
  // style only where this renderer sets it (field widths, table columns)
  purify.addHook('uponSanitizeAttribute', (node, data) => {
    if (data.attrName === 'style' && !(node as Element).hasAttribute?.('data-mlc-style')) data.keepAttr = false;
  });
  const sanitize = (html: string) =>
    purify.sanitize(html, {
      ADD_ATTR: ['data-line', 'data-tex', 'data-field', 'data-mlc-style', 'target'],
      ADD_TAGS: ['details', 'summary'],
      FORBID_TAGS: ['style', 'form', 'textarea', 'select', 'button'],
    }) as unknown as string;

  let ids = new Set<string>();
  const explicitId = (text: string) => /\s*\{#([\p{L}\p{N}_-]+)\}\s*$/u.exec(text);
  const marked = new Marked({ gfm: true, breaks: !!opts.breaks });
  marked.use(markedFootnote({ description: '', refMarkers: true }));
  marked.use({
    extensions: [
      {
        name: 'mathBlock',
        level: 'block',
        start: (src) => src.match(/^\$\$/m)?.index,
        tokenizer(src) {
          const m = /^\$\$([\s\S]+?)\$\$[ \t]*(?:\n|$)/.exec(src);
          if (m) return { type: 'mathBlock', raw: m[0], text: m[1].trim() };
          return undefined;
        },
        renderer: (t) => mathHtml((t as unknown as { text: string }).text, true),
      },
      {
        name: 'mathInline',
        level: 'inline',
        start: (src) => src.indexOf('$'),
        tokenizer(src) {
          const m = /^\$(?![\s$])((?:\\\$|[^$\n])+?)(?<!\s)\$(?!\d)/.exec(src);
          if (m) return { type: 'mathInline', raw: m[0], text: m[1] };
          return undefined;
        },
        renderer: (t) => mathHtml((t as unknown as { text: string }).text, false),
      },
    ],
    renderer: {
      heading(this: { parser: { parseInline(t: Token[]): string } }, { tokens, depth, text }: Tokens.Heading) {
        let id: string;
        let inner = this.parser.parseInline(tokens);
        const ex = explicitId(text);
        if (ex) {
          id = ex[1];
          inner = inner.replace(/\s*\{#[\p{L}\p{N}_-]+\}\s*$/u, '');
        } else {
          const base = headingSlug(plain(text)) || 'heading';
          id = base;
          for (let n = 1; ids.has(id); n++) id = `${base}-${n}`;
        }
        ids.add(id);
        return `<h${depth} id="${escapeHtml(id)}">${inner}</h${depth}>\n`;
      },
      blockquote(this: { parser: { parse(t: Token[]): string } }, { tokens }: Tokens.Blockquote) {
        const first = tokens[0];
        if (first?.type === 'paragraph' && ALERT_RE.test((first as Tokens.Paragraph).raw)) {
          const kind = ALERT_RE.exec((first as Tokens.Paragraph).raw)![1].toLowerCase() as keyof Labels;
          const p = first as Tokens.Paragraph;
          const rest = marked.lexer(p.raw.replace(ALERT_RE, '')) as Token[];
          const body = this.parser.parse([...rest, ...tokens.slice(1)]);
          return `<div class="mlc-md-alert mlc-md-alert-${kind}" role="note"><p class="mlc-md-alert-title">${escapeHtml(labels[kind])}</p>${body}</div>\n`;
        }
        return `<blockquote>${this.parser.parse(tokens)}</blockquote>\n`;
      },
      table(this: { parser: { parseInline(t: Token[]): string } }, token: Tokens.Table) {
        const widths = tableWidths(token.raw);
        const cell = (c: Tokens.TableCell, tag: 'th' | 'td') =>
          `<${tag}${c.align ? ` align="${c.align}"` : ''}>${this.parser.parseInline(c.tokens)}</${tag}>`;
        const cols = widths ? `<colgroup>${widths.map((w) => `<col data-mlc-style style="width:${w.toFixed(1)}%">`).join('')}</colgroup>` : '';
        const head = `<thead><tr>${token.header.map((c) => cell(c, 'th')).join('')}</tr></thead>`;
        const rows = token.rows.map((r) => `<tr>${r.map((c) => cell(c, 'td')).join('')}</tr>`).join('');
        return `<table${widths ? ' class="mlc-md-table-full"' : ''}>${cols}${head}<tbody>${rows}</tbody></table>\n`;
      },
      html({ text, block }: Tokens.HTML | Tokens.Tag) {
        const t = text.trim();
        if (PAGE_BREAK_RE.test(t)) return `<div class="mlc-md-page-break"><span>${escapeHtml(labels.pageBreak)}</span></div>\n`;
        const a = ANCHOR_RE.exec(t);
        if (a) return `<a id="${escapeHtml(a[1])}"></a>`;
        if (t === '</a>' && !block) return '';
        if (opts.rawHtml === 'sanitize') return text;
        return ''; // other raw HTML is not shown (comments, scripts, styling)
      },
      code({ text, lang }: Tokens.Code) {
        const l = (lang ?? '').trim().split(/\s+/)[0];
        if (l === 'mermaid' || l === 'mmd') return `<pre class="mlc-md-mermaid">${escapeHtml(text)}</pre>\n`;
        const hl = l && opts.highlight ? opts.highlight(text, l) : undefined;
        const cls = l ? ` class="language-${escapeHtml(l)}"` : '';
        return `<pre><code${cls}>${hl ?? escapeHtml(text)}</code></pre>\n`;
      },
      link(this: { parser: { parseInline(t: Token[]): string } }, { href, title, tokens }: Tokens.Link) {
        const url = opts.resolveUrl ? opts.resolveUrl(href, 'link') : href;
        const ext = /^https?:\/\//i.test(url);
        return `<a href="${escapeHtml(url)}"${title ? ` title="${escapeHtml(title)}"` : ''}${ext ? ' class="mlc-md-external" rel="noopener noreferrer"' : ''}>${this.parser.parseInline(tokens)}</a>`;
      },
      image({ href, title, text }: Tokens.Image) {
        const url = opts.resolveUrl ? opts.resolveUrl(href, 'image') : href;
        return `<img src="${escapeHtml(url)}" alt="${escapeHtml(text)}"${title ? ` title="${escapeHtml(title)}"` : ''}>`;
      },
    },
  });

  function mathHtml(tex: string, display: boolean): string {
    if (opts.math) return opts.math(tex, display);
    const tag = display ? 'div' : 'span';
    return `<${tag} class="mlc-md-math${display ? ' mlc-md-math-display' : ''}" data-tex="${escapeHtml(tex)}">${escapeHtml(tex)}</${tag}>`;
  }

  /** Render top-level tokens one by one, each block tagged with its source line. */
  function blocks(text: string, firstLine: number): string {
    const tokens = marked.lexer(text);
    let line = firstLine;
    let html = '';
    for (const tok of tokens) {
      const list = Object.assign([tok], { links: tokens.links }) as unknown as Token[] & { links: typeof tokens.links };
      let out = marked.parser(list);
      if (firstLine > 0 && tok.type !== 'space' && out.trim()) {
        out = out.replace(/^(\s*<[a-zA-Z][a-zA-Z0-9]*)/, `$1 data-line="${line}"`);
      }
      html += out;
      line += (tok.raw.match(/\n/g) ?? []).length;
    }
    return html;
  }

  function fieldHtml(f: Field): string {
    const title = f.tooltip && f.tooltip !== f.pattern ? f.tooltip : f.pattern;
    const attrs = ` data-field="${escapeHtml(f.name)}"${title ? ` title="${escapeHtml(title)}"` : ''}`;
    const hint = f.tooltip && f.tooltip !== f.pattern ? `<span class="mlc-md-field-hint">${escapeHtml(f.tooltip)}</span>` : '';
    if (f.kind === 'check') return `<span class="mlc-md-field mlc-md-field-check"${attrs}></span>${hint}`;
    if (f.kind === 'multiline') {
      const rows = Math.max(f.rows, 1);
      const value = f.value.split('\n').map(escapeHtml).join('<br>');
      const fill = f.fillPage ? ' mlc-md-field-fill' : '';
      return `<span class="mlc-md-field mlc-md-field-multiline${fill}"${attrs} data-mlc-style style="min-height:${(rows * 1.6).toFixed(1)}em">${value}</span>${hint}`;
    }
    const w = (Math.min(Math.max(f.chars, 4), 70) * 0.55).toFixed(1);
    const inner = f.value ? escapeHtml(f.value) : f.pattern ? `<span class="mlc-md-field-pattern">${escapeHtml(f.pattern)}</span>` : '';
    return `<span class="mlc-md-field mlc-md-field-line"${attrs} data-mlc-style style="min-width:${w}em">${inner}</span>${hint}`;
  }

  function substitute(html: string, items: Placeholder[]): string {
    // A fill tip alone in its paragraph becomes a block; elsewhere it stays inline.
    html = html.replace(/<p( data-line="\d+")?>(\d+)\s*<\/p>/g, (m, dl: string | undefined, i: string) => {
      const it = items[Number(i)];
      if (it?.type !== 'tip') return m;
      return `<details class="mlc-md-tip"${dl ?? ''}><summary>${escapeHtml(labels.fillTip)}</summary>${renderBody(it.markdown, 0)}</details>`;
    });
    return html.replace(PH_RE, (_m, i: string) => {
      const it = items[Number(i)];
      if (!it) return '';
      if (it.type === 'field') return fieldHtml(it.field);
      if (it.type === 'tooltip') return `<span class="mlc-md-field-hint">${escapeHtml(it.text)}</span>`;
      return `<span class="mlc-md-tip-inline">${escapeHtml(it.markdown)}</span>`;
    });
  }

  function linkifyHtml(html: string): string {
    if (!opts.linkify) return html;
    const root = win!.document.createElement('div');
    root.innerHTML = html;
    const walker = win!.document.createTreeWalker(root, 4 /* SHOW_TEXT */);
    const texts: Text[] = [];
    for (let n = walker.nextNode(); n; n = walker.nextNode()) {
      if (!(n.parentElement?.closest('a, code, pre, .mlc-md-field'))) texts.push(n as Text);
    }
    for (const t of texts) {
      const parts = opts.linkify(t.nodeValue ?? '');
      if (!parts || parts.every((p) => typeof p === 'string')) continue;
      const frag = win!.document.createDocumentFragment();
      for (const p of parts) {
        if (typeof p === 'string') frag.append(p);
        else {
          const a = win!.document.createElement('a');
          a.href = p.href;
          a.textContent = p.text;
          a.className = 'mlc-md-ref';
          frag.append(a);
        }
      }
      t.replaceWith(frag);
    }
    return root.innerHTML;
  }

  /** Body → HTML; firstLine 0 = no data-line (nested content such as a fill tip). */
  function renderBody(body: string, firstLine: number): string {
    const { text, items } = extractPlaceholders(body);
    const raw = blocks(text, firstLine);
    return substitute(sanitize(raw), items);
  }

  /** Render a document (with or without front matter). */
  function render(md: string): Rendered {
    const fm = splitFrontMatter(md ?? '');
    ids = new Set<string>();
    for (const m of fm.body.matchAll(/^#{1,6}[ \t].*\{#([\p{L}\p{N}_-]+)\}[ \t]*$/gmu)) ids.add(m[1]); // explicit ids first, like mlcmd2pdf
    const html = linkifyHtml(renderBody(fm.body, fm.lines + 1));
    return { html, frontMatter: fm.frontMatter, frontMatterLines: fm.lines };
  }

  return { render, sanitize };
}

/** Text of a heading without Markdown markup and an explicit {#id}. */
function plain(text: string): string {
  return text
    .replace(/\s*\{#[\p{L}\p{N}_-]+\}\s*$/u, '')
    .replace(/!\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[`*_~]/g, '')
    .replace(/<[^>]+>/g, '');
}

/** One-off rendering; for repeated calls keep a createRenderer(). */
export function renderMarkdown(md: string, opts: RenderOptions = {}): Rendered {
  return createRenderer(opts).render(md);
}
