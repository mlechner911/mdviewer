/**
 * Folder check after saving (.marksafe.yaml, plan P-MARKS-20261112-01).
 * A folder names a check command; MarkSafe runs it after each save, once
 * the user allowed that exact command, and shows the findings: inline for
 * the open, unchanged files, and as a list for the whole folder.
 */
import * as backend from './backend';
import type { FolderCheckResult, FolderFinding, FolderCheckSpec } from './backend';
import type { MlcDiagnostic } from '../../../packages/mlcext-editor/src/index';

/** What the status bar and the list show for one folder. */
export interface FolderRun {
  dir: string;
  running: boolean;
  /** The user did not allow the check (this session). */
  refused?: boolean;
  /** .marksafe.yaml is invalid or the backend refused. */
  error?: string;
  result?: FolderCheckResult;
  /** The folder commits on save (commit: true). */
  commit?: boolean;
}

/** The commit dialog: which file, the prefilled message. */
export interface CommitRequest {
  path: string;
  message: string;
}

/** Asks a yes/no question (App's ConfirmModal). */
export type Ask = (title: string, message: string, confirmLabel: string, cancelLabel: string) => Promise<boolean>;
export type Tr = (key: string, ...args: unknown[]) => string;

class FolderCheck {
  /** Folder (dir of .marksafe.yaml) → its last run. */
  runs = $state<Record<string, FolderRun>>({});
  /** File path → folder whose check covers it. */
  dirOf = $state<Record<string, string>>({});
  /** Bumped when findings change; the editor then shows them at once. */
  revision = $state(0);
  /** Saved but not committed yet (commit folders only). */
  uncommitted = $state<Record<string, boolean>>({});
  /** Open commit dialog, if any. */
  commitRequest = $state<CommitRequest | null>(null);
  /** dir|hash the user declined this session: not asked on every save. */
  private declined = new Set<string>();
  private seq = new Map<string, number>();

  /** The run that covers path, if any. */
  runFor(path: string | null | undefined): FolderRun | undefined {
    const dir = path ? this.dirOf[path] : undefined;
    return dir ? this.runs[dir] : undefined;
  }

  /**
   * After saving path: find, ask for permission if needed, run. saved is
   * false for "check again": that alone leaves nothing new to commit.
   */
  async afterSave(path: string, ask: Ask, tr: Tr, saved = true): Promise<void> {
    let spec: FolderCheckSpec | null | undefined;
    let trusted = false;
    try {
      const info = await backend.getFolderCheck(path);
      spec = info?.spec;
      trusted = !!info?.trusted;
    } catch (err) {
      this.set(path, path, { dir: path, running: false, error: String(err) });
      return;
    }
    if (!spec?.check) {
      delete this.dirOf[path];
      return;
    }
    const dir = spec.dir;
    if (spec.commit && saved) this.uncommitted[path] = true;
    if (!trusted && !(await this.allow(path, spec, ask, tr))) {
      this.set(path, dir, { dir, running: false, refused: true, commit: spec.commit });
      return;
    }
    await this.run(path, dir, spec.commit);
  }

  /** Check again (list dialog); asks again if the user declined before. */
  async rerun(path: string, ask: Ask, tr: Tr): Promise<void> {
    const dir = this.dirOf[path];
    if (!dir) return;
    for (const key of [...this.declined]) if (key.startsWith(dir + '|')) this.declined.delete(key);
    await this.afterSave(path, ask, tr, false);
  }

  /** Findings for the editor: only for a tab whose text is as saved. */
  inline(path: string | null | undefined, dirty: boolean, tr: Tr): MlcDiagnostic[] {
    if (!path || dirty) return [];
    const findings = this.runFor(path)?.result?.findings ?? [];
    return findings
      .filter((f) => f.path === path && (f.line ?? 0) > 0)
      .map((f) => ({
        line: f.line ?? 1,
        col: f.col || 1,
        // Anything but a warning counts as an error, as in the backend.
        severity: f.severity === 'warning' ? 'warning' : 'error',
        code: f.code ?? '',
        message: f.message,
        source: tr('folderCheckTitle'),
      }));
  }

  private async allow(path: string, spec: FolderCheckSpec, ask: Ask, tr: Tr): Promise<boolean> {
    const key = spec.dir + '|' + spec.hash;
    if (this.declined.has(key)) return false;
    const cmd = (spec.check?.command ?? []).map(quote).join(' ');
    const ok = await ask(
      tr('folderCheckTrustTitle'),
      tr('folderCheckTrustBody', spec.dir, cmd, spec.check?.cwd ?? spec.dir),
      tr('folderCheckTrustAllow'),
      tr('folderCheckTrustDeny'),
    );
    if (!ok) {
      this.declined.add(key);
      return false;
    }
    try {
      await backend.trustFolderCheck(path, spec.hash);
      return true;
    } catch (err) {
      this.set(path, spec.dir, { dir: spec.dir, running: false, error: String(err) });
      return false;
    }
  }

  private async run(path: string, dir: string, commit: boolean): Promise<void> {
    const n = (this.seq.get(dir) ?? 0) + 1;
    this.seq.set(dir, n);
    // Old findings point at lines of the old text: drop them while running.
    this.set(path, dir, { dir, running: true, commit });
    let next: FolderRun;
    try {
      const result = await backend.runFolderCheck(path);
      next = { dir, running: false, result: result ?? undefined, commit };
    } catch (err) {
      next = { dir, running: false, error: String(err), commit };
    }
    // A newer save started another run: its result wins.
    if (this.seq.get(dir) !== n) return;
    this.set(path, dir, next);
    if (commit && isClean(next)) await this.requestCommit(path);
  }

  /** Opens the commit dialog for path, if its last check was clean. */
  async requestCommit(path: string): Promise<void> {
    const run = this.runFor(path);
    if (!run?.commit || !isClean(run) || !this.uncommitted[path]) return;
    const message = await backend.commitSuggestion(path);
    this.commitRequest = { path, message };
  }

  /** Commits; returns an error text for the dialog, or null when done. */
  async commit(path: string, message: string, tr: Tr): Promise<string | null> {
    try {
      await backend.commitFolderFile(path, message);
    } catch (err) {
      const text = String(err);
      if (!text.includes('no changes to commit')) return commitError(text, tr);
    }
    delete this.uncommitted[path];
    this.commitRequest = null;
    return null;
  }

  /** The dialog was closed without committing: stays "not committed". */
  cancelCommit() {
    this.commitRequest = null;
  }

  private set(path: string, dir: string, run: FolderRun) {
    this.dirOf[path] = dir;
    this.runs[dir] = run;
    // Every file the findings name belongs to this folder's check too.
    for (const f of run.result?.findings ?? []) if (f.path) this.dirOf[f.path] = dir;
    this.revision++;
  }
}

/** A finished run without errors (warnings are fine): may commit. */
export function isClean(run: FolderRun): boolean {
  if (run.running || run.refused || run.error || !run.result) return false;
  if (run.result.failed || run.result.error) return false;
  return counts(run.result.findings).errors === 0;
}

// Backend errors (foldercheck/git.go, foldercommit.go) in the user's words.
const COMMIT_ERRORS: [string, string][] = [
  ['git is not installed', 'commitNoGit'],
  ['not in a git repository', 'commitNoRepo'],
  ['user.name or user.email', 'commitNoIdentity'],
  ['changed since the last check', 'commitNotChecked'],
  ['commit message is empty', 'commitEmpty'],
];

function commitError(text: string, tr: Tr): string {
  const hit = COMMIT_ERRORS.find(([needle]) => text.includes(needle));
  return hit ? tr(hit[1]) : text;
}

function quote(arg: string): string {
  return /[\s"']/.test(arg) ? JSON.stringify(arg) : arg;
}

/** Counts for the status bar. */
export function counts(findings: FolderFinding[] | null | undefined): { errors: number; warnings: number } {
  const list = findings ?? [];
  const warnings = list.filter((f) => f.severity === 'warning').length;
  return { errors: list.length - warnings, warnings };
}

export const folderCheck = new FolderCheck();
