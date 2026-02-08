# Theming Architecture

MD Viewer features a dual-theming system that separates the application's user interface from the rendered content.

## 1. App Theme (The Frame)

The **App Theme** controls the "Application Frame," which includes the Toolbar, Editor, and Resizer.

- **Options**: `Dark`, `Light`, `Auto` (System Preference).
- **Implementation**: Managed in `App.svelte` via Tailwind classes and a system media query listener.
- **Toggle**: A top-right SVG button cycles through these modes.

## 2. Preview Theme (The Content)

The **Preview Theme** controls the visual style of the rendered Markdown document.

- **Options**: Defined in `frontend/src/themes.ts`.
- **Source of Truth**: The `themes` array in `themes.ts` defines everything from background colors to the syntax highlighting style.
- **Properties**:
  - `chromaStyle`: The name of the highlighting style passed to the Go backend.
  - `containerClass`: Background and text colors for the preview pane.
  - `proseClass`: Tailwind Typography (`prose`) variants.
  - `mermaidTheme`: The base Mermaid diagram theme.
  - `mermaidVars`: Custom CSS variables for fine-tuning diagrams.

## Synchronized Rendering

When a Preview Theme is changed:
1.  The frontend notifies the Go backend of the new `chromaStyle`.
2.  The backend generates class-based HTML and returns the specific CSS for that style.
3.  The frontend injects the CSS and updates the `Preview` component classes.
4.  Mermaid diagrams are re-initialized with the theme-specific variables.

## Adding New Preview Themes

1.  Open `frontend/src/themes.ts`.
2.  Add a new entry to the `themes` array.
3.  Choose a matching Chroma style (e.g., `monokai`, `solarized-dark`).
4.  Define matching Tailwind and Mermaid variables.