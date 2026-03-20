# MD Viewer - Developer Documentation

This document provides technical information for developers who wish to build, modify, or contribute to MD Viewer.

## 🛠 Tech Stack

- **Backend**: Go 1.21+
  - [Wails v2](https://wails.io/) - Native desktop application framework.
  - [Goldmark](https://github.com/yuin/goldmark) - Markdown parser with Emoji, Math, and custom Alert extensions.
  - [Chroma](https://github.com/alecthomas/chroma) - Syntax highlighting.
- **Frontend**: Svelte 3 + TypeScript
  - [Tailwind CSS v3](https://tailwindcss.com/) - Styling.
  - [KaTeX](https://katex.org/) - Mathematical expressions.
  - [Mermaid.js](https://mermaid.js.org/) - Diagram rendering.

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- (Optional) [Task](https://taskfile.dev/) task runner.

### Development Commands

| Command | Description |
|---------|-------------|
| `task install` | Installs Go and NPM dependencies. |
| `task dev` | Runs the app in development mode. |
| `task build` | Compiles a production-ready binary for the current platform. |

## 🏗 Project Structure

- `app.go`: Main Wails application logic and backend bindings.
- `/internal/markdown`: Core rendering logic and Goldmark configuration.
- `/internal/filesystem`: Native file I/O utilities.
- `/frontend/src/App.svelte`: Main application UI and state management.
- `/frontend/src/components/Preview.svelte`: Complex rendering logic (KaTeX/Mermaid).
- `/frontend/src/themes.ts`: Centralized theme definitions.
- `/frontend/src/i18n.ts`: Translation dictionary and logic.

## ⚙️ Configuration & Security

MD Viewer persists user settings and security whitelists in a `config.json` file. This file is stored in the system's standard application configuration directory:

- **Linux**: `~/.config/md-viewer/config.json`
- **macOS**: `~/Library/Application Support/md-viewer/config.json`
- **Windows**: `%AppData%\md-viewer\config.json`

### Structure
```json
{
  "whitelisted_paths": [
    "/absolute/path/to/directory"
  ],
  "whitelisted_urls": [
    "example.com"
  ]
}
```

## 🤖 CI/CD

The project uses GitHub Actions for multi-platform builds. The workflow is defined in `.github/workflows/release.yml`. It automatically generates:
- **Windows**: NSIS Installer (AMD64).
- **macOS**: Universal Binary (packaged as ZIP).
- **Linux**: Standalone Binary (AMD64).

Releases are triggered by pushing a tag (e.g., `v0.5.5`).

## 📄 License

MIT License - Copyright (c) 2026 Michael Lechner
