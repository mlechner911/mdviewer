// An image ![alt](target) can be edited in the image dialog: while the
// cursor is in it, a small box after it ("✎ Bild") opens the dialog; a
// double click on the image's syntax does the same. Positions are
// offsets in the document.
import { type Extension, type EditorState } from '@codemirror/state';
import { Decoration, EditorView, ViewPlugin, WidgetType, type DecorationSet, type ViewUpdate } from '@codemirror/view';
import { imageSpans, type ImageSpan } from 'mlc-markdown';

/** The image at pos (offsets in the document), or null. */
export function imageAt(state: EditorState, pos: number): ImageSpan | null {
  const line = state.doc.lineAt(pos);
  for (const s of imageSpans(line.text)) {
    if (pos >= line.from + s.from && pos <= line.from + s.to) {
      return { ...s, from: line.from + s.from, to: line.from + s.to, altFrom: line.from + s.altFrom, altTo: line.from + s.altTo };
    }
  }
  return null;
}

class ImageWidget extends WidgetType {
  constructor(
    readonly label: string,
    readonly title: string,
    readonly image: ImageSpan,
    readonly edit: (img: ImageSpan) => void,
  ) {
    super();
  }
  eq(o: ImageWidget) {
    return o.label === this.label && o.image.from === this.image.from && o.image.to === this.image.to;
  }
  toDOM() {
    const el = document.createElement('span');
    el.className = 'cm-mlc-field-hint';
    el.textContent = this.label;
    el.title = this.title;
    el.setAttribute('role', 'button');
    el.addEventListener('mousedown', (e) => {
      e.preventDefault();
      this.edit(this.image);
    });
    return el;
  }
  ignoreEvent() {
    return true;
  }
}

export interface ImageHintOptions {
  /** Text of the box, e.g. "✎ Bild". */
  label: string;
  title: string;
  onEdit: (img: ImageSpan) => void;
}

export function imageHint(opt: ImageHintOptions): Extension {
  return [
    ViewPlugin.fromClass(
      class {
        decorations: DecorationSet = Decoration.none;
        constructor(view: EditorView) {
          this.compute(view.state);
        }
        update(u: ViewUpdate) {
          if (u.docChanged || u.selectionSet) this.compute(u.state);
        }
        compute(state: EditorState) {
          const img = imageAt(state, state.selection.main.head);
          this.decorations = img
            ? Decoration.set([Decoration.widget({ widget: new ImageWidget(opt.label, opt.title, img, opt.onEdit), side: 1 }).range(img.to)])
            : Decoration.none;
        }
      },
      { decorations: (v) => v.decorations },
    ),
    EditorView.domEventHandlers({
      dblclick(e, view) {
        const pos = view.posAtCoords({ x: e.clientX, y: e.clientY });
        const img = pos == null ? null : imageAt(view.state, pos);
        // A double click in the alt text still selects a word there.
        if (!img || (pos! > img.altFrom && pos! < img.altTo)) return false;
        e.preventDefault();
        opt.onEdit(img);
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
