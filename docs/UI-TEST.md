# UI test build (Playwright over CDP)

A Windows test build that opens the WebView2 debugging port (Chrome DevTools
Protocol), so Playwright can drive MarkSafe with `chromium.connectOverCDP`.
All dialogs live in the WebView, so a test can click through them like a
user.

> [!CAUTION]
> Whoever reaches the port controls MarkSafe completely, including the
> permission and the command of the folder check. The test build is for test
> machines only and never handed out.

## Build

`task handtest:prepare` builds it next to the normal test builds, into
`dropzone/marksafe-test/`:

| File | What |
| --- | --- |
| `test-marksafe-uitest.exe` | MarkSafe with the debugging port |
| `test-marksafeplus-uitest.exe` | MarkSafePlus with the debugging port (with `SENSOR=1`) |

A single build: `task build:windows UITEST=1` (with `SENSOR=1` for Plus); the
result is `bin/marksafe.exe`.

**Fail closed:** only the Go build tag `uitest` compiles `uitest_on.go`; in
every other build `uitest_off.go` neither reads the environment nor opens a
port. `task release` refuses `UITEST=1` and fails when a release binary
contains `remote-debugging-port` or `[UI-TEST]`.

## Run

```powershell
$env:MARKSAFE_CDP_PORT = "9333"   # optional, 9333 is the default
.\test-marksafeplus-uitest.exe C:\test\vorlagen\sop.md
```

- The port listens on `127.0.0.1` only.
- Every window title starts with `[UI-TEST]`.
- Wails empties `WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS`, and the Edge policy
  `AdditionalBrowserArguments` has no effect either; only this build opens
  the port.

```ts
import { chromium } from 'playwright';

const browser = await chromium.connectOverCDP('http://127.0.0.1:9333');
const page = browser.contexts()[0].pages()[0];
await page.getByTestId('foldercheck-permission-confirm').click();
```

## Example: the folder check

`testing/uitest/` goes through the dialogs of the folder check: permission,
a finding, exit code 2 with its message and a clean run. The folder
`foldercheck/` holds a document, `.marksafe.yaml` and a stand-in check
(`check/`) whose answer comes from `mode.txt`.

```powershell
cd testing\uitest\foldercheck
cd check; go build -o ..\bin\check.exe .; cd ..       # its own Go module
..\..\..\test-marksafe-uitest.exe "$PWD\sop.md"   # the document next to .marksafe.yaml
cd ..; npm install playwright-core; node foldercheck.mjs
```

The document has to lie in the folder of `.marksafe.yaml` (or below it in
an opened folder): opened by argument, MarkSafe only allows the file's own
folder, and it never searches above an allowed folder. The first run asks
for permission; the script allows it. MarkSafe remembers that, so delete
the entry or change `.marksafe.yaml` to see the dialog again.

From another machine: tunnel the port (`ssh -L 9333:127.0.0.1:9333
user@host`), start the test build in the user's session (`schtasks /create
… /it`, then `schtasks /run`), and call `UITEST_SSH=user@host node
foldercheck.mjs 9333 C:\path\to\foldercheck`. Checked like that on win11
(I-MARKS-20261117-01).

## Test IDs

Stable `data-testid` attributes, independent of the interface language.

| Test ID | Element |
| --- | --- |
| `foldercheck-permission` | Permission dialog of the folder check |
| `foldercheck-permission-confirm` / `-cancel` | Allow / Deny |
| `toggle-editor` | Shows or hides the editor; `data-hidden="true"` while it is hidden (a document opened by argument starts in the preview). The editor itself is `.cm-content` |
| `status-foldercheck` | Folder check entry in the status bar (opens the panel) |
| `status-commit` | "Saved, not committed" in the status bar |
| `foldercheck-panel` | Panel with the result |
| `foldercheck-running` | Check is running |
| `foldercheck-refused` | Permission was denied |
| `foldercheck-ok` | No findings |
| `foldercheck-findings` | List of findings of one file (one per file) |
| `foldercheck-finding` | One finding; attributes `data-code` and `data-severity` |
| `foldercheck-error` | Message of exit code 2 |
| `foldercheck-failed`, `foldercheck-stderr` | The check failed, end of stderr |
| `foldercheck-rerun`, `foldercheck-close` | Buttons of the panel |
| `commit-dialog` | Commit dialog |
| `commit-message` | Commit message (prefilled `<doc_id> v<version>: `) |
| `commit-error` | Error of git |
| `commit-submit`, `commit-later` | Commit / Later |

Other confirmation dialogs carry a test ID only where one is set
(`askConfirm(…, testId)`); add one when a test needs it.
