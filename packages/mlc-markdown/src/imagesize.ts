// Image size after "|" in the alt text, as in Obsidian but in the plain
// Markdown form: ![Foto|200](a.png), ![Foto|200x100](a.png), ![Foto|50%](a.png).
// Port of mlcext/imagesize.go (same vectors: mlcext/testdata/imagesize.json);
// the check with messages lives in Go.

/** Width/height in CSS px (height only for a box), or percent of the text column. */
export interface ImageSize {
  width?: number;
  height?: number;
  percent?: number;
}

const RE_SIZE = /^(?:(\d+)|(\d+)x(\d+)|(\d+)%)$/;

/** Reads "200", "200x100" or "50%"; null when the text is no valid size. */
export function parseImageSize(s: string): ImageSize | null {
  const m = RE_SIZE.exec(s.trim());
  if (!m) return null;
  if (m[4] !== undefined) {
    const percent = Number(m[4]);
    return percent >= 1 && percent <= 100 ? { percent } : null;
  }
  if (m[1] !== undefined) {
    const width = Number(m[1]);
    return width > 0 ? { width } : null;
  }
  const width = Number(m[2]);
  const height = Number(m[3]);
  return width > 0 && height > 0 ? { width, height } : null;
}

/** The written form: "200", "200x100", "50%", "" for no size. */
export function formatImageSize(s: ImageSize): string {
  if (s.percent) return `${s.percent}%`;
  if (s.width && s.height) return `${s.width}x${s.height}`;
  if (s.width) return String(s.width);
  return '';
}

/**
 * Splits an alt text into the text and the size after its last "|". Text
 * after "|" that does not start with a digit stays part of the text; a size
 * that cannot be read gives error = true and the text before "|".
 */
export function splitImageAlt(alt: string): { text: string; size: ImageSize; error: boolean } {
  const k = alt.lastIndexOf('|');
  if (k < 0) return { text: alt, size: {}, error: false };
  const raw = alt.slice(k + 1).trim();
  if (!/^[0-9]/.test(raw)) return { text: alt, size: {}, error: false };
  const size = parseImageSize(raw);
  return { text: alt.slice(0, k).replace(/ +$/, ''), size: size ?? {}, error: size === null };
}

/** Writes text and size as an alt text ("Foto|200"). */
export function imageAlt(text: string, size: ImageSize): string {
  const s = formatImageSize(size);
  return s ? `${text}|${s}` : text;
}

/** An inline image in a text: offsets of the whole image and of its alt text. */
export interface ImageSpan {
  from: number;
  to: number;
  altFrom: number;
  altTo: number;
  alt: string;
  /** The target as written, without <…>. */
  target: string;
}

// Same as mlcext's reImage: alt, target (plain or <with spaces>), title.
const RE_IMAGE = /!\[([^\]\n]*)\]\(\s*(<[^>\n]+>|[^)\s]+)(\s+"[^"\n]*")?\s*\)/g;

/** The inline images ![alt](target "title") of text, in order. */
export function imageSpans(text: string): ImageSpan[] {
  const out: ImageSpan[] = [];
  for (const m of text.matchAll(RE_IMAGE)) {
    const from = m.index;
    const t = m[2];
    out.push({
      from,
      to: from + m[0].length,
      altFrom: from + 2,
      altTo: from + 2 + m[1].length,
      alt: m[1],
      target: t.startsWith('<') ? t.slice(1, -1) : t,
    });
  }
  return out;
}
