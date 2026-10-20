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
| `ParseHeader`, `ParseFrontMatter` | Header lines (`**Dokumenten-ID:**`, `**Version:**`) and front matter |
| `EventKinds`, `PrefillFields`, `FrontMatterKeys` | The shared lists; change them here first |
| `HeadingSlug` | Heading anchors exactly like GitHub and mlcmd2pdf |
| `Analyze` | Everything an editor needs: findings, field and hint positions, lists |
| `Templates`, `NewDocument`, `SuggestRefNo` | New documents from the built-in templates (SOP, VA, FB, INFO) |
| `StoredText` | The text as the Sensor Monitor stores and renders it (upload front matter removed) |
| `mlcdoc.Preview` | Page images from the external `mlcdoc` binary (subpackage `mlcext/mlcdoc`) |

Severity: **error** means the Sensor Monitor rejects the document or mlcdoc
cannot render it; **warning** means it is accepted but probably not intended;
**info** marks intended behavior, such as a field filled from an event.

Tests: `go test ./...`. `MLCEXT_SEEDS=<dir> go test ./...` additionally
checks every `.md` file in `<dir>`, for example the Sensor Monitor's template
seeds, which must all be free of errors.
