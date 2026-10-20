// Document id, version, title, the variables {{doc_id}}, {{version}},
// {{title}} and the Sensor Monitor's signature block — the display side of
// the rules in the Go module mlcext (meta.go, signature.go). Checked against
// the same cases (mlcext/testdata/meta.json); nothing is validated here.

export interface DocMeta {
  docId: string;
  version: string;
  title: string;
}

export interface Signature {
  /** "end": comment <!--mlc-signature … --> at the end; "front": old block before the front matter. */
  form: 'end' | 'front';
  /** doc_no, meaning, signed_by, signed_at, sha256, alg, key_id, sig … */
  fields: Record<string, string>;
}

/** Document kind from the doc_id prefix (SOP-, VA-, FB-, INFO-), as mlcext.KindFromRef. */
export function kindFromRef(ref: string): 'sop' | 'va' | 'form' | 'info' | '' {
  switch (ref.split('-')[0].toUpperCase()) {
    case 'SOP':
      return 'sop';
    case 'VA':
      return 'va';
    case 'FB':
      return 'form';
    case 'INFO':
    case 'IN':
      return 'info';
  }
  return '';
}

/** The only variable names a document may use. */
export const VARIABLES = ['doc_id', 'version', 'title'] as const;

const SIG_END_RE = /<!--mlc-signature[ \t]*\r?\n([\s\S]*?)-->\s*$/;
const SIG_FRONT_RE = /^---\r?\n([\s\S]*?)\r?\n---\r?\n/;

/** Go's strconv.Unquote for the values the signer writes ("…" with escapes, `…`). */
function unquote(v: string): string {
  if (v.length >= 2 && v[0] === '"' && v[v.length - 1] === '"') {
    try {
      return JSON.parse(v) as string;
    } catch {
      return v;
    }
  }
  if (v.length >= 2 && v[0] === '`' && v[v.length - 1] === '`') return v.slice(1, -1);
  return v;
}

function signatureFields(block: string, front: boolean): Record<string, string> | null {
  let ls = block.replace(/\r\n/g, '\n').split('\n');
  if (front) {
    if (ls[0].trim() !== 'signature:') return null;
    ls = ls.slice(1);
  }
  const fields: Record<string, string> = {};
  let n = 0;
  for (const l of ls) {
    if (!l.trim()) continue;
    if (front && !l.startsWith(' ') && !l.startsWith('\t')) return null; // fail closed
    const t = l.trim();
    const i = t.indexOf(':');
    if (i < 0 || !t.slice(0, i).trim()) return null;
    fields[t.slice(0, i).trim()] = unquote(t.slice(i + 1).trim());
    n++;
  }
  return n ? fields : null;
}

function endSignature(md: string): { at: number; fields: Record<string, string> } | null {
  const m = SIG_END_RE.exec(md);
  if (!m || (m.index > 0 && md[m.index - 1] !== '\n')) return null;
  const fields = signatureFields(m[1], false);
  return fields ? { at: m.index, fields } : null;
}

function frontSignature(md: string): { len: number; fields: Record<string, string> } | null {
  const m = SIG_FRONT_RE.exec(md);
  if (!m) return null;
  const fields = signatureFields(m[1], true);
  return fields ? { len: m[0].length, fields } : null;
}

/** The signature (null when unsigned) and the signed content, md without the block. */
export function splitSignature(md: string): { signature: Signature | null; content: string } {
  const end = endSignature(md);
  if (end) return { signature: { form: 'end', fields: end.fields }, content: md.slice(0, end.at) };
  const front = frontSignature(md);
  if (front) return { signature: { form: 'front', fields: front.fields }, content: md.slice(front.len) };
  return { signature: null, content: md };
}

interface FmBlock {
  raw: string;
  body: string;
  present: boolean;
  /** Lines before the body, counted from the start of the file. */
  bodyLine: number;
}

/** The document's front matter; a leading signature block of the old form is skipped. */
export function frontMatterBlock(md: string): FmBlock {
  const skip = frontSignature(md)?.len ?? 0;
  const before = (md.slice(0, skip).match(/\n/g) ?? []).length;
  const text = md.slice(skip).replace(/^﻿/, '');
  const ls = text.split('\n');
  if (!ls.length || ls[0].replace(/\r$/, '') !== '---') return { raw: '', body: md.slice(skip), present: false, bodyLine: before };
  for (let i = 1; i < ls.length; i++) {
    const t = ls[i].replace(/\r$/, '');
    if (t === '---' || t === '...') {
      const rest = ls.slice(i + 1).join('\n');
      const trimmed = rest.replace(/^[\r\n]+/, '');
      const skipped = (rest.slice(0, rest.length - trimmed.length).match(/\n/g) ?? []).length;
      return { raw: ls.slice(1, i).join('\n'), body: trimmed, present: true, bodyLine: before + i + 1 + skipped };
    }
  }
  return { raw: '', body: md.slice(skip), present: false, bodyLine: before };
}

