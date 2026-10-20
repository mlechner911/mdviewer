# mlcext — checker for the MLC Markdown Extension

> [!IMPORTANT]
> The MLC Markdown Extension is **not standard Markdown**. It is an
> MLC-specific extension for documents of the MLC Sensor Monitor (SOPs, forms)
> that the MLC PDF toolkit turns into fillable PDFs. See
> [docs/MLC-EXTENSIONS.md](../docs/MLC-EXTENSIONS.md).

A small, standalone Go module with its own `go.mod`. MarkSafe and the Sensor
Monitor both use it, so there is exactly one implementation of these rules.

```go
import "github.com/mlechner911/mdviewer/mlcext"

if mlcext.Detect(md) {
    for _, d := range mlcext.Check(md, mlcext.Options{RequireHeader: true}) {
        fmt.Printf("%d:%d %s %s: %s\n", d.Line, d.Col, d.Severity, d.Code, d.Message)
    }
}
```

| Function | What it does |
| --- | --- |
| `Check` | All findings, with line, column, severity, a stable code and a German message |
| `Detect` | Whether a document uses the extension (plain Markdown gets no findings) |
| `ExtractFields` | The form fields `[@…]`, ported from mlcmd2pdf `src/form.rs` with its tests |
| `DocumentMeta` | Id, version, title and kind: front matter (`doc_id`, `version`, `title`) first, the legacy header lines as a fallback |
| `ParseFrontMatter`, `ParseHeader` | Front matter, and the legacy header lines (`**Dokumenten-ID:**`, `**Version:**`) |
| `ExpandVariables`, `Variables` | `{{doc_id}}`, `{{version}}`, `{{title}}` replaced with the front matter values (not in code) |
| `SplitSignature` | The Sensor Monitor's signature in both forms (comment at the end, old block in front) and the signed content, byte for byte |
| `LayoutYAML` | Layout file for `mlcdoc md-to-pdf --layout`: title, page header with kind, id and version, watermark, organization and logo from the caller |
| `SuggestFileName` | `INFO-QM-001.md`, for exports with the version: `INFO-QM-001_v1.0.pdf` |
| `EventKinds`, `PrefillFields`, `FrontMatterKeys` | The shared lists; change them here first |
| `HeadingSlug` | Heading anchors exactly like GitHub and mlcmd2pdf |
| `Analyze` | Everything an editor needs: findings, field and hint positions, lists |
| `LoadTemplates`, `TemplateInfo.Render`, `SuggestRefNo` | Templates as data: schema (`*.json`, see `mlc-template.schema.json`) + Markdown body; built-in starting set, overridable from a folder |
| `StoredText` | The text as it is rendered: signature removed, variables replaced, upload front matter removed |
| `mlcdoc.New`, `Client.RenderPDF`, `Client.RenderPages` | PDF and page images from the external `mlcdoc` binary (subpackage `mlcext/mlcdoc`): version check, front matter preparation (`mlcdoc.Prepare`), typed `RenderError` with line |

Severity: **error** means the Sensor Monitor rejects the document or mlcdoc
cannot render it; **warning** means it is accepted but probably not intended;
**info** marks intended behavior, such as a field filled from an event.

Shared test cases: `testdata/fields.json` (form fields) and `testdata/meta.json`
(variables, signature, rendered text) are run by the TypeScript renderer
(`packages/mlc-markdown`) as well; regenerate with `go test -run
'TestFieldVectors|TestMetaVectors' -update` after a deliberate change.

Tests: `go test ./...`. `MLCEXT_SEEDS=<dir> go test ./...` additionally
checks every `.md` file in `<dir>`, for example the Sensor Monitor's template
seeds, which must all be free of errors.

## Templates

A template is a schema plus a body, both in one folder:

```json
{
  "$schema": "https://raw.githubusercontent.com/mlechner911/mdviewer/main/mlcext/mlc-template.schema.json",
  "name": "Formblatt CAPA",
  "body": "formblatt-capa.md",
  "category": "mlc",
  "file": "{{doc_id}}.md",
  "fields": [
    { "id": "title",       "label": "Titel",         "type": "text",  "required": true, "target": "frontmatter" },
    { "id": "bereich",     "label": "Bereich",       "type": "text",  "default": "QM" },
    { "id": "doc_id",      "label": "Dokumenten-ID", "type": "refno", "prefix": "FB", "area": "bereich", "required": true, "target": "frontmatter" },
    { "id": "version",     "label": "Version",       "type": "text",  "default": "1.0", "required": true, "target": "frontmatter" },
    { "id": "event_kinds", "label": "Gilt bei",      "type": "multiselect", "options": "event_kinds", "target": "frontmatter" }
  ]
}
```

Fields with `"target": "frontmatter"` go into the front matter under their
id; for MLC templates (`"category": "mlc"`) only the keys of
`FrontMatterKeys` are allowed there, and `doc_id`, `version`, `title` must
be such fields. The body uses `{{doc_id}}`, `{{version}}`, `{{title}}`:
document variables that stay in the file (`# {{title}}`). Other
placeholders (here `{{bereich}}`) are replaced when the document is created.
A template in the old form (`kennung`, `titel`, header line in the text)
still renders and gets a warning in `TemplateInfo.Warnings`.

`SeedTemplates(dir)` copies the built-in templates (MLC and general:
README, CHANGELOG, empty document, minutes) into a user folder once, keeping
the user's changes on later calls; load it with `LoadTemplatesFrom(false,
dir, …)`.

Field types: `text`, `multiline`, `number` (`min`, `max`, `step`), `date`
(`format` TT.MM.JJJJ or JJJJ-MM-TT), `time` (also `hour`), `select`,
`multiselect`, `refno`, `docref` (other documents, filled with
`TemplateInfo.WithDocuments`). Placeholders in the body and in `file`:
`{{id}}`, `{{id|yaml}}` (YAML-safe), `{{id|list}}` (YAML list),
`{{id|ref}}` (docref: "ID (Titel)"). `LoadTemplates`
reports problems per template (unknown type, missing body, placeholder
without a field …); such a template cannot be rendered.
