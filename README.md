# MarkSafe 1.3.2

> **[mlcgo.eu](https://mlcgo.eu)** — tools, libraries and manuals · [Product page](https://mlcgo.eu/products/marksafe/)

A lightweight, high-performance Markdown Viewer and Editor for **Windows**, **macOS**, and **Linux**.
Designed for speed, security, and a polished user experience.

MarkSafe is more than just a viewer; it's a secure, native environment for reading and editing Markdown. Whether you are a developer, a technical writer, or a power user, MarkSafe provides a clean, distraction-free interface with industry-standard security features.

## ✨ Key Features

- **📝 Modern CodeMirror 6 Editor**: In-editor Markdown syntax highlighting, line numbers, bracket matching, and smooth synchronized scrolling.
- **📑 Multi-Tab Interface**: Work with multiple documents simultaneously.
- **✨ Rich Markdown Support**:
  - **GFM**: Tables, tasklists, and footnotes.
  - **YAML Front Matter**: Automatic parsing and structured metadata accordion (`--- ... ---`).
  - **GitHub Alerts**: High-visibility notes, tips, and warnings using `> [!NOTE]` syntax.
  - **Math**: Integrated KaTeX for complex mathematical expressions.
  - **Diagrams**: Native Mermaid.js support for flowcharts, sequences, Gantt charts, and architecture diagrams.
- **🌓 Synchronized Dark / Light Mode**: Unified theme system across editor, preview, syntax highlighting, and UI frames (with automatic OS preference detection).
- **🖨️ High-Contrast Print & PDF**: Dedicated print stylesheet (`@media print`) that removes UI chrome, sets clean page margins, prevents bad page breaks, and optimizes colors for paper.
- **📄 Export to HTML**: Export standalone HTML documents with embedded CSS and styles.
- **🌍 Multilingual**: Support for English, German, Spanish, and French.
- **⚡ Real-time Rendering**: Instant preview as you type.

## 🛡️ Security First

MarkSafe introduces a **Sandboxed Resource Model** to protect you from malicious markdown files:

1. **Directory Whitelisting**: The application can only access files in directories you have explicitly allowed. Opening a file via the dialog automatically whitelists its parent directory.
2. **External Resource Control**: Every time a document attempts to load images or content from a new domain, MarkSafe prompts for permission.
3. **Safe Link Interception**: External links are checked against your whitelist before being opened in your default browser.

## 📥 Installation

Download the latest version for your platform from the **[Releases](https://github.com/mlechner911/mdviewer/releases)** page or local dropzone.

### Windows
- Download `marksafe-1.3.1-windows-setup.exe` and run the installer.

### macOS
- Download `marksafe-mac.zip`, extract, and move `MarkSafe.app` to your `/Applications` folder.

### Linux
- Download `marksafe-1.3.1-linux-amd64`.
- Make it executable: `chmod +x marksafe-1.3.1-linux-amd64`.
- Run `./marksafe-1.3.1-linux-amd64` or add it to your `$PATH`.

## 📄 License

MIT License - Copyright (c) 2026 Michael Lechner

---
*For technical details and build instructions, see **[DEVELOPER.md](./DEVELOPER.md)** and **[THEMING.md](./THEMING.md)**.*
