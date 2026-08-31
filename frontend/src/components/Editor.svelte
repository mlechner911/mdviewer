<script lang="ts">
  /**
   * CodeMirror 6 Markdown Editor component for MarkSafe.
   * Full Dark/Light theme support, syntax highlighting, line numbers, and smooth sync scroll.
   */
  import { onMount, untrack } from 'svelte';
  import { basicSetup, EditorView } from 'codemirror';
  import { EditorState, Compartment } from '@codemirror/state';
  import { markdown } from '@codemirror/lang-markdown';
  import { HighlightStyle, syntaxHighlighting } from '@codemirror/language';
  import { tags as t } from '@lezer/highlight';
  import type { EffectiveTheme_t } from '../lib/constants';

  // --- Svelte 5 Runes: Props ---
  let { 
    value = $bindable(""), 
    theme = 'dark',
    placeholder = "",
    onchange,
    onscroll
  } = $props<{
    value: string;
    theme: EffectiveTheme_t;
    placeholder?: string;
    onchange?: (val: string) => void;
    onscroll?: (percentage: number) => void;
  }>();

  let containerEl = $state<HTMLDivElement>();
  let editorView: EditorView | null = null;
  let isUpdatingFromProp = false;
  let scrollLock = false;

  const themeCompartment = new Compartment();
  const highlightCompartment = new Compartment();

  // Dark Theme Definition
  const marksafeDarkTheme = EditorView.theme({
    "&": {
      color: "#f1f5f9",
      backgroundColor: "#0f172a",
      height: "100%",
      fontSize: "14px",
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace"
    },
    ".cm-content": {
      caretColor: "#60a5fa",
      padding: "16px 8px"
    },
    "&.cm-focused .cm-cursor": {
      borderLeftColor: "#60a5fa"
    },
    "&.cm-focused .cm-selectionBackground, ::selection": {
      backgroundColor: "rgba(59, 130, 246, 0.35) !important"
    },
    ".cm-gutters": {
      backgroundColor: "#0f172a",
      color: "#64748b",
      borderRight: "1px solid #334155",
      paddingRight: "8px"
    },
    ".cm-activeLineGutter": {
      backgroundColor: "#1e293b",
      color: "#94a3b8"
    },
    ".cm-activeLine": {
      backgroundColor: "rgba(255, 255, 255, 0.03)"
    },
    ".cm-placeholder": {
      color: "#64748b"
    }
  }, { dark: true });

  const marksafeDarkHighlighting = HighlightStyle.define([
    { tag: t.heading1, color: "#60a5fa", fontWeight: "bold", fontSize: "1.2em" },
    { tag: t.heading2, color: "#38bdf8", fontWeight: "bold", fontSize: "1.1em" },
    { tag: t.heading3, color: "#818cf8", fontWeight: "bold", fontSize: "1.05em" },
    { tag: [t.heading4, t.heading5, t.heading6], color: "#a78bfa", fontWeight: "bold" },
    { tag: t.strong, fontWeight: "bold", color: "#f8fafc" },
    { tag: t.emphasis, fontStyle: "italic", color: "#cbd5e1" },
    { tag: t.monospace, color: "#f472b6", backgroundColor: "rgba(244, 114, 182, 0.1)", borderRadius: "3px" },
    { tag: [t.processingInstruction, t.string, t.inserted], color: "#34d399" },
    { tag: [t.meta, t.comment], color: "#64748b", fontStyle: "italic" },
    { tag: t.link, color: "#60a5fa", textDecoration: "underline" },
    { tag: t.url, color: "#93c5fd" },
    { tag: t.quote, color: "#94a3b8", fontStyle: "italic" },
    { tag: t.list, color: "#f59e0b" }
  ]);

  // Light Theme Definition
  const marksafeLightTheme = EditorView.theme({
    "&": {
      color: "#0f172a",
      backgroundColor: "#ffffff",
      height: "100%",
      fontSize: "14px",
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace"
    },
    ".cm-content": {
      caretColor: "#2563eb",
      padding: "16px 8px"
    },
    "&.cm-focused .cm-cursor": {
      borderLeftColor: "#2563eb"
    },
    "&.cm-focused .cm-selectionBackground, ::selection": {
      backgroundColor: "rgba(59, 130, 246, 0.2) !important"
    },
    ".cm-gutters": {
      backgroundColor: "#f8fafc",
      color: "#94a3b8",
      borderRight: "1px solid #e2e8f0",
      paddingRight: "8px"
    },
    ".cm-activeLineGutter": {
      backgroundColor: "#e2e8f0",
      color: "#475569"
    },
    ".cm-activeLine": {
      backgroundColor: "rgba(0, 0, 0, 0.02)"
    },
    ".cm-placeholder": {
      color: "#94a3b8"
    }
  }, { dark: false });

  const marksafeLightHighlighting = HighlightStyle.define([
    { tag: t.heading1, color: "#1d4ed8", fontWeight: "bold", fontSize: "1.2em" },
    { tag: t.heading2, color: "#0284c7", fontWeight: "bold", fontSize: "1.1em" },
    { tag: t.heading3, color: "#4f46e5", fontWeight: "bold", fontSize: "1.05em" },
    { tag: [t.heading4, t.heading5, t.heading6], color: "#7c3aed", fontWeight: "bold" },
    { tag: t.strong, fontWeight: "bold", color: "#0f172a" },
    { tag: t.emphasis, fontStyle: "italic", color: "#334155" },
    { tag: t.monospace, color: "#db2777", backgroundColor: "rgba(219, 39, 119, 0.08)", borderRadius: "3px" },
    { tag: [t.processingInstruction, t.string, t.inserted], color: "#059669" },
    { tag: [t.meta, t.comment], color: "#64748b", fontStyle: "italic" },
    { tag: t.link, color: "#2563eb", textDecoration: "underline" },
    { tag: t.url, color: "#3b82f6" },
    { tag: t.quote, color: "#475569", fontStyle: "italic" },
    { tag: t.list, color: "#d97706" }
  ]);

  /**
   * getScrollPercentage calculates scroll progress (0 - 1).
   */
  export function getScrollPercentage(): number {
    if (!editorView) return 0;
    const { scrollTop, scrollHeight, clientHeight } = editorView.scrollDOM;
    if (scrollHeight <= clientHeight) return 0;
    return scrollTop / (scrollHeight - clientHeight);
  }

  /**
   * setScrollPercentage updates editor scroll position.
   */
  export function setScrollPercentage(percentage: number) {
    if (!editorView) return;
    const { scrollHeight, clientHeight } = editorView.scrollDOM;
    scrollLock = true;
    editorView.scrollDOM.scrollTop = percentage * (scrollHeight - clientHeight);
    setTimeout(() => { scrollLock = false; }, 50);
  }

  /**
   * wrapSelection wraps selected text with prefix/suffix (bold, italic, code).
   */
  export function wrapSelection(prefix: string, suffix: string) {
    if (!editorView) return;
    const { from, to } = editorView.state.selection.main;
    const selectedText = editorView.state.sliceDoc(from, to);
    const replacement = prefix + selectedText + suffix;
    editorView.dispatch({
      changes: { from, to, insert: replacement },
      selection: { anchor: from + prefix.length, head: to + prefix.length }
    });
    editorView.focus();
  }

  /**
   * prefixSelection prefixes the current line with markdown token (# , ## ).
   */
  export function prefixSelection(prefix: string) {
    if (!editorView) return;
    const { from } = editorView.state.selection.main;
    const line = editorView.state.doc.lineAt(from);
    editorView.dispatch({
      changes: { from: line.from, to: line.from, insert: prefix },
      selection: { anchor: from + prefix.length, head: from + prefix.length }
    });
    editorView.focus();
  }

  /**
   * Focus the CodeMirror editor.
   */
  export function focus() {
    editorView?.focus();
  }

  function handleScroll() {
    if (scrollLock || !editorView) return;
    const percentage = getScrollPercentage();
    onscroll?.(percentage);
  }

  onMount(() => {
    if (!containerEl) return;

    const isDark = theme === 'dark';
    const startState = EditorState.create({
      doc: value,
      extensions: [
        basicSetup,
        markdown(),
        EditorView.lineWrapping,
        themeCompartment.of(isDark ? marksafeDarkTheme : marksafeLightTheme),
        highlightCompartment.of(syntaxHighlighting(isDark ? marksafeDarkHighlighting : marksafeLightHighlighting)),
        EditorView.updateListener.of((update) => {
          if (update.docChanged && !isUpdatingFromProp) {
            const newDoc = update.state.doc.toString();
            value = newDoc;
            onchange?.(newDoc);
          }
        })
      ]
    });

    editorView = new EditorView({
      state: startState,
      parent: containerEl
    });

    const scrollDOM = editorView.scrollDOM;
    scrollDOM.addEventListener('scroll', handleScroll, { passive: true });

    return () => {
      scrollDOM.removeEventListener('scroll', handleScroll);
      editorView?.destroy();
      editorView = null;
    };
  });

  // Reconfigure theme dynamically on Dark/Light change
  $effect(() => {
    const isDark = theme === 'dark';
    if (editorView) {
      untrack(() => {
        editorView?.dispatch({
          effects: [
            themeCompartment.reconfigure(isDark ? marksafeDarkTheme : marksafeLightTheme),
            highlightCompartment.reconfigure(syntaxHighlighting(isDark ? marksafeDarkHighlighting : marksafeLightHighlighting))
          ]
        });
      });
    }
  });

  // Sync external document changes (tab switch, file open) into CodeMirror
  $effect(() => {
    if (editorView && value !== undefined) {
      const currentDoc = editorView.state.doc.toString();
      if (currentDoc !== value) {
        isUpdatingFromProp = true;
        editorView.dispatch({
          changes: { from: 0, to: currentDoc.length, insert: value }
        });
        isUpdatingFromProp = false;
      }
    }
  });
</script>

<div bind:this={containerEl} class="cm-editor-wrapper h-full w-full overflow-hidden select-text"></div>

<style>
  .cm-editor-wrapper {
    display: flex;
    flex-direction: column;
  }
  :global(.cm-editor) {
    height: 100% !important;
    outline: none !important;
  }
  :global(.cm-scroller) {
    overflow: auto !important;
    font-family: inherit !important;
  }
</style>
