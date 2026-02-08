# MD Viewer

A modern, high-performance Markdown Viewer and Editor built with **Wails**, **Go**, and **Svelte**.

## 🚀 Features

- **Real-time Rendering**: Instant preview as you type.
- **GFM Support**: GitHub Flavored Markdown (Tables, Checklists, Emojis).
- **Syntax Highlighting**: Beautiful code blocks powered by Chroma.
- **XSS Protection**: Secure HTML sanitization with bluemonday.
- **Native Dialogs**: System-native file open and save functionality.
- **Split View**: Dual-pane layout with a focus-friendly editor and a professional typography preview.

## 🛠 Tech Stack

- **Backend**: Go 1.23+
  - [Wails v2](https://wails.io/) (Desktop App Framework)
  - [Goldmark](https://github.com/yuin/goldmark) (Markdown Parser)
  - [Bluemonday](https://github.com/microcosm-cc/bluemonday) (HTML Sanitizer)
- **Frontend**: Svelte 3 + TypeScript
  - [Tailwind CSS v3](https://tailwindcss.com/)
  - [@tailwindcss/typography](https://tailwind-typography.netlify.app/)

## 👨‍💻 Development

### Prerequisites

- Go 1.23+
- Node.js & npm
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- `xvfb` (for remote/headless development)

### Task Runner

The project uses a `Taskfile.yml` for common operations.

| Command | Description |
|---------|-------------|
| `task install` | Installs Go and NPM dependencies. |
| `task dev` | Runs the app in development mode (wrapped in `xvfb-run` for remote work). |
| `task build` | Compiles a production-ready binary. |
| `task tidy` | Cleans up Go modules. |

### Remote Development Note

Since this project is often developed over remote SSH sessions, the `task dev` command is configured to use `xvfb-run`. This allows the Wails development server to start even without a physical display attached, which is necessary for the initial frontend compilation and binding generation.

## 🏗 Project Structure

- `/app.go`: Main application logic and Go-to-Frontend bindings.
- `/main.go`: Entry point and Wails configuration.
- `/frontend/src/App.svelte`: Main UI implementation.
- `/frontend/tailwind.config.js`: Tailwind and Typography configuration.