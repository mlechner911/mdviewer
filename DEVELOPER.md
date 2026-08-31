# MarkSafe - Developer Documentation

Technical reference for building and extending MarkSafe.

## 🛠 Tech Stack

- **Backend**: Go 1.23+
  - [Wails v2](https://wails.io/) - Desktop framework.
  - [Goldmark](https://github.com/yuin/goldmark) - Extensible Markdown parser.
  - [goldmark-meta](https://github.com/yuin/goldmark-meta) - YAML Front Matter parser.
  - [Chroma](https://github.com/alecthomas/chroma) - Syntax highlighting.
  - [Bluemonday](https://github.com/microcosm-cc/bluemonday) - HTML sanitization.
- **Frontend**: Svelte 5 (Runes) + TypeScript + Vite 8
  - [CodeMirror 6](https://codemirror.net/) - Modular markdown editor with syntax highlighting and line numbers.
  - [Tailwind CSS v3](https://tailwindcss.com/) - Utility-first CSS (`@tailwindcss/typography`).
  - [KaTeX](https://katex.org/) - Mathematical formula rendering.
  - [Mermaid.js](https://mermaid.js.org/) - Diagram and chart rendering.

## 🏗 Project Structure

```text
/internal/markdown    -> Goldmark configuration, AST transformers & frontmatter parsing.
/internal/config      -> JSON configuration & whitelist management.
/internal/filesystem  -> Safe file I/O wrappers.
/frontend/src/components -> Svelte 5 components (Editor, Preview, TabsBar, Toolbar, etc.).
/frontend/src/lib     -> Shared stores, constants, and backend bindings.
/frontend/src/themes  -> Unified theme definitions (base.json + dark/light presets).
/frontend/src/i18n.ts -> Translation dictionary and locale logic.
```

## ⚙️ Core Logic

### Markdown Rendering
The rendering pipeline is split between Go and Svelte:
1. **Go (`internal/markdown/markdown.go`)**: Parses Markdown, extracts Front Matter via `goldmark-meta`, applies `GitHubAlertTransformer`, highlights code blocks with Chroma (`github-dark` / `github`), and sanitizes output via `bluemonday`.
2. **Svelte (`Preview.svelte`)**: Injects HTML, renders Front Matter metadata box, and executes client-side Mermaid diagrams and KaTeX formulas.

### CodeMirror 6 Editor
`components/Editor.svelte` wraps CodeMirror 6 using Svelte 5 Runes. Themes are dynamically reconfigured via CodeMirror `Compartment` without unmounting or losing document/cursor state.

### Security Whitelisting
All file and URL access is intercepted by `Preview.svelte`. It calls `backend.isPathAllowed` or `backend.isURLAllowed` before rendering resources. If a resource is blocked, a `security-request` event is dispatched to trigger the UI modal.

## 🚀 Development & Build

### Prerequisites
- Go 1.23+, Node.js 20+, Wails CLI v2.12+, NSIS (`makensis`).

| Task | Command | Description |
|---|---|---|
| Install Deps | `task install` | Install Go and NPM dependencies |
| Dev Mode | `task dev` | Run Wails dev with live reload |
| Build Linux | `task build` | Compile Linux AMD64 binary |
| Build Windows | `task build:windows` | Compile Windows binary & NSIS installer |
| Build All | `task build:all` | Build both platforms |
| Release | `task release` | Build and copy installers to `/mnt/data2tb/dropzone/` |
| Check | `task check` | Run `svelte-check` and `go test` |

## 📄 License

MIT License - Copyright (c) 2026 Michael Lechner
