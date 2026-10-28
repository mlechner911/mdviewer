// Form fields "[@…]" of the MLC Markdown Extension (non-standard), recognised
// exactly like the Go checker mlcext (fields.go), which ports the scanner of
// mlcmd2pdf (src/form.rs). Both run the same test vectors
// (mlcext/testdata/fields.json). This module only *finds* fields so they can
// be displayed; a malformed field stays text — the diagnostics come from
// mlcext, never from here.

export type FieldKind = 'check' | 'text' | 'multiline' | 'signature';

export interface Field {
  kind: FieldKind;
  /** Unique; without "#name" it is "<section>_<n>" after the heading above. */
  name: string;
  named: boolean;
  /** Preset value, lines joined with "\n". */
  value: string;
  /** A check box written "[@x]" opens ticked (mlcmd2pdf 3.27.0). */
  checked: boolean;
  /** Who signs a signature field "[@Unterschrift: Zeuge 1]" (mlcmd2pdf 3.28.0). */
  role: string;
  /** Width in characters of a one-line field. */
  chars: number;
  rows: number;
  fillPage: boolean;
  pattern: string;
  tooltip: string;
  /** 1-based line and column (code points) of "[@". */
  line: number;
  col: number;
  /** UTF-16 offsets: start of "[@", end after "]" and after a tooltip comment. */
  start: number;
  markerEnd: number;
  end: number;
}

export interface FieldError {
  code: string;
  line: number;
  col: number;
}

export interface Hint {
  kind: 'tooltip' | 'fill-hint';
  start: number;
  end: number;
  text: string;
}

export interface Scan {
  fields: Field[];
  errors: FieldError[];
  hints: Hint[];
}

const isSpace = (c: string) => /\s/u.test(c);
const isLetterOrNumber = (c: string) => /[\p{L}\p{N}]/u.test(c);

/** Heading text → anchor, exactly like mlcmd2pdf heading_slug and GitHub. */
export function headingSlug(text: string): string {
  let out = '';
  for (const c of text.toLowerCase()) {
    if (c === ' ') out += '-';
    else if (c === '-' || c === '_' || isLetterOrNumber(c)) out += c;
  }
  return out;
}

/** Find every field, malformed field and hint in md. */
export function scanFields(md: string): Scan {
  const s = new Scanner(md);
  s.run();
  return { fields: s.fields, errors: s.errors, hints: s.hints };
}

class Scanner {
  pos = 0;
  fields: Field[] = [];
  errors: FieldError[] = [];
  hints: Hint[] = [];
  section = 'feld';
  sectionCount = 0;
  inHeading = false;

  constructor(readonly src: string) {}

  run() {
    let lineStart = true;
    while (this.pos < this.src.length) {
      const rest = this.src.slice(this.pos);
      if (lineStart) {
        const fence = codeFence(rest);
        if (fence) {
          this.skipFencedBlock(fence);
          continue;
        }
        this.inHeading = false;
        const title = atxHeading(rest);
        if (title !== null) {
          this.inHeading = true;
          this.section = headingSlug(title) || 'feld';
          this.sectionCount = 0;
        }
      }
      const c = String.fromCodePoint(rest.codePointAt(0)!);
      if (c === '\\') {
        let n = 1;
        if (rest.length > 1) n += String.fromCodePoint(rest.codePointAt(1)!).length;
        this.pos += n;
        lineStart = false;
      } else if (c === '`') {
        this.skipCodeSpan();
        lineStart = false;
      } else if (c === '<' && rest.startsWith('<!--')) {
        const end = rest.indexOf('-->');
        const n = end >= 0 ? end + 3 : rest.length;
        this.hint(this.pos, this.pos + n);
        this.pos += n;
        lineStart = false;
      } else if (c === '[' && rest.startsWith('[@')) {
        this.pos += this.field();
        lineStart = false;
      } else if (c === '\n') {
        this.pos += 1;
        lineStart = true;
      } else {
        this.pos += c.length;
        lineStart = lineStart && isSpace(c);
      }
    }
  }

  skipFencedBlock(fence: string) {
    let first = true;
    while (this.pos < this.src.length) {
      const rest = this.src.slice(this.pos);
      const nl = rest.indexOf('\n');
      const len = nl >= 0 ? nl + 1 : rest.length;
      const closes = !first && rest.slice(0, len).trimStart().startsWith(fence);
      this.pos += len;
      first = false;
      if (closes) return;
    }
  }

  skipCodeSpan() {
    const rest = this.src.slice(this.pos);
    let ticks = 0;
    while (ticks < rest.length && rest[ticks] === '`') ticks++;
    const end = rest.slice(ticks).indexOf(rest.slice(0, ticks));
    this.pos += end >= 0 ? ticks + end + ticks : ticks;
  }

  hint(from: number, to: number) {
    const raw = this.src.slice(from, to);
    let kind: Hint['kind'];
    if (raw.startsWith('<!--*')) kind = 'tooltip';
    else if (raw.startsWith('<!--!')) kind = 'fill-hint';
    else return;
    const text = raw.slice(5, raw.endsWith('-->') ? -3 : undefined).trim();
    this.hints.push({ kind, start: from, end: to, text });
  }

  fail(code: string) {
    const [line, col] = position(this.src, this.pos);
    this.errors.push({ code, line, col });
  }

  /** Parse a field at "[@"; returns the characters consumed. */
  field(): number {
    const rest = this.src.slice(this.pos);
    const start = this.pos;
    if (this.inHeading) {
      this.fail('field-in-heading');
      return 2;
    }
    const closeAt = unescaped(rest.slice(2), ']');
    if (closeAt < 0) {
      this.fail('field-unclosed');
      return 2;
    }
    let body = rest.slice(2, 2 + closeAt);
    let n = 2 + closeAt + 1;
    if (rest.slice(n).startsWith('(')) return 2; // "[@team](https://…)" is a link
    const ls = lines(body);
    if (body.includes('\n\n') || ls.slice(1).some((l) => l.trim() === '')) {
      this.fail('field-blank-line');
      return n;
    }
    if (unescaped(body, '|') >= 0) {
      this.fail('field-choice');
      return n;
    }

    let name = '';
    let named = false;
    if (body.startsWith('#')) {
      const after = body.slice(1);
      let end = after.search(/\s/u);
      if (end < 0) end = after.length;
      const id = after.slice(0, end);
      if (id === '' || /[^\p{L}\p{N}_-]/u.test(id)) {
        this.fail('field-name');
        return n;
      }
      name = id;
      named = true;
      const r = after.slice(end).replace(/^[ \t]+/, '');
      body = r.startsWith('\n') ? r.slice(1) : r;
    }

    const markerEnd = start + n;
    let tooltip = '';
    const after = rest.slice(n);
    const gap = after.length - after.replace(/^[ \t]+/, '').length;
    if (after.slice(gap).startsWith('<!--*')) {
      const comment = after.slice(gap + 5);
      const end = comment.indexOf('-->');
      if (end >= 0) {
        tooltip = comment.slice(0, end).split(/\s+/u).filter(Boolean).join(' ');
        this.hint(start + n + gap, start + n + gap + 5 + end + 3);
        n += gap + 5 + end + 3;
      }
    }

    const c = classify(body);
    if ('code' in c) {
      this.fail(c.code);
      return n;
    }
    const f = c.field;
    if (named) {
      f.name = name;
      f.named = true;
    } else {
      this.sectionCount++;
      f.name = `${this.section}_${this.sectionCount}`;
    }
    if (tooltip) f.tooltip = tooltip;
    if (this.fields.some((o) => o.name === f.name)) {
      this.fail('field-duplicate');
      return n;
    }
    [f.line, f.col] = position(this.src, start);
    f.start = start;
    f.markerEnd = markerEnd;
    f.end = start + n;
    this.fields.push(f);
    return n;
  }
}

function newField(): Field {
  return {
    kind: 'text', name: '', named: false, value: '', checked: false, role: '', chars: 0, rows: 1, fillPage: false,
    pattern: '', tooltip: '', line: 0, col: 0, start: 0, markerEnd: 0, end: 0,
  };
}

const cpLen = (s: string) => Array.from(s).length;

// A signature field opens with one of these, German or English, either case.
const SIGNATURE_KEYWORDS = ['unterschrift:', 'signature:'];

function signatureRole(body: string): string | null {
  if (body.includes('\n')) return null;
  const t = body.trim();
  for (const k of SIGNATURE_KEYWORDS) {
    if (t.slice(0, k.length).toLowerCase() === k) return unescape(t.slice(k.length).trim());
  }
  return null;
}

