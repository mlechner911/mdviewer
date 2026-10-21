// Inline images as base64 (data: URLs) are shown shortened in the editor:
// "data:image/png;base64,‹2,8 MB›". The text stays unchanged (copy, save,
// search see all of it); a click on the box shows it in full. Long base64
// lines otherwise make the editor slow and the document unreadable.
import { StateEffect, StateField, type Extension } from '@codemirror/state';
import { Decoration, EditorView, MatchDecorator, ViewPlugin, WidgetType, type DecorationSet, type ViewUpdate } from '@codemirror/view';

/** Only payloads this long are shortened (small icons stay readable). */
const MIN_PAYLOAD = 120;

const expand = StateEffect.define<number>();

/** Start offsets of payloads the user opened; mapped through edits. */
const expanded = StateField.define<Set<number>>({
  create: () => new Set(),
  update(set, tr) {
    let next = set;
    if (tr.docChanged && set.size) {
      next = new Set([...set].map((p) => tr.changes.mapPos(p)));
    }
    for (const e of tr.effects) if (e.is(expand)) next = new Set([...next, e.value]);
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
      view.dispatch({ effects: expand.of(this.from) });
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
  locale: string;
}

export function dataUriFolding(labels: DataUriLabels): Extension {
  const decorator = new MatchDecorator({
    regexp: /data:([a-zA-Z]+\/[a-zA-Z0-9.+-]+)?;base64,([A-Za-z0-9+/=]+)/g,
    decorate(add, from, _to, m, view) {
      const payload = m[2];
      if (payload.length < MIN_PAYLOAD) return;
      const start = from + m[0].length - payload.length;
      if (view.state.field(expanded).has(start)) return;
      const type = (m[1] ?? '').split('/')[1]?.toUpperCase() ?? '';
      const bytes = Math.floor((payload.length * 3) / 4) - (payload.endsWith('==') ? 2 : payload.endsWith('=') ? 1 : 0);
      const label = `‹${[labels.image, type].filter(Boolean).join(' ')}, ${formatBytes(bytes, labels.locale)}›`;
      add(start, start + payload.length, Decoration.replace({ widget: new PayloadWidget(start, label, labels.expandHint) }));
    },
  });
  const plugin = ViewPlugin.fromClass(
    class {
      decorations: DecorationSet;
      constructor(view: EditorView) {
        this.decorations = decorator.createDeco(view);
      }
      update(u: ViewUpdate) {
        const opened = u.transactions.some((tr) => tr.effects.some((e) => e.is(expand)));
        this.decorations = opened ? decorator.createDeco(u.view) : decorator.updateDeco(u, this.decorations);
      }
    },
    { decorations: (v) => v.decorations },
  );
  return [
    expanded,
    plugin,
    // The cursor jumps over a shortened payload instead of into it.
    EditorView.atomicRanges.of((view) => view.plugin(plugin)?.decorations ?? Decoration.none),
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
