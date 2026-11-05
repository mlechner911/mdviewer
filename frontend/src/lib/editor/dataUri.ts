// Inline images as base64 (data: URLs) are shown shortened in the editor:
// "🖼 logo.png ‹2,8 MB›". The text stays unchanged (copy, save,
// search see all of it); a click on the box shows it in full. Long base64
// lines otherwise make the editor slow and the document unreadable.
import { Prec, RangeSetBuilder, StateEffect, StateField, type EditorState, type Extension } from '@codemirror/state';
import { Decoration, EditorView, WidgetType, keymap, type DecorationSet } from '@codemirror/view';

/** Only payloads this long are shortened (small icons stay readable). */
const MIN_PAYLOAD = 120;

const expand = StateEffect.define<number>();
const collapse = StateEffect.define<number>();

/**
 * Start offsets of payloads the user opened; mapped through edits. A payload
 * closes again by its "▲" button at its start, by Escape, or when the cursor
 * leaves its line (a payload wraps over many rows, so arrow keys stay in it).
 */
const expanded = StateField.define<Set<number>>({
  create: () => new Set(),
  update(set, tr) {
    let next = set;
    if (tr.docChanged && set.size) {
      next = new Set([...set].map((p) => tr.changes.mapPos(p)));
    }
    for (const e of tr.effects) {
      if (e.is(expand)) next = new Set([...next, e.value]);
      if (e.is(collapse)) next = e.value < 0 ? new Set() : new Set([...next].filter((p) => p !== e.value));
    }
    if (next.size && (tr.selection || tr.docChanged)) {
      const line = tr.state.doc.lineAt(tr.state.selection.main.head).number;
      const kept = [...next].filter((p) => p <= tr.state.doc.length && tr.state.doc.lineAt(p).number === line);
      if (kept.length !== next.size) next = new Set(kept);
    }
    return next;
  },
});

export function formatBytes(n: number, locale: string): string {
  const fmt = (v: number) => v.toLocaleString(locale, { maximumFractionDigits: 1 });
  if (n >= 1024 * 1024) return `${fmt(n / (1024 * 1024))} MB`;
  if (n >= 1024) return `${fmt(n / 1024)} KB`;
  return `${n} B`;
}

class PayloadWidget extends WidgetType {
  constructor(
    readonly from: number,
    readonly label: string,
    readonly title: string,
  ) {
    super();
  }
  eq(other: PayloadWidget) {
    return other.from === this.from && other.label === this.label;
  }
  toDOM(view: EditorView) {
    const el = document.createElement('span');
    el.className = 'cm-data-uri';
    el.textContent = this.label;
    el.title = this.title;
    el.setAttribute('role', 'button');
    el.addEventListener('mousedown', (e) => {
      e.preventDefault();
      // The cursor goes into the payload's line: leaving it closes it again.
      view.dispatch({ effects: expand.of(this.from), selection: { anchor: this.from } });
      view.focus();
    });
    return el;
  }
  ignoreEvent() {
    return false;
  }
}

class CollapseWidget extends WidgetType {
  constructor(
    readonly from: number,
    readonly title: string,
  ) {
    super();
  }
  eq(other: CollapseWidget) {
    return other.from === this.from;
  }
  toDOM(view: EditorView) {
    const el = document.createElement('span');
    el.className = 'cm-data-uri';
    el.textContent = '▲';
    el.style.marginRight = '0.25em';
    el.title = this.title;
    el.setAttribute('role', 'button');
    el.addEventListener('mousedown', (e) => {
      e.preventDefault();
      view.dispatch({ effects: collapse.of(this.from) });
    });
    return el;
  }
  ignoreEvent() {
    return false;
  }
}

export interface DataUriLabels {
  /** "Bild" … */
  image: string;
  /** Tooltip, e.g. "Klicken zeigt den vollständigen Inhalt". */
  expandHint: string;
  /** Tooltip of the button that shortens it again. */
  collapseHint: string;
  locale: string;
}

const PAYLOAD = /data:([a-zA-Z]+\/[a-zA-Z0-9.+-]+)?;base64,([A-Za-z0-9+/=]+)(?:[ \t]+"([^"\n]*)")?/g;

/**
 * The shortened payloads of the whole document. Computed for all of it, not
 * for the visible part: a payload is one line of megabytes, and a search that
 * starts in the middle of it (scrolling) would not find its "data:".
 */
function payloads(state: EditorState, labels: DataUriLabels): DecorationSet {
  const opened = state.field(expanded);
  const b = new RangeSetBuilder<Decoration>();
  for (let n = 1; n <= state.doc.lines; n++) {
    const line = state.doc.line(n);
    if (line.length < MIN_PAYLOAD || !line.text.includes(';base64,')) continue;
    for (const m of line.text.matchAll(PAYLOAD)) {
      const payload = m[2];
      if (payload.length < MIN_PAYLOAD) continue;
      const from = line.from + (m.index ?? 0);
      const start = from + m[0].indexOf(';base64,') + ';base64,'.length;
      if (opened.has(start)) {
        b.add(start, start, Decoration.widget({ widget: new CollapseWidget(start, labels.collapseHint), side: -1 }));
        continue;
      }
      const type = (m[1] ?? '').split('/')[1]?.toUpperCase() ?? '';
      const bytes = Math.floor((payload.length * 3) / 4) - (payload.endsWith('==') ? 2 : payload.endsWith('=') ? 1 : 0);
      const size = formatBytes(bytes, labels.locale);
      const label = m[3] ? `🖼 ${m[3]} ‹${size}›` : `🖼 ‹${[labels.image, type].filter(Boolean).join(' ')}, ${size}›`;
      b.add(from, from + m[0].length, Decoration.replace({ widget: new PayloadWidget(start, label, labels.expandHint) }));
    }
  }
  return b.finish();
}

export function dataUriFolding(labels: DataUriLabels): Extension {
  const shortened = StateField.define<DecorationSet>({
    create: (state) => payloads(state, labels),
    update(deco, tr) {
      const toggled = tr.startState.field(expanded) !== tr.state.field(expanded);
      return tr.docChanged || toggled ? payloads(tr.state, labels) : deco;
    },
    provide: (f) => EditorView.decorations.from(f),
  });
  return [
    expanded,
    shortened,
    // The cursor jumps over a shortened payload instead of into it.
    Prec.high(
      keymap.of([
        {
          key: 'Escape',
          run: (view) => {
            if (!view.state.field(expanded).size) return false;
            view.dispatch({ effects: collapse.of(-1) });
            return true;
          },
        },
      ]),
    ),
    EditorView.atomicRanges.of((view) => view.state.field(shortened)),
    EditorView.baseTheme({
      '.cm-data-uri': {
        padding: '0 0.35em',
        borderRadius: '3px',
        border: '1px solid rgba(127, 127, 127, 0.45)',
        background: 'rgba(127, 127, 127, 0.15)',
        fontSize: '0.85em',
        cursor: 'pointer',
        whiteSpace: 'nowrap',
      },
    }),
  ];
}