function classify(body: string): { field: Field } | { code: string } {
  const f = newField();
  if (body.trim() === '' && !body.includes('\n')) {
    f.kind = 'check';
    return { field: f };
  }
  // "[@x]": ticked, as on paper and as "- [x]" in a task list.
  if ((body.trim() === 'x' || body.trim() === 'X') && !body.includes('\n')) {
    f.kind = 'check';
    f.checked = true;
    return { field: f };
  }
  const role = signatureRole(body);
  if (role !== null) {
    f.kind = 'signature';
    f.role = role;
    f.tooltip = role;
    return { field: f };
  }
  const t = body.trim();
  if (t === '…' || t === '...') {
    f.kind = 'multiline';
    f.fillPage = true;
    return { field: f };
  }
  let ls = lines(body);
  if (ls.length > 1) {
    f.kind = 'multiline';
    const last = ls[ls.length - 1].trim();
    if (last === '…' || last === '...') {
      ls = ls.slice(0, -1);
      f.fillPage = true;
    }
    f.rows = Math.max(ls.length, 1);
    const values = ls.map((l) => {
      f.chars = Math.max(f.chars, cpLen(unescape(l.trim())));
      return preset(l);
    });
    while (values.length && values[values.length - 1] === '') values.pop();
    f.value = values.join('\n');
    return { field: f };
  }
  f.chars = cpLen(unescape(body.trim()));
  const eq = unescaped(body, '=');
  if (eq >= 0 && !body.slice(0, eq).includes('_') && body.slice(0, eq).trim() !== '') {
    const pattern = unescape(body.slice(0, eq).trim());
    const value = preset(body.slice(eq + 1));
    const limit = cpLen(pattern);
    if (cpLen(value) > limit) return { code: 'field-preset-too-long' };
    Object.assign(f, { chars: limit, value, pattern, tooltip: pattern });
    return { field: f };
  }
  if (body.includes('_')) {
    f.value = preset(body);
    return { field: f };
  }
  const pattern = unescape(body.trim());
  if (cpLen(pattern) > 40) return { code: 'field-not-a-field' };
  f.pattern = pattern;
  f.tooltip = pattern;
  return { field: f };
}

/** UTF-16 index of the first wanted character not escaped with "\", or -1. */
function unescaped(text: string, wanted: string): number {
  let escaped = false;
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (escaped) escaped = false;
    else if (c === '\\') escaped = true;
    else if (c === wanted) return i;
  }
  return -1;
}

const ASCII_PUNCT = /[!-/:-@[-`{-~]/;

/** "\x" read as "x" for ASCII punctuation, as Markdown does. */
export function unescape(text: string): string {
  const cs = Array.from(text);
  let out = '';
  for (let i = 0; i < cs.length; i++) {
    if (cs[i] === '\\' && i + 1 < cs.length && ASCII_PUNCT.test(cs[i + 1])) {
      out += cs[i + 1];
      i++;
    } else out += cs[i];
  }
  return out;
}

/** Preset of one marker line: surrounding underscores and spaces removed. */
function preset(line: string): string {
  const text = line.trim().replace(/^_+/, '');
  const trimmed = text.replace(/_+$/, '');
  let out = trimmed;
  const backslashes = trimmed.length - trimmed.replace(/\\+$/, '').length;
  if (trimmed.length < text.length && backslashes % 2 === 1) out += '_';
  return unescape(out.trim());
}

function codeFence(line: string): string {
  const indent = line.length - line.replace(/^ +/, '').length;
  if (indent > 3) return '';
  const t = line.slice(indent);
  if (t === '' || (t[0] !== '`' && t[0] !== '~')) return '';
  let run = 0;
  while (run < t.length && t[run] === t[0]) run++;
  return run >= 3 ? t.slice(0, run) : '';
}

function atxHeading(rest: string): string | null {
  let line = rest;
  const nl = line.indexOf('\n');
  if (nl >= 0) line = line.slice(0, nl);
  const t = line.replace(/^ +/, '');
  let marks = 0;
  while (marks < t.length && t[marks] === '#') marks++;
  if (marks < 1 || marks > 6) return null;
  let text = t.slice(marks);
  if (text !== '' && text[0] !== ' ' && text[0] !== '\t') return null;
  text = text.trim();
  const i = text.lastIndexOf('{#');
  if (i >= 0 && text.endsWith('}')) text = text.slice(0, i).replace(/[ \t]+$/, '');
  return text.replace(/#+$/, '').trim();
}

/** Split like Rust's str::lines: no trailing empty line, "\r" removed. */
function lines(s: string): string[] {
  if (s === '') return [];
  const parts = s.split('\n');
  if (parts[parts.length - 1] === '') parts.pop();
  return parts.map((p) => p.replace(/\r$/, ''));
}

/** 1-based line and column (code points) of a UTF-16 offset. */
export function position(src: string, offset: number): [number, number] {
  const before = src.slice(0, offset);
  const line = before.split('\n').length;
  const lineStart = before.lastIndexOf('\n') + 1;
  return [line, cpLen(before.slice(lineStart)) + 1];
}
