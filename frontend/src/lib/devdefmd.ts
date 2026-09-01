/**
 * Per-locale welcome document content (body only — H1 title is from i18n).
 */
export const c_welcomeMarkdown: Record<'en' | 'de' | 'es' | 'fr', string> = {
	en: `**Live preview • KaTeX math • Mermaid diagrams**

Welcome to **MarkSafe** — a fast local Markdown editor and previewer built for developers and technical writers. Start typing on the left and watch the right-hand preview update live.

---

## Why you'll love it

- Blazing-fast rendering using Goldmark and Chroma for syntax highlighting.
- Interactive Mermaid diagrams rendered on-the-fly.
- Beautiful math via KaTeX (inline and block LaTeX supported).
- Multiple preview themes (Dark / Light / Sepia / Monochrome).

## Quick Demo

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

Inline: $e^{i\\pi} + 1 = 0$

Block:
$$
\\int_{-\\infty}^{\\infty} e^{-x^2} \\,dx = \\sqrt{\\pi}
$$
`,
	de: `**Live-Vorschau • KaTeX-Mathematik • Mermaid-Diagramme**

Willkommen bei **MarkSafe** — einem schnellen lokalen Markdown-Editor und Vorschau-Tool für Entwickler und technische Schriftsteller. Beginnen Sie links zu tippen und beobachten Sie, wie sich die rechtsseitige Vorschau in Echtzeit aktualisiert.

---

## Warum Sie es lieben werden

- Blitzschnelle Rendering mittels Goldmark und Chroma für Syntax-Hervorhebung.
- Interaktive Mermaid-Diagramme, die in Echtzeit gerendert werden.
- Schöne Mathematik via KaTeX (inline und block LaTeX unterstützt).
- Mehrere Vorschau-Themes (Dark / Light / Sepia / Monochrome).

## Schnelle Demo

### Mermaid

\`\`\`mermaid
sequenceDiagram
    participant U as Benutzer
    participant E as Editor
    participant B as Backend
    U->>E: markdown bearbeiten
    E->>B: render anfrage
    B-->>E: gesauberter HTML
    E->>U: vorschau aktualisieren
\`\`\`

### Mathematik (KaTeX)

Inline: $e^{i\\pi} + 1 = 0$

Block:
$$
\\int_{-\\infty}^{\\infty} e^{-x^2} \\,dx = \\sqrt{\\pi}
$$
`,
	es: `**Vista previa en vivo • Matemáticas KaTeX • Diagramas Mermaid**

Bienvenido a **MarkSafe** — un editor y visor de Markdown local rápido construido para desarrolladores y escritores técnicos. Comience a escribir a la izquierda y observe cómo la vista previa del lado derecho se actualiza en vivo.

---

## Por qué te encantará

- Renderizado rapidísimo usando Goldmark y Chroma para resaltado de sintaxis.
- Diagramas interactivos de Mermaid renderizados en tiempo real.
- Hermosas matemáticas vía KaTeX (Soporte para LaTeX inline y bloque).
- Múltiples temas de vista previa (Oscuro / Claro / Sepia / monocromo).

## Demo rápida

### Mermaid

\`\`\`mermaid
sequenceDiagram
    participant U as Usuario
    participant E as Editor
    participant B as Backend
    U->>E: editar markdown
    E->>B: solicitud de renderizado
    B-->>E: HTML sanitizado
    E->>U: actualizar vista previa
\`\`\`

### Matemáticas (KaTeX)

Inline: $e^{i\\pi} + 1 = 0$

Block:
$$
\\int_{-\\infty}^{\\infty} e^{-x^2} \\,dx = \\sqrt{\\pi}
$$
`,
	fr: `**Aperçu en direct • Mathématiques KaTeX • Diagrammes Mermaid**

Bienvenue sur **MarkSafe** — un éditeur et visualiseur Markdown local rapide conçu pour les développeurs et rédacteurs techniques. Commencez à taper à gauche et regardez l'aperçu à droite se mettre à jour en direct.

---

## Pourquoi vous allez l'adorer

- Rendu ultra-rapide utilisant Goldmark et Chroma pour la coloration de la syntaxe.
- Diagrammes Mermaid interactifs rendus en temps réel.
- Beaux mathématiques via KaTeX (Support LaTeX inline et bloc).
- Plusieurs thèmes d'aperçu (Sombre / Clair / Sepia / Monochrome).

## Démo rapide

### Mermaid

\`\`\`mermaid
sequenceDiagram
    participant U as Utilisateur
    participant E as Éditeur
    participant B as Backend
    U->>E: éditer markdown
    E->>B: demande de rendu
    B-->>E: HTML sanatisé
    E->>U: mise à jour de l'aperçu
\`\`\`

### Mathématiques (KaTeX)

Inline: $e^{i\\pi} + 1 = 0$

Block:
$$
\\int_{-\\infty}^{\\infty} e^{-x^2} \\,dx = \\sqrt{\\pi}
$$
`
};

/**
 * Detect browser locale fallback to 'en'. Called once at app start.
 */
export const c_getLocaleInitial = (): string => {
	if (typeof navigator === 'undefined') return 'en';
	const lang = navigator.language.split('-')[0].toLowerCase();
	return ['en', 'de', 'es', 'fr'].includes(lang) ? lang : 'en';
};
