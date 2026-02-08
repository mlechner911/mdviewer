# MD Viewer

A modern, high-performance Markdown Viewer and Editor built with **Wails**, **Go**, and **Svelte**.

Note: there are other really good markdown viewers available with some enhanced features like pdf outrput etc. This is not the idea behind this viewer. I want a quick tool - where i can
preview an md file or drop it on an icon  to see the preview.

If you neeed pdf - you can export to html and print to pdf.



## 🚀 Features

- **Real-time Rendering**: Instant preview as you type with debounced updates for smooth performance.
- **Advanced Markdown**:
  - GFM Support (Tables, Tasklists).
  - GitHub-style Alerts (`> [!NOTE]`, `[!TIP]`, etc.).
  - Mathematical Expressions (KaTeX integration).
  - Diagrams (Mermaid.js support with theme-aware styling).
  - Emojis (Shortcode support like `:rocket:`).
- **Dual-Theming System**:
  - **App Frame**: Independent Light/Dark/Auto modes for the editor and toolbar.
  - **Preview**: Customizable styles including Dark, Light, Sepia, and Monochrome.
- **Professional UI**:
  - Resizable split-pane layout.
  - Icon-based toolbar with internationalized tooltips.
  - Integrated Zoom/Font-size control for the preview pane.
- **Internationalization (i18n)**: Support for English, German, Spanish, and French.
- **Native Integration**: System-native file dialogs and external link handling.

## 🛠 Tech Stack

- **Backend**: Go 1.23+
  - [Wails v2](https://wails.io/)
  - [Goldmark](https://github.com/yuin/goldmark) (with Emoji, Math, and custom Alert extensions)
  - [Chroma](https://github.com/alecthomas/chroma) (Class-based syntax highlighting)
- **Frontend**: Svelte 3 + TypeScript
  - [Tailwind CSS v3](https://tailwindcss.com/)
  - [KaTeX](https://katex.org/)
  - [Mermaid.js](https://mermaid.js.org/)

## 👨‍💻 Development

### Task Runner

| Command | Description |
|---------|-------------|
| `task install` | Installs Go and NPM dependencies. |
| `task dev` | Runs the app in development mode (remote-ready with `xvfb-run`). |
| `task build` | Compiles a production-ready binary for the current platform. |

## 🏗 Project Structure

- `/internal/markdown`: Core rendering logic and Goldmark configuration.
- `/internal/filesystem`: Native file I/O utilities.
- `/frontend/src/components/Preview.svelte`: Complex rendering logic (KaTeX/Mermaid).
- `/frontend/src/themes.ts`: Centralized theme definitions.
- `/frontend/src/i18n.ts`: Translation dictionary and logic.

## 📄 License

MIT License - Copyright (c) 2026 Michael Lechner
