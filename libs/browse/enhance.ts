import renderMathInElement from 'katex/dist/contrib/auto-render';

// Post-render enhancements shared by every browse host (KaTeX math,
// Mermaid diagrams, code copy buttons). Markup comes from the shared Go
// renderer (internal/markdown); this file only finishes it client-side.

const MERMAID_VARS = {
  dark: {
    lineColor: '#adbac7',
    textColor: '#f3f4f6',
    background: '#14161a',
    mainBkg: '#14161a',
    primaryColor: '#818cf8',
    edgeLabelBackground: '#14161a',
    tertiaryColor: '#14161a',
  },
  light: {
    lineColor: '#334155',
    textColor: '#0f172a',
    background: '#ffffff',
    mainBkg: '#f8fafc',
    primaryColor: '#4f46e5',
    edgeLabelBackground: '#ffffff',
    tertiaryColor: '#ffffff',
  },
};

async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    try {
      // Fallback for non-secure contexts.
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.select();
      const ok = document.execCommand('copy');
      ta.remove();
      return ok;
    } catch {
      return false;
    }
  }
}

function addCopyButtons(root: HTMLElement): void {
  root.querySelectorAll('pre').forEach((pre) => {
    if (pre.querySelector(':scope > .copy-btn')) return;
    const code = pre.querySelector('code');
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'copy-btn';
    btn.textContent = 'Copy';
    btn.setAttribute('aria-label', 'Code kopieren');
    btn.addEventListener('click', async () => {
      const ok = await copyText(code?.innerText ?? pre.innerText);
      btn.textContent = ok ? 'Kopiert!' : 'Fehler';
      btn.classList.toggle('copied', ok);
      setTimeout(() => {
        btn.textContent = 'Copy';
        btn.classList.remove('copied');
      }, 1500);
    });
    pre.appendChild(btn);
  });
}

export async function enhanceContent(root: HTMLElement, dark: boolean): Promise<void> {
  // Mermaid (~3MB): loaded on demand, only for pages that need it.
  // The static import would make every page pay the parse cost.
  const fences = root.querySelectorAll('pre code.language-mermaid');
  if (fences.length > 0) {
    // Parse-check each diagram first: unparseable input keeps its
    // readable code block instead of a red "Syntax error" box.
    const { default: mermaid } = await import('mermaid');
    mermaid.initialize({
      startOnLoad: false,
      theme: dark ? 'dark' : 'default',
      themeVariables: dark ? MERMAID_VARS.dark : MERMAID_VARS.light,
      fontFamily: 'inherit',
    });
    const pending: Element[] = [];
    for (const el of fences) {
      const parent = el.parentElement;
      if (!parent || parent.tagName !== 'PRE') continue;
      let ok = true;
      try {
        await mermaid.parse(el.textContent || '');
      } catch {
        ok = false;
      }
      if (!ok) continue;
      const div = document.createElement('div');
      div.className = 'mermaid';
      div.textContent = el.textContent || '';
      parent.replaceWith(div);
      pending.push(div);
    }
    try {
      if (pending.length > 0) {
        await mermaid.run({ querySelector: '.mermaid', suppressErrors: true });
      }
    } catch (err) {
      console.error('Mermaid render failed:', err);
    }
  }
  // Copy buttons on the remaining (real code) blocks.
  addCopyButtons(root);
  // KaTeX: inline $…$, display $$…$$, \(…\) and \[…\].
  try {
    renderMathInElement(root, {
      delimiters: [
        { left: '$$', right: '$$', display: true },
        { left: '$', right: '$', inline: true },
        { left: '\\(', right: '\\)', inline: true },
        { left: '\\[', right: '\\]', display: true },
      ],
      throwOnError: false,
    });
  } catch (err) {
    console.error('KaTeX render failed:', err);
  }
}
