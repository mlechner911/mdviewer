// Form fields of the MLC Markdown Extension (non-standard) can be edited
// in a dialog instead of by their syntax: while the cursor is in a field,
// a small box after it ("✎ Feld") opens the host's field dialog; a double
// click on the field does the same. The host reads the exact field and
// writes it back (mlcext.FieldAt / FormatField).
import { type Extension, type EditorState } from '@codemirror/state';
import { Decoration, EditorView, ViewPlugin, WidgetType, type DecorationSet, type ViewUpdate } from '@codemirror/view';
import { currentAnalysis, toPos, type MlcFieldSpan } from './index';

/** The field at pos (from the latest analysis), or null. */
export function fieldAt(state: EditorState, pos: number): MlcFieldSpan | null {
  const a = currentAnalysis(state);
  if (!a?.detected) return null;
  for (const f of a.fields ?? []) {
    const from = toPos(state.doc, f.line, f.col);
    const to = toPos(state.doc, f.endLine, f.endCol);
    if (pos >= from && pos <= to) return f;
  }
  return null;
}

/** The field the cursor is in, or null. */
export function fieldAtCursor(state: EditorState): MlcFieldSpan | null {
  return fieldAt(state, state.selection.main.head);
}

class FieldWidget extends WidgetType {
  constructor(
    readonly label: string,
    readonly title: string,
    readonly field: MlcFieldSpan,
    readonly edit: (f: MlcFieldSpan) => void,
  ) {
    super();
  }
  eq(o: FieldWidget) {
    return o.label === this.label && o.field.line === this.field.line && o.field.col === this.field.col;
  }
  toDOM() {
    const el = document.createElement('span');
    el.className = 'cm-mlc-field-hint';
    el.textContent = this.label;
    el.title = this.title;
    el.setAttribute('role', 'button');
    el.addEventListener('mousedown', (e) => {
      e.preventDefault();
      this.edit(this.field);
    });
    return el;
  }
  ignoreEvent() {
    return true;
  }
}

export interface FieldHintOptions {
  /** Text of the box, e.g. "✎ Feld". */
  label: string;
  title: string;
  /** Open the field dialog for the field starting at line, col (1-based). */
  onEdit: (line: number, col: number) => void;
}

export function fieldHint(opt: FieldHintOptions): Extension {
  const edit = (f: MlcFieldSpan) => opt.onEdit(f.line, f.col);
  return [
    ViewPlugin.fromClass(
      class {
        decorations: DecorationSet = Decoration.none;
        constructor(view: EditorView) {
          this.compute(view.state);
        }
        update(u: ViewUpdate) {
          if (u.docChanged || u.selectionSet || currentAnalysis(u.startState) !== currentAnalysis(u.state)) this.compute(u.state);
        }
        compute(state: EditorState) {
          const f = fieldAtCursor(state);
          if (!f) {
            this.decorations = Decoration.none;
            return;
          }
          const at = Math.min(toPos(state.doc, f.endLine, f.endCol), state.doc.length);
          this.decorations = Decoration.set([Decoration.widget({ widget: new FieldWidget(opt.label, opt.title, f, edit), side: 1 }).range(at)]);
        }
      },
      { decorations: (v) => v.decorations },
    ),
    EditorView.domEventHandlers({
      dblclick(e, view) {
        const pos = view.posAtCoords({ x: e.clientX, y: e.clientY });
        const f = pos == null ? null : fieldAt(view.state, pos);
        if (!f) return false;
        e.preventDefault();
        edit(f);
        return true;
      },
    }),
    EditorView.baseTheme({
      '.cm-mlc-field-hint': {
        marginLeft: '0.4em',
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
