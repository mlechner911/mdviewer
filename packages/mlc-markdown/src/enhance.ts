// Client-side finishing of rendered HTML: Mermaid diagrams and formulas.
// The libraries are passed in (lazy-loaded by the host), so the core has no
// hard dependency on them.

export interface EnhanceOptions {
  /** The mermaid module (import('mermaid')), already initialised by the host. */
  mermaid?: { render(id: string, src: string): Promise<{ svg: string }> };
  /** KaTeX (import('katex')) for elements with data-tex. */
  katex?: { renderToString(tex: string, opts?: Record<string, unknown>): string };
}

let seq = 0;

/** Render diagrams and formulas inside root. Failures leave the source visible. */
export async function enhance(root: Element, opts: EnhanceOptions): Promise<void> {
  if (opts.katex) {
    for (const el of Array.from(root.querySelectorAll<HTMLElement>('.mlc-md-math[data-tex]'))) {
      try {
        el.innerHTML = opts.katex.renderToString(el.dataset.tex ?? '', {
          displayMode: el.classList.contains('mlc-md-math-display'),
          throwOnError: false,
          output: 'htmlAndMathml',
        });
        el.classList.add('mlc-md-math-done');
      } catch {
        /* keep the TeX source */
      }
    }
  }
  if (opts.mermaid) {
    for (const pre of Array.from(root.querySelectorAll<HTMLElement>('pre.mlc-md-mermaid'))) {
      const src = pre.textContent ?? '';
      try {
        const { svg } = await opts.mermaid.render(`mlc-md-mermaid-${++seq}`, src);
        const div = document.createElement('div');
        div.className = 'mlc-md-diagram';
        div.dataset.line = pre.dataset.line ?? '';
        div.innerHTML = svg;
        const tpl = document.createElement('template');
        tpl.className = 'mlc-md-diagram-source';
        tpl.content.append(src);
        div.prepend(tpl);
        pre.replaceWith(div);
      } catch {
        pre.classList.add('mlc-md-mermaid-error');
      }
    }
  }
}
