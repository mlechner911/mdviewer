# marksafe_browse

Standalone CLI tool: serve a Markdown directory over HTTP and browse it
with an interactive Svelte 5 frontend (dark/light themes, recursive
tree, fuzzy search, validated links). Thin host shell around the shared
UI in `../frontend/src/shared/browse/` and the Go library
`../internal/mdbrowse/` — behavior changes belong there, not here.

## Quickstart

```bash
task dev        # build frontend + backend, serve ./example on :8312
task build      # binary -> bin/marksafe_browse
task release    # Linux + Windows binaries -> /mnt/data2tb/dropzone/
```

```bash
./bin/marksafe_browse --path ./docs --port 3000
./bin/marksafe_browse --path ./docs/index.md   # single file as root
```

Flags: `--path/-p` (required), `--port/-P` (default 8080),
`--bind/-b` (default 127.0.0.1), `--theme/-t` (dark|light|auto).

## HTTP API

| Endpoint | Description |
|---|---|
| `/` | App shell (empty `#app`; client fetches `index.md` itself) |
| `/tree` | TOC as JSON (`path/title/children/isDir`, lowercase) |
| `/render?path=&theme=` | `{title, html, path, size, modified}`; directories serve `index.md` or a virtual listing |
| `/api/search?q=` | Ranked fuzzy hits with `<mark>` snippets (max 20) |
| `/raw?path=` | Raw file with content type (images, PDFs, …) |
| `/md/<path>` | Raw markdown as text |
| `/assets/*`, `/favicon.ico` | Embedded frontend (content-hashed names, discovered at startup) |

Security: everything is confined to the root directory (`ResolvePath`
rejects escapes → 403, missing → 404); `..` in URLs is blocked.

## Structure

```text
cmd/marksafe_browse  -> Cobra CLI (thin)
internal/browse      -> HTTP glue only: routes, shell template, favicon
frontend/src         -> App shell (theme state) + index.html + shell CSS
../frontend/src/shared/browse -> the actual UI (shared with the Wails app)
../internal/mdbrowse -> the actual backend (shared with the Wails app)
```

Frontend changes: `task frontend` always rebuilds (no fingerprinting —
a skipped build embeds a stale bundle). The Go server discovers the
hashed `index-*.js/css` names from the embedded assets at startup.
