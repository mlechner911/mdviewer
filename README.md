# MarkSafe

A lightweight, high-performance Markdown Viewer and Editor for **Windows**, **macOS**, and **Linux**.
Designed for speed, security, and a polished user experience.

MarkSafe is more than just a viewer; it's a secure, native environment for reading and editing Markdown. Whether you are a developer, a technical writer, or a power user, MarkSafe provides a clean, distraction-free interface with industry-standard security features.

## ✨ Key Features


- ** Multi-Tab Interface**: Work with multiple documents simultaneously.

- ** Rich Markdown Support**:
  - **GFM**: Tables, tasklists, and footnotes.
  - **GitHub Alerts**: High-visibility notes, tips, and warnings using `> [!NOTE]` syntax.
  - **Math**: Integrated KaTeX for complex mathematical expressions.
  - **Diagrams**: Native Mermaid.js support for flowcharts, sequences, and more.



- **📄 Export & Print**: Export to standalone HTML (with embedded CSS) or print to PDF.
- **🌍 Multilingual**: Support for English, German, Spanish, and French.
- ** Real-time Rendering**: Instant preview as you type.

##  Security First

MarkSafe introduces a **Sandboxed Resource Model** to protect you from malicious markdown files:

1.  **Directory Whitelisting**: The application can only access files in directories you have explicitly allowed. Opening a file via the dialog automatically whitelists its parent directory.
2.  **External Resource Control**: Every time a document attempts to load images or content from a new domain, MarkSafe prompts for permission.
3.  **Safe Link Interception**: External links are checked against your whitelist before being opened in your default browser.

## 📥 Installation

Download the latest version for your platform from the **[Releases](https://github.com/mlechner911/mdviewer/releases)** page.

### Windows
- Download `marksafe-installer.exe` and run it to install.

### macOS
- Download `marksafe-mac.zip`, extract, and move `MarkSafe.app` to your `/Applications` folder.

### Linux
- Download the `marksafe` binary.
- Make it executable: `chmod +x marksafe`.
- Run it or add it to your `$PATH`.

## 📄 License

MIT License - Copyright (c) 2026 Michael Lechner

---
*For technical details and build instructions, see **[DEVELOPER.md](./DEVELOPER.md)**.*
