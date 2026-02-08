export const c_initialmd = `
# MD Viewer — Beautiful Markdown, Instantly ✨

**Live preview • KaTeX math • Mermaid diagrams • Syntax highlighting**

Welcome to **MD Viewer** — a fast local Markdown editor and previewer built for developers and technical writers. Start typing on the left and watch the right-hand preview update live.

---

## Why you'll love it

- Blazing-fast rendering using Goldmark and Chroma for syntax highlighting.
- Interactive Mermaid diagrams rendered on-the-fly.
- Beautiful math via KaTeX (inline and block LaTeX supported).
- Multiple preview themes (Dark / Light / Sepia / Monochrome).

## Quick Demo

Type or paste Markdown on the left — the preview updates automatically.

### Mermaid
\`\`\`mermaid
sequenceDiagram
    participant U as User
    participant E as Editor
    participant B as Backend
    U->>E: edit markdown
    E->>B: render request
    B-->>E: sanitized HTML
    E->>U: update preview
\`\`\`

### Math (KaTeX)
Inline: $e^{i\pi} + 1 = 0$

Block:
$$
\int_{-\infty}^{\infty} e^{-x^2} \,dx = \sqrt{\pi}
$$

## Features at a glance
| Feature | Status |
| :--- | :---: |
| Live Preview | ✅ |
| Mermaid Diagrams | ✅ |
| KaTeX Math | ✅ |
| Syntax Highlighting | ✅ |
| Export HTML | ✅ |

---

Made with ❤️ — open an issue or contribute on GitHub.
`;