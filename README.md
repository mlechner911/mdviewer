# MarkSafe 1.8.0

> **[mlcgo.eu](https://mlcgo.eu)** — tools, libraries and manuals · [Product page](https://mlcgo.eu/products/marksafe/)

A lightweight, high-performance Markdown Viewer and Editor for **Windows**, **macOS**, and **Linux**.
Designed for speed, security, and a polished user experience.

MarkSafe is more than just a viewer; it's a secure, native environment for reading and editing Markdown. Whether you are a developer, a technical writer, or a power user, MarkSafe provides a clean, distraction-free interface with industry-standard security features.

##  Key Features

- **  Modern CodeMirror 6 Editor**: In-editor Markdown syntax highlighting, line numbers, bracket matching, and smooth synchronized scrolling.
- **  Multi-Tab Interface**: Work with multiple documents simultaneously.
- ** Rich Markdown Support**:
  - **GFM**: Tables, tasklists, and footnotes.
  - **YAML Front Matter**: Automatic parsing and structured metadata accordion (`--- ... ---`).
  - **GitHub Alerts**: High-visibility notes, tips, and warnings using `> [!NOTE]` syntax.
  - **Math**: Integrated KaTeX for complex mathematical expressions.
  - **Diagrams**: Native Mermaid.js support for flowcharts, sequences, Gantt charts, and architecture diagrams.
  - **MLC Markdown Extension** *(non-standard)*: page breaks and anchor targets for MLC Sensor Monitor documents (SOPs, forms). Not part of any Markdown standard — see **[docs/MLC-EXTENSIONS.md](./docs/MLC-EXTENSIONS.md)** for what is standard, what is a common extension and what is MLC-specific.
- ** Synchronized Dark / Light Mode**: Unified theme system across editor, preview, syntax highlighting, and UI frames (with automatic OS preference detection).
- ** High-Contrast Print & PDF**: Dedicated print stylesheet (`@media print`) that removes UI chrome, sets clean page margins, prevents bad page breaks, and optimizes colors for paper.
- ** Export to HTML**: Export standalone HTML documents with embedded CSS and styles.
** Multilingual**: Support for English, German, Spanish, and French.
- ** Real-time Rendering**: Instant preview as you type.
- ** MLC Sensor Monitor** *(optional module, not in the public builds)*: open SOPs and forms from the Sensor Monitor, save changes as drafts and see whether a file matches the approved version — see **[docs/SENSOR-MONITOR.md](./docs/SENSOR-MONITOR.md)**.
- ** Browser Mode**: Open the current document's directory in a standalone window — recursive tree, fuzzy full-text search (`Ctrl+K`), validated intra-doc links, directory indexes, and the same theme as the main window.

## 🛡️ Security First

MarkSafe introduces a **Sandboxed Resource Model** to protect you from malicious markdown files:

1. **Directory Whitelisting**: The application can only access files in directories you have explicitly allowed. Opening a file via the dialog automatically whitelists its parent directory.
2. **External Resource Control**: Every time a document attempts to load images or content from a new domain, MarkSafe prompts for permission.
3. **Safe Link Interception**: External links are checked against your whitelist before being opened in your default browser.

##  Installation

Download the latest version for your platform from the **[Releases](https://github.com/mlechner911/mdviewer/releases)** page or directly from the official **[mlcgo.eu Download Page](https://mlcgo.eu/products/marksafe/en/download/)**.

> [!IMPORTANT]
> **macOS Note:** Officially code-signed and Apple-notarized packages (`.dmg` for Apple Silicon + Intel) are hosted exclusively on **[mlcgo.eu](https://mlcgo.eu/products/marksafe/en/download/)**.
> GitHub Releases provide standard unsigned/ad-hoc builds. For a seamless macOS experience without Gatekeeper warnings, please download the notarized disk image from **[mlcgo.eu](https://mlcgo.eu/downloads/marksafe/marksafe-macos.dmg)**.

### macOS
- **Recommended:** Download the signed & notarized Universal DMG from **[mlcgo.eu](https://mlcgo.eu/downloads/marksafe/marksafe-macos.dmg)**, open it, and drag `MarkSafe.app` into `/Applications`.
- *Alternative:* Download `marksafe-macos-arm64.zip` from GitHub Releases.

### Windows
- Download `marksafe-windows-setup.exe` (or `marksafe-1.8.0-windows-setup.exe`) and run the installer.
- A standalone portable executable (`marksafe-windows.exe`) is also available on **[mlcgo.eu](https://mlcgo.eu/downloads/marksafe/marksafe-windows.exe)**.

### Linux
- Download `marksafe-linux` (or `marksafe-1.8.0-linux-amd64`).
- Make it executable: `chmod +x marksafe-1.8.0-linux-amd64`.
- Run `./marksafe-1.8.0-linux-amd64` or add it to your `$PATH`.

##  License

MIT License - Copyright (c) 2026 Michael Lechner

Exceptions:

- **Sensor Monitor module** (`sensormonitor/`, `sensormonitor_on.go`, `frontend/src/sensormonitor/`): [Elastic License 2.0](./sensormonitor/LICENSE). It is built in only with `SENSOR=1`. Public builds contain none of it and are MIT only.
- **MLC PDF toolkit `mlcdoc`**, shipped with the installers: its own license files, installed next to it (`licenses/mlcdoc`).

---
*For technical details and build instructions, see **[DEVELOPER.md](./DEVELOPER.md)** and **[THEMING.md](./THEMING.md)**. Changes per release: **[CHANGELOG.md](./CHANGELOG.md)**. The standalone directory browser: **[marksafe_browse](./marksafe_browse/README.md)**.*

<!-- mlcai-private -->
## Project documentation (`.mlcai/`)

`.mlcai/` is a **private git submodule**: internal planning, backlog and work notes, maintained with the MLC Doc Hub. It is not publicly accessible — clone **without** `--recurse-submodules`; the build does not need it. Links into `.mlcai/` only work with access (`git submodule update --init .mlcai`).

## Who is "Claude" in the commits?

Some commits in this repository are co-authored by Claude, Anthropic's AI
model. It helps write code, keeps our documentation and backlog up to date and
digs through failing builds — every change is reviewed before it is merged.
We don't hide it: [how we work with Claude](https://mlcgo.eu/ai/en.html).
