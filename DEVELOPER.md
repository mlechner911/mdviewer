# MarkSafe - Developer Documentation

Technical reference for building and extending MarkSafe.

## 🛠 Tech Stack

- **Backend**: Go 1.21+
  - [Wails v2](https://wails.io/) - Desktop framework.
  - [Goldmark](https://github.com/yuin/goldmark) - Extensible Markdown parser.
  - [Chroma](https://github.com/alecthomas/chroma) - Syntax highlighting.
  - [Bluemonday](https://github.com/microcosm-cc/bluemonday) - HTML sanitization.
- **Frontend**: Svelte 3 + TypeScript
  - [Tailwind CSS v3](https://tailwindcss.com/) - Utility-first CSS.
  - [KaTeX](https://katex.org/) - Math rendering.
  - [Mermaid.js](https://mermaid.js.org/) - Diagram rendering.

## 🏗 Project Structure

```text
/internal/markdown    -> Goldmark configuration & custom AST transformers.
/internal/config      -> JSON configuration & whitelist management.
/internal/filesystem  -> Safe file I/O wrappers.
/frontend/src/lib     -> Shared stores, constants, and backend bindings.
/frontend/src/themes  -> Theme definitions (base.json + presets).
/frontend/src/i18n.ts -> Translation dictionary and locale logic.
```

## ⚙️ Core Logic

### Markdown Rendering
The rendering pipeline is split between Go and Svelte:
1.  **Go**: Parses Markdown, applies `GitHubAlertTransformer`, highlights code with Chroma, and sanitizes the final HTML via `bluemonday`.
2.  **Svelte (`Preview.svelte`)**: Injects the HTML, scans for Mermaid diagrams and KaTeX formulas, and executes their respective client-side rendering engines.

### Security Whitelisting
All file and URL access is intercepted by `Preview.svelte`. It calls `backend.isPathAllowed` or `backend.isURLAllowed` before rendering resources. If a resource is blocked, a `security-request` event is dispatched to trigger the UI modal.

## 🎨 Extending MarkSafe

### Adding a Preview Theme
1.  Create a new JSON preset in `/frontend/src/themes/presets/`.
2.  Import it in `/frontend/src/themes.ts`.
3.  Add it to the `themes` array using the `createTheme` helper.

### Adding a Language
1.  Add the new locale code (e.g., `it`) to the `supported` array in `getInitialLocale()` in `i18n.ts`.
2.  Add the translations to the `translations` object in `i18n.ts`.
3.  Update the language submenu in `app.go` (`UpdateMenu` method).

## 🚀 Development

### Prerequisites
- Go 1.21+, Node.js 18+, Wails CLI.

| Task | Command |
|---------|-------------|
| Install Deps | `task install` |
| Dev Mode | `task dev` |
| Build | `task build` |

## 🤖 CI/CD

Multi-platform builds are automated via GitHub Actions (`.github/workflows/release.yml`). Releases are triggered by version tags (e.g., `v1.0.0`).

## 📄 License

MIT License - Copyright (c) 2026 Michael Lechner
