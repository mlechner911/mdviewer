// Light and dark editor themes. Every colour is a CSS variable, so a host
// sets its own palette on any ancestor (--mlc-editor-bg: …) without
// rebuilding the editor; the fallbacks are MarkSafe's colours.
import { EditorView } from '@codemirror/view';
import { HighlightStyle } from '@codemirror/language';
import { tags as t } from '@lezer/highlight';

/**
 * CSS variables and their fallbacks per theme (documented in the README).
 * @type {Record<'light' | 'dark', Record<string, string>>}
 */
export const PALETTE = {
  dark: {
    fg: '#f1f5f9',
    bg: '#0f172a',
    caret: '#60a5fa',
    selection: 'rgba(59, 130, 246, 0.35)',
    'gutter-bg': '#0f172a',
    'gutter-fg': '#64748b',
    'gutter-border': '#334155',
    'active-line': 'rgba(255, 255, 255, 0.03)',
    'active-gutter-bg': '#1e293b',
    'active-gutter-fg': '#94a3b8',
    placeholder: '#64748b',
    h1: '#60a5fa',
    h2: '#38bdf8',
    h3: '#818cf8',
    h4: '#a78bfa',
    strong: '#f8fafc',
    emphasis: '#cbd5e1',
    code: '#f472b6',
    'code-bg': 'rgba(244, 114, 182, 0.1)',
    string: '#34d399',
    comment: '#64748b',
    link: '#60a5fa',
    url: '#93c5fd',
    quote: '#94a3b8',
    list: '#f59e0b',
  },
  light: {
    fg: '#0f172a',
    bg: '#ffffff',
    caret: '#2563eb',
    selection: 'rgba(59, 130, 246, 0.2)',
    'gutter-bg': '#f8fafc',
    'gutter-fg': '#94a3b8',
    'gutter-border': '#e2e8f0',
    'active-line': 'rgba(0, 0, 0, 0.02)',
    'active-gutter-bg': '#e2e8f0',
    'active-gutter-fg': '#475569',
    placeholder: '#94a3b8',
    h1: '#1d4ed8',
    h2: '#0284c7',
    h3: '#4f46e5',
    h4: '#7c3aed',
    strong: '#0f172a',
    emphasis: '#334155',
    code: '#db2777',
    'code-bg': 'rgba(219, 39, 119, 0.08)',
    string: '#059669',
    comment: '#64748b',
    link: '#2563eb',
    url: '#3b82f6',
    quote: '#475569',
    list: '#d97706',
  },
};

/**
 * @param {'light' | 'dark'} mode
 * @param {string} name
 */
function v(mode, name) {
  return `var(--mlc-editor-${name}, ${PALETTE[mode][name]})`;
}

/** @param {'light' | 'dark'} mode */
function editorTheme(mode) {
  /** @param {string} name */
  const c = (name) => v(mode, name);
  return EditorView.theme(
    {
      '&': {
        color: c('fg'),
        backgroundColor: c('bg'),
        height: '100%',
        fontSize: 'var(--mlc-editor-font-size, 14px)',
        fontFamily: 'var(--mlc-editor-font, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace)',
      },
      '.cm-content': { caretColor: c('caret'), padding: '16px 8px' },
      '&.cm-focused .cm-cursor': { borderLeftColor: c('caret') },
      '&.cm-focused .cm-selectionBackground, ::selection': { backgroundColor: `${c('selection')} !important` },
      '.cm-gutters': { backgroundColor: c('gutter-bg'), color: c('gutter-fg'), borderRight: `1px solid ${c('gutter-border')}`, paddingRight: '8px' },
      '.cm-activeLineGutter': { backgroundColor: c('active-gutter-bg'), color: c('active-gutter-fg') },
      '.cm-activeLine': { backgroundColor: c('active-line') },
      '.cm-placeholder': { color: c('placeholder') },
    },
    { dark: mode === 'dark' },
  );
}

/** @param {'light' | 'dark'} mode */
function highlightStyle(mode) {
  /** @param {string} name */
  const c = (name) => v(mode, name);
  return HighlightStyle.define([
    { tag: t.heading1, color: c('h1'), fontWeight: 'bold', fontSize: '1.2em' },
    { tag: t.heading2, color: c('h2'), fontWeight: 'bold', fontSize: '1.1em' },
    { tag: t.heading3, color: c('h3'), fontWeight: 'bold', fontSize: '1.05em' },
    { tag: [t.heading4, t.heading5, t.heading6], color: c('h4'), fontWeight: 'bold' },
    { tag: t.strong, fontWeight: 'bold', color: c('strong') },
    { tag: t.emphasis, fontStyle: 'italic', color: c('emphasis') },
    { tag: t.monospace, color: c('code'), backgroundColor: c('code-bg'), borderRadius: '3px' },
    { tag: [t.processingInstruction, t.string, t.inserted], color: c('string') },
    { tag: [t.meta, t.comment], color: c('comment'), fontStyle: 'italic' },
    { tag: t.link, color: c('link'), textDecoration: 'underline' },
    { tag: t.url, color: c('url') },
    { tag: t.quote, color: c('quote'), fontStyle: 'italic' },
    { tag: t.list, color: c('list') },
  ]);
}

export const themes = {
  dark: { editor: editorTheme('dark'), highlight: highlightStyle('dark') },
  light: { editor: editorTheme('light'), highlight: highlightStyle('light') },
};