/** A top-level YAML scalar ("key: value", quotes removed); "" when absent. */
function yamlScalar(raw: string, key: string): string {
  const m = new RegExp(`^${key}[ \\t]*:[ \\t]*(.*)$`, 'm').exec(raw);
  if (!m) return '';
  let v = m[1].trim();
  if (v.startsWith('"')) {
    const end = v.lastIndexOf('"');
    try {
      return end > 0 ? (JSON.parse(v.slice(0, end + 1)) as string) : '';
    } catch {
      return v.slice(1, end > 0 ? end : undefined);
    }
  }
  if (v.startsWith("'")) {
    const end = v.lastIndexOf("'");
    return (end > 0 ? v.slice(1, end) : v.slice(1)).replace(/''/g, "'");
  }
  v = v.replace(/\s+#.*$/, '');
  return v === '~' || v === 'null' || v === '|' || v === '>' ? '' : v;
}

const HDR_RE = /\*\*\s*(Titel|Dokumenten-ID|Dokument-ID|Version|Gültig ab)\s*:\s*\*\*\s*([^|\n*]*)/g;

/** Legacy header lines "**Dokumenten-ID:** … | **Version:** …" (first 40 lines). */
function parseHeader(body: string): { refNo: string; version: string; title: string } {
  const h = { refNo: '', version: '', title: '' };
  const ls = body.split('\n').slice(0, 40);
  for (let i = 0; i < ls.length; i++) {
    const l = ls[i];
    if (i > 0 && l.trim() === '---' && (h.refNo || h.version)) break;
    for (const m of l.matchAll(HDR_RE)) {
      if (m[1] === 'Gültig ab') continue;
      const v = m[2].trim().replace(/\\$/, '').trim();
      if (!v || v.startsWith('[')) continue;
      if (m[1] === 'Titel' && !h.title) h.title = v;
      else if ((m[1] === 'Dokumenten-ID' || m[1] === 'Dokument-ID') && !h.refNo) h.refNo = v;
      else if (m[1] === 'Version' && !h.version) h.version = v;
    }
  }
  return h;
}

/** Id, version and title: front matter first, legacy header lines as a fallback. */
export function documentMeta(md: string): DocMeta {
  const b = frontMatterBlock(md);
  const fm = b.present ? { docId: yamlScalar(b.raw, 'doc_id'), version: yamlScalar(b.raw, 'version'), title: yamlScalar(b.raw, 'title') } : { docId: '', version: '', title: '' };
  const h = parseHeader(b.body);
  return { docId: fm.docId || h.refNo, version: fm.version || h.version, title: fm.title || h.title };
}

function codeFence(line: string): string {
  const indent = line.length - line.replace(/^ +/, '').length;
  if (indent > 3) return '';
  const t = line.slice(indent);
  if (!t || (t[0] !== '`' && t[0] !== '~')) return '';
  let run = 0;
  while (run < t.length && t[run] === t[0]) run++;
  return run < 3 ? '' : t.slice(0, run);
}

function codeSpans(l: string): Array<[number, number]> {
  const spans: Array<[number, number]> = [];
  for (let i = 0; i < l.length; ) {
    if (l[i] !== '`') {
      i++;
      continue;
    }
    let n = 0;
    while (i + n < l.length && l[i + n] === '`') n++;
    const j = l.indexOf(l.slice(i, i + n), i + n);
    if (j >= 0) {
      spans.push([i, j + n]);
      i = j + n;
    } else i += n;
  }
  return spans;
}

const VAR_RE = /\{\{([^{}\n]*)\}\}/g;

/** "{{…}}" outside the front matter, code and a signature block. */
function varRefs(md: string): Array<{ name: string; start: number; end: number }> {
  const b = frontMatterBlock(md);
  const start = md.length - b.body.length;
  const end = Math.max(endSignature(md)?.at ?? md.length, start);
  const refs: Array<{ name: string; start: number; end: number }> = [];
  let inFence = false;
  let fence = '';
  let off = start;
  for (const l of md.slice(start, end).split(/(?<=\n)/)) {
    const f = codeFence(l);
    if (f && (!inFence || l.trim().startsWith(fence))) {
      if (inFence) inFence = false;
      else {
        inFence = true;
        fence = f;
      }
    } else if (!inFence) {
      const code = codeSpans(l);
      for (const m of l.matchAll(VAR_RE)) {
        if (code.some(([a, z]) => m.index! >= a && m.index! < z)) continue;
        refs.push({ name: m[1].trim(), start: off + m.index!, end: off + m.index! + m[0].length });
      }
    }
    off += l.length;
  }
  return refs;
}

/** Replace {{doc_id}}, {{version}}, {{title}}; unknown names and empty values stay visible. */
export function expandVariables(md: string): string {
  const refs = varRefs(md);
  if (!refs.length) return md;
  const meta = documentMeta(md);
  const value: Record<string, string> = { doc_id: meta.docId, version: meta.version, title: meta.title };
  let out = '';
  let last = 0;
  for (const r of refs) {
    const v = (VARIABLES as readonly string[]).includes(r.name) ? value[r.name] : '';
    if (!v) continue;
    out += md.slice(last, r.start) + v;
    last = r.end;
  }
  return out + md.slice(last);
}
