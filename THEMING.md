# Theming & Print Architecture

MarkSafe features a streamlined, fully synchronized theming architecture focused on **Dark Mode** and **Light Mode** (with automatic System default support), along with a dedicated high-contrast **Print Stylesheet**.

## 1. Unified Dark / Light System

Instead of disjointed UI vs. Preview theme selections, the application theme (`appTheme`) serves as the single source of truth:
- **`dark`**: Dark application frame, editor pane, markdown preview, `github-dark` syntax highlighting, and dark Mermaid diagrams.
- **`light`**: Crisp white/light frame, editor pane, markdown preview, `github` syntax highlighting, and light Mermaid diagrams.
- **`auto`**: Dynamically follows the OS system appearance preference (`prefers-color-scheme`).

### Theme Definitions
Themes are located in `frontend/src/themes/`:
- `base.json`: Base configuration shared by themes (base CSS, mermaid markers, error styling).
- `presets/dark.json`: Dark mode styling tokens and variables.
- `presets/light.json`: Light mode styling tokens and variables.

Theme merging is handled in `frontend/src/themes.ts` via `getTheme('dark' | 'light')`.

### Document stylesheet (single source)
Everything rendered from Markdown is styled by one file,
`frontend/src/shared/markdown/document.css`, used by the editor preview,
the browse view (app window and `marksafe_browse`) and the HTML export:
- Put `class="md-doc"` on the element holding the rendered HTML. Light is
  the default; `data-theme="dark"` on it or any ancestor switches to dark.
- Alert headings follow the nearest `lang` attribute (`de` → Hinweis, Tipp,
  Wichtig, Warnung, Vorsicht), like the mlcdoc PDF toolkit.
- Hosts keep only their chrome (sidebar, toolbar, scroll container).
- `<div class="page-break"></div>` shows as a marker on screen and breaks
  the page in print.

## 2. Dynamic Syntax Highlighting & Diagram Sync

When the active theme transitions:
1. `$effectiveAppTheme` resolves `'auto'` to the active `'dark'` or `'light'` mode.
2. The Chroma syntax highlighting CSS is dynamically retrieved from the backend renderer (`GetStyleCSS`) and applied.
3. Mermaid initializes with matching theme parameters (`'dark'` vs `'default'`) and re-renders SVG diagrams.
4. Tailwind and custom CSS classes update seamlessly across the entire interface.

## 3. High-Contrast Print & PDF Output

MarkSafe includes comprehensive `@media print` styling:
- **UI Chrome Removal**: Toolbar, tabs, editor, status bar, modals, and toasts are automatically excluded from print.
- **Paper Optimization**: Sets clean page margins (`@page`), pure white backgrounds, crisp dark typography, and proper line heights.
- **Page-Break Protection**:
  - Headings (`h1`-`h6`) enforce `break-after: avoid;` to prevent orphan headers.
  - Code blocks (`pre`), tables, blockquotes, GitHub alerts, images, and Mermaid diagrams enforce `break-inside: avoid;` to prevent awkward splits across pages.
  - Code blocks wrap long lines gracefully (`white-space: pre-wrap; word-break: break-word;`).
- **High-Contrast Syntax Highlighting**: Overrides syntax tokens in `@media print` for maximum legibility on physical paper and PDF exports.
