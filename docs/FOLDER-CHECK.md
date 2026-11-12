# Folder check and commit on save (`.marksafe.yaml`)

A folder can ask MarkSafe to check it after every save, and to commit the
saved file when the check finds no errors. Typical use: a folder of templates
or SOPs kept in git, with a project-specific checker that knows the folder's
rules (translations, catalogs, signature flows …).

MarkSafe knows nothing about these rules. The folder brings them along in a
small file, `.marksafe.yaml`, that names a command. Without that file nothing
changes.

## The file

Put `.marksafe.yaml` into the folder the check covers:

```yaml
# .marksafe.yaml
version: 1
check:
  command: ["bin/check", "-json"]   # program and arguments, no shell
  cwd: .                            # working folder, relative to this file
  timeout: 60s                      # optional, default 60s
commit: true                        # optional: saving also commits
```

| Key | Meaning |
| --- | --- |
| `version` | Format version. Only `1` exists; another value is reported as an error. |
| `check.command` | The program and its arguments, as a list. It runs **without a shell**: no pipes, no `&&`, no wildcards. A bare name (`go`) is looked up in `PATH`; a path (`bin/check`) is relative to `cwd`. On Windows MarkSafe tries the name with `.exe` first. |
| `check.cwd` | Working folder of the command, relative to `.marksafe.yaml`. Default: the folder of `.marksafe.yaml`. |
| `check.timeout` | Time limit, e.g. `30s`, `2m`. A check that runs longer is stopped and counts as failed. |
| `commit` | `true`: after a check without errors, MarkSafe offers to commit the saved file. Needs a `check`. |

**Which file applies:** MarkSafe looks for `.marksafe.yaml` in the folder of
the saved file and then upwards, at most up to the document folder (or, for a
file outside it, up to the allowed folder that holds it). The first file found
applies. Folders MarkSafe has not been allowed to open are never searched.

## Permission

`.marksafe.yaml` makes MarkSafe start a program, so MarkSafe asks first. The
first time a file in the folder is saved, a dialog shows the command and its
working folder:

- **Allow** remembers the permission for this folder and exactly this command.
  When `check` or `commit` in `.marksafe.yaml` change, MarkSafe asks again.
  Comments and formatting do not count as a change.
- **Not now** saves the file without a check and does not ask again until
  MarkSafe restarts. **Check again** in the list of findings asks once more.

Only allow a folder you trust: the command runs with your rights.

## What the check must print

The command prints **one line of JSON** on stdout. Lines before it (build
output, logs) are ignored; only the last non-empty line counts.

```json
{"ok":false,"findings":[{"file":"common/sop.md","line":12,"col":3,"severity":"error","code":"signature-flow","message":"Step sig_lead has no signature field"}]}
```

| Field | Meaning |
| --- | --- |
| `findings[].file` | Path relative to the folder of `.marksafe.yaml`, with `/`. Empty for a finding about the whole folder. |
| `findings[].line`, `col` | 1-based position. Missing or `0` when the rule knows no position. |
| `findings[].severity` | `error` or `warning`. Anything else counts as an error. |
| `findings[].code` | Short rule name, shown next to the message. Optional. |
| `findings[].message` | Text for the user. |
| `error` | Only with exit code 2, see below. |

**Exit codes:**

| Code | Meaning | MarkSafe shows |
| --- | --- | --- |
| 0 | Checked, no findings (or only warnings) | the findings, if any |
| 1 | Checked, with findings | the findings |
| 2 | The folder is not one this check knows | the message from `error` |
| other | The check itself failed | "Folder check failed" with the end of stderr |

A check that times out, cannot be started, prints no JSON, or exits with 1
but without findings also counts as failed.

## Where the findings appear

- **Status bar:** "Folder check: 2 error(s), 1 warning(s)", "no problems",
  "running…" or "failed". A click opens the list.
- **List:** all findings of the folder, grouped by file; findings without a
  file come first. A click opens the file at the line. **Check again** runs
  the check once more.
- **Editor:** findings with a line show in the text of every open file they
  name, like the editor's own checks. Once you type, they disappear from that
  file until the next save, because their line numbers no longer fit.

## Commit on save

With `commit: true`, a check without errors opens the commit dialog:

- Only **the saved file** is committed (`git add`, then `git commit --only`).
  Other changes you have staged stay as they are.
- The message is prefilled with `<doc_id> v<version>: ` from the front matter,
  or with the file name. `Ctrl+Enter` commits.
- The author is the git identity of the computer (`user.name`, `user.email`).
- MarkSafe **never pushes**. Pushing, review and release stay with you.

The commit is refused when the check found errors, failed or did not run, and
when the file changed after the check. The status bar then shows **Saved, not
committed**. Fix the findings and save again; a click on the entry commits
later once the check is clean. **Later** in the dialog leaves the file saved
and not committed in the same way.

Git hooks of the repository run as usual; a hook that rejects the commit shows
its message in the dialog. MarkSafe also says so when git is not installed,
the file is not in a repository, or no git identity is set.

## Example

The MLC Sensor Monitor keeps its shipped templates in
`internal/sop/seeds/` and checks them with its own tool:

```yaml
# internal/sop/seeds/.marksafe.yaml
version: 1
check:
  command: ["go", "run", "./cmd/vorlagenkatalog", "-pruefen", "internal/sop/seeds", "-json"]
  cwd: ../../..
  timeout: 120s
commit: true
```

On a computer without Go the command points to a prebuilt binary instead.
