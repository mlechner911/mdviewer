# Theming Architecture

This document describes how themes are managed in the MD Viewer.

## Centralized Configuration: `themes.ts`

The source of truth for all themes is located in `frontend/src/themes.ts`. This file defines the `Theme` interface and the available theme presets.

### Theme Properties
- `id`: Unique identifier.
- `chromaStyle`: The syntax highlighting style used by the Go backend (Chroma).
- `containerClass`: Tailwind classes for the preview background and base text color.
- `proseClass`: Tailwind Typography (`@tailwindcss/typography`) classes to style the rendered HTML.
- `mermaidTheme`: The built-in Mermaid theme (`dark`, `default`, etc.).
- `mermaidVars`: Custom CSS variables to fine-tune Mermaid diagram colors.

## How it works

1.  **Selection**: When a user selects a theme in the Toolbar, the `currentTheme` object in `App.svelte` is updated.
2.  **Backend Sync**: The `chromaStyle` property is sent to the Go backend during the Markdown conversion. This ensures that the inline styles for syntax highlighting match the frontend aesthetic.
3.  **Frontend Rendering**: The `Preview.svelte` component receives the `Theme` object and applies the CSS classes and Mermaid variables dynamically.

## Current Scope & Limitations

### 1. Component Isolation (Current State)
Currently, **only the Preview component** is affected by the theme selection. 
- The **Toolbar** and **Editor** are hardcoded to a "Dark" (Slate-based) look to maintain a consistent "Application Frame" feel.
- The **Editor** (`textarea`) does not currently support syntax highlighting itself; it is a plain mono-spaced input.

### 2. Adding a New Theme
To add a new theme (e.g., "Solarized"):
1.  Open `frontend/src/themes.ts`.
2.  Add a new entry to the `themes` array.
3.  Choose a matching Chroma style (see [Chroma Gallery](https://xyproto.github.io/splash/)).
4.  Define the Tailwind classes for the container and prose.

## Future Improvements
- **Global Theming**: Extend the `Theme` interface to include `toolbarClass` and `editorClass` to allow the entire application to change its skin.
- **Custom Chroma Styles**: Support for loading external `.xml` or `.json` Chroma style definitions in the Go backend.
- **Editor Highlighting**: Integration of a full editor like Monaco or CodeMirror that can also sync with the selected theme.
