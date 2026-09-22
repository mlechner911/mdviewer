# Changelog

All notable changes to MarkSafe. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.6.1] – 2026-09-22

### Added
- **Browser mode**: open the active document's directory in a standalone window (menu *View → Browser-Modus*, `Ctrl/Cmd+Shift+B`, or the new toolbar button). Recursive directory tree, fuzzy full-text search (`Ctrl/Cmd+K`) with snippets, validated document links, and automatic directory indexes.
- Shared UI library (`frontend/src/shared/browse`) powering both the in-app browser mode and the standalone `marksafe_browse` tool: sidebar, search modal, theme toggle, info panel, external-link confirmation.
- Shared Go library (`internal/mdbrowse`): rooted file access with hard boundary, TOC scanning, title extraction, goldmark render pipeline, native fuzzy search.
- Document info panel (toolbar): title, relative path tooltip in the tree, file size, modification date, word/character counts, reading time.
- Adjustable line density (compact/comfortable/spacious) covering content and tree; resizable, collapsible sidebar persisted per profile.
- Window title of the browse window shows the current root directory.

### Fixed
- Theme toggle fully re-renders preview (Mermaid diagrams no longer flash red syntax-error boxes).
- Search no longer misses short queries: fixed slice-out-of-range panics and rune-misaligned highlighting in snippets (e.g. `roll` vs `Rollout`).
- Stale frontend bundles: the build fingerprint now covers the shared UI, and shipped assets use content hashes.
- Sidebar auto-expands and scrolls to the document opened via link or history.
- Native language menu marks the active language; editor hidden-state persists; About dialog follows the app theme (native dialog removed).

## [1.6.0] – 2026-09-18

### Added
- `marksafe_browse`: standalone CLI tool serving a Markdown directory over HTTP (`task dev`, binary in `bin/`), with interactive Svelte 5 frontend (dark/light/system themes, collapsible tree, search, validated links, directory fallback pages).
- Developer look & feel: Inter/JetBrains Mono typography, monochrome IDE palette, SVG icon set (no emoji), sticky toolbar, SEO meta tags.
- Link validation: missing/out-of-root targets render as non-clickable placeholders instead of 404s; images served via `/raw`.
- Forgiving parser: repairs glued headings/rules from pasted sources, truncates runaway titles.

### Fixed
- Svelte hydration errors (`mount` on server context), missing favicon route, lowercase TOC JSON contract.
- Intermittent blank pages: sequenced loader with stale-response dropping, abort handling, visible retry panel, per-document remount.
- `svelte-check` clean (0/0) in both frontends; Go vet clean.

## [1.5.1] and earlier

See [GitHub Releases](https://github.com/mlechner911/mdviewer/releases) for earlier versions.
