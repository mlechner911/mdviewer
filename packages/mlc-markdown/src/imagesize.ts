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
