// mlc-markdown — display of Markdown with the MLC Markdown Extension
// (non-standard; docs/MLC-EXTENSIONS.md in the MarkSafe repository).
// Display only: the rules live in the Go module mlcext.
export { createRenderer, renderMarkdown, splitFrontMatter, frontMatterEntries, escapeHtml } from './render.js';
export type { RenderOptions, Rendered } from './render.js';
export { scanFields, headingSlug, unescape } from './fields.js';
export { documentMeta, expandVariables, splitSignature, frontMatterBlock, kindFromRef, VARIABLES } from './meta.js';
export type { DocMeta, Signature } from './meta.js';
export type { Field, FieldKind, FieldError, Hint, Scan } from './fields.js';
export { labelsFor, LABELS } from './labels.js';
export type { Labels, Locale } from './labels.js';
export { enhance } from './enhance.js';
export type { EnhanceOptions } from './enhance.js';
