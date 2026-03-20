# MD Viewer

A modern, high-performance Markdown Viewer and Editor for **Windows**, **macOS**, and **Linux**.
Designed for speed, simplicity, and a polished user experience.

MD Viewer is a quick, native tool for previewing and editing Markdown files. Whether you are a developer, a technical writer, or someone who just needs to read a README, MD Viewer provides a clean, distraction-free environment.

## 🚀 Key Features

- **Cross-Platform**: Consistent experience across Windows, macOS, and Linux.
- **Multi-File Support**: Work with multiple documents simultaneously using a clean tabbed interface.
- **Real-time Rendering**: See your changes instantly as you type.
- **Comprehensive Markdown Support**:
  - **GFM**: Tables and interactive tasklists.
  - **GitHub-style Alerts**: High-visibility notes, tips, and warnings.
  - **Math**: Support for KaTeX mathematical expressions.
  - **Diagrams**: Integrated Mermaid.js support for flowcharts and diagrams.
- **Dual-Theming**:
  - **App Frame**: Choose between Dark, Light, or Auto modes for the editor.
  - **Preview**: High-quality preview styles including Dark, Light, Sepia, and Monochrome.
- **Focus Mode**: Hide everything except the content for a true distraction-free experience.
- **Print & Export**: One-click "Print to PDF" or export to standalone HTML.
- **Multilingual (i18n)**: Native support for English, German, Spanish, and French.

## 🛡️ Secured by Default

MD Viewer is built with security in mind. By default, it operates in a restricted environment:
- **Directory Whitelisting**: The application only accesses files in directories you have explicitly allowed. Opening or saving a file automatically whitelists its location.
- **External Resource Control**: Every time a document tries to load images or content from a new external website, you will be prompted for permission.
- **Safe Link Handling**: Links to other Markdown files are intercepted and opened within the viewer after a security check.

## 📥 Installation

The easiest way to get MD Viewer is to download the latest version from the **[Releases](https://github.com/mlechner911/mdviewer/releases)** page.

### Windows
1. Download `md-viewer-installer.exe`.
2. Run the installer to set up MD Viewer on your system.

### macOS
1. Download `md-viewer-mac.zip`.
2. Extract the ZIP and move `MD Viewer.app` to your `Applications` folder.

### Linux
1. Download the `md-viewer` binary.
2. Grant execution permissions: `chmod +x md-viewer`.
3. Run it directly or add it to your path.

## 📄 License

MIT License - Copyright (c) 2026 Michael Lechner

---
*For technical details, build instructions, and contribution guidelines, see **[DEVELOPER.md](./DEVELOPER.md)**.*
