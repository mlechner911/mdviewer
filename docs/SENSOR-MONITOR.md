# Connection to the MLC Sensor Monitor

> [!NOTE]
> This only matters if your organisation uses the **MLC Sensor Monitor**.
> The connection is an optional module: the public MarkSafe downloads do
> not contain it. Your organisation gets a separate build with it (see
> [Building with the module](#building-with-the-module)). Even in that
> build, MarkSafe shows nothing of it until its address is set.

The Sensor Monitor keeps the SOPs and forms of an organisation, with
versions, approval and review dates. MarkSafe can open them, save changes
back as a **draft** and show whether the file in front of you is the
approved version.

MarkSafe cannot approve a document, set its status or "valid from", or
delete it. Those steps stay in the Sensor Monitor, which enforces the
four-eyes principle and the signing PIN.

## Turning it on

1. Open **File › Settings…** (`Ctrl+,`, on macOS `⌘,`).
2. Under **Connections**, enter the address of your Sensor Monitor, for
   example `mlcmonit.example.com`. It must use https. Plain http is accepted
   only for `localhost`, because the sign-in token goes with every request.
3. **File › Sensor Monitor** now appears in the menu.

To turn it off, clear the address. This also ends the sign-in.

## Signing in

**File › Sensor Monitor › Sign In…** opens the sign-in page of your
organisation in the browser and shows a code. Check that the browser shows
the same code and confirm it there.

- The sign-in lasts one hour and cannot be renewed; then sign in again.
- MarkSafe keeps the token only in memory, never on disk.
- While you are signed in, the status bar shows "Sensor Monitor · signed in
  until …".
- **Open…** and **Save as Draft…** ask you to sign in first if needed and
  then go on.

## Opening a document

**File › Sensor Monitor › Open…** asks for the document ID (`doc_id`, for
example `SOP-QM-001`). It shows:

- the approved version, with its version number and when the review is due
- a draft waiting for approval, if there is one

**Open** loads that version into a new tab. The front matter (`doc_id`,
`version`, `title`, `description`, `event_kinds`) comes from the data in
the Sensor Monitor. The signature block is left out, because a signature
belongs to the stored version, not to a working copy.

## Saving as a draft

**File › Sensor Monitor › Save as Draft…** sends the document to the Sensor
Monitor as a draft. The document needs a `doc_id` in its front matter.

- If an approved version with that ID exists, the draft becomes its next
  version. Otherwise a new document is created.
- **Reason for the change** goes into the change log. It is required if
  your organisation says so.
- The field for the **signing PIN** appears only when the Sensor Monitor
  asks for it.
- You need the rights D and K in the Sensor Monitor.

These are the usual refusals:

| Message | What to do |
| --- | --- |
| A draft is already waiting for this ID | Approve or discard that draft in the Sensor Monitor first. |
| The content matches the approved version | Nothing to save. |
| The Sensor Monitor rejects the document | Fix the errors the check reports. MarkSafe runs the same check in the status bar. |
| Your organisation requires a reason / the PIN | Fill in the field and save again. |

## Comparing with the approved version

While you are signed in, MarkSafe compares a document that has a `doc_id`
with the Sensor Monitor. It does so shortly after each change, and nothing
is saved. The status bar shows the result:

- **matches the approved version 1.0** (green)
- **matches the waiting draft 1.1**
- **differs from the approved version 1.0**
- **not yet in the Sensor Monitor**

The comparison ignores the front matter, the signature and line endings,
just as the Sensor Monitor does when it stores a document.

## Watermark when printing

If the document does not match the approved version, **printing** and
**PDF export** ask whether to add the watermark **ENTWURF** (draft). The
answer applies to that one PDF only.

MarkSafe asks only when you have not set a watermark of your own; your
setting always wins. Nothing is enforced: MarkSafe is a general Markdown
editor and only makes the suggestion.

## For administrators

- The API is described in the Sensor Monitor's documentation (`MARKSAFE.md`,
  section "Schnittstelle für MarkSafe"). MarkSafe uses `/mlcmonit/api/v1`
  and the device login `/mlcauth/device/…` on the same host.
- Personal API tokens (`mlc_pat_…`) are not used yet.

## Building with the module

The module lives in its own folders and is compiled only on request:

| Part | Where |
| --- | --- |
| Go service (bindings, settings file `sensormonitor.json`) | `sensormonitor/` |
| API client and device login | `sensormonitor/api/` |
| Hook into MarkSafe (build tag `sensormonitor`) | `sensormonitor_on.go`, `extensions.go` |
| Frontend (dialogs, comparison, texts) | `frontend/src/sensormonitor/` |
| Empty stand-in for the standard build | `frontend/src/sensormonitor-off/` |

```sh
task build:windows SENSOR=1        # Windows with the module
task build SENSOR=1                # Linux
task build:macos:remote SENSOR=1   # macOS on the Mac Studio
task release SENSOR=1              # all, named marksafe-<version>-sensormonitor-…
```

Without `SENSOR=1` neither the Go code nor the frontend code of the module
is in the build.

**License:** the module is under the [Elastic License 2.0](../sensormonitor/LICENSE),
the rest of MarkSafe under MIT. Every file of the module carries
`SPDX-License-Identifier: Elastic-2.0`. A `SENSOR=1` build names the module's
license in the About window. The portable packages and the macOS app also contain it as
`LICENSE-sensormonitor`. The frontend variant is written to `frontend/.variant` and
read by `frontend/vite.config.ts`. `task publish:downloads` refuses
`SENSOR=1`: public downloads are always the standard build.

MarkSafe 1.8 kept the address in `config.json`. The module takes it over
once into its own `sensormonitor.json`.

