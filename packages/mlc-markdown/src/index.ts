// mlc-markdown — display of Markdown with the MLC Markdown Extension
// (non-standard; docs/MLC-EXTENSIONS.md in the MarkSafe repository).
// Display only: the rules live in the Go module mlcext.
export { createRenderer, renderMarkdown, splitFrontMatter, frontMatterEntries, escapeHtml } from './render';
export type { RenderOptions, Rendered } from './render';
export { scanFields, headingSlug, unescape } from './fields';
export type { Field, FieldKind, FieldError, Hint, Scan } from './fields';
export { labelsFor, LABELS } from './labels';
export type { Labels, Locale } from './labels';
export { enhance } from './enhance';
export type { EnhanceOptions } from './enhance';
