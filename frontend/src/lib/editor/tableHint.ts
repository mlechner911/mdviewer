// While the cursor is in a pipe table, a small box at the end of its
// separator row ("⇔ Spaltenbreiten") opens the column width menu, so the
// function is found where it is needed. onCursorTable reports whether the
// cursor is in a table (toolbar menu), with the line of the separator row.
import { type Extension, type EditorState } from '@codemirror/state';
import { Decoration, EditorView, ViewPlugin, WidgetType, type DecorationSet, type ViewUpdate } from '@codemirror/view';

const SEP_RE = /^\s*\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)*\|?\s*$/;

/** The separator row (1-based) of the table the cursor is in, or 0. */
export function tableAtCursor(state: EditorState): number {
  const doc = state.doc;
  const here = doc.lineAt(state.selection.main.head).number;
  if (!doc.line(here).text.includes('|')) return 0;
  let first = here;
  while (first > 1 && doc.line(first - 1).text.includes('|') && doc.line(first - 1).text.trim() !== '') first--;
  for (let n = first; n <= Math.min(first + 1, doc.lines); n++) {
    if (n > 1 && SEP_RE.test(doc.line(n).text) && doc.line(n - 1).text.includes('|')) return n;
  }
  return 0;
}

class HintWidget extends WidgetType {
  constructor(
    readonly label: string,
    readonly title: string,
    readonly line: number,
    readonly open: (x: number, y: number, line: number) => void,
  ) {
    super();
  }
  eq(o: HintWidget) {
    return o.label === this.label && o.line === this.line;
  }
  toDOM() {
    const el = document.createElement('span');
    el.className = 'cm-table-hint';
    el.textContent = this.label;
    el.title = this.title;
    el.setAttribute('role', 'button');
    el.addEventListener('mousedown', (e) => {
      e.preventDefault();
      const r = el.getBoundingClientRect();
      this.open(r.left, r.bottom + 2, this.line);
    });
    return el;
  }
  ignoreEvent() {
    return true;
  }
}

export interface TableHintOptions {
  label: string;
  title: string;
  open: (x: number, y: number, line: number) => void;
  onCursorTable?: (separatorLine: number) => void;
}

export function tableHint(opt: TableHintOptions): Extension {
  return [
    ViewPlugin.fromClass(
      class {
        decorations: DecorationSet = Decoration.none;
        line = 0;
        constructor(view: EditorView) {
          this.compute(view);
        }
        update(u: ViewUpdate) {
          if (u.docChanged || u.selectionSet) this.compute(u.view);
        }
        compute(view: EditorView) {
          const line = tableAtCursor(view.state);
          if (line !== this.line) opt.onCursorTable?.(line);
          this.line = line;
          if (!line) {
            this.decorations = Decoration.none;
            return;
          }
          const at = view.state.doc.line(line).to;
          this.decorations = Decoration.set([Decoration.widget({ widget: new HintWidget(opt.label, opt.title, line, opt.open), side: 1 }).range(at)]);
        }
      },
      { decorations: (v) => v.decorations },
    ),
    EditorView.baseTheme({
      '.cm-table-hint': {
        marginLeft: '0.6em',
        padding: '0 0.4em',
        borderRadius: '3px',
        border: '1px solid rgba(127, 127, 127, 0.45)',
        background: 'rgba(127, 127, 127, 0.15)',
        fontSize: '0.8em',
        cursor: 'pointer',
        whiteSpace: 'nowrap',
        fontFamily: 'system-ui, sans-serif',
      },
    }),
  ];
}
