// The folder check flow in the frontend (plan P-MARKS-20261112-01): when it
// asks, runs, offers the commit and shows findings. The backend is mocked.
import { beforeEach, describe, expect, it, vi } from 'vitest';

const be = vi.hoisted(() => ({
  getFolderCheck: vi.fn(),
  trustFolderCheck: vi.fn(),
  runFolderCheck: vi.fn(),
  commitSuggestion: vi.fn(),
  commitFolderFile: vi.fn(),
}));
vi.mock('./backend', () => be);

const DIR = '/docs/seeds';
const DOC = DIR + '/sop.md';
const OTHER = DIR + '/sop.en.md';
const spec = (commit = true, hash = 'h1') => ({ dir: DIR, version: 1, commit, hash, check: { command: ['bin/check', '.'], cwd: DIR, timeout: 0 } });
const clean = { findings: [] };
const withError = {
  findings: [
    { file: 'sop.md', line: 3, col: 1, severity: 'error', code: 'x', message: 'bad', path: DOC },
    { file: 'sop.en.md', severity: 'error', message: 'version', path: OTHER },
    { file: 'sop.md', line: 5, severity: 'warning', message: 'meh', path: DOC },
  ],
};
const tr = (key: string, ...args: unknown[]) => [key, ...args].join('|');

let fc: typeof import('./folderCheck.svelte');
let ask: ReturnType<typeof vi.fn<import('./folderCheck.svelte').Ask>>;

beforeEach(async () => {
  vi.resetModules();
  Object.values(be).forEach((f) => f.mockReset());
  be.commitSuggestion.mockResolvedValue('SOP-1 v1: ');
  ask = vi.fn<import('./folderCheck.svelte').Ask>().mockResolvedValue(true);
  fc = await import('./folderCheck.svelte');
});

describe('folder check', () => {
  it('does nothing without a .marksafe.yaml', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: null, trusted: false });
    await fc.folderCheck.afterSave(DOC, ask, tr);
    expect(ask).not.toHaveBeenCalled();
    expect(be.runFolderCheck).not.toHaveBeenCalled();
    expect(fc.folderCheck.runFor(DOC)).toBeUndefined();
  });

  it('asks once with the command, then runs and offers the commit', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: false });
    be.runFolderCheck.mockResolvedValue(clean);
    await fc.folderCheck.afterSave(DOC, ask, tr);
    expect(ask).toHaveBeenCalledOnce();
    expect(ask.mock.calls[0][1]).toContain('bin/check .');
    expect(be.trustFolderCheck).toHaveBeenCalledWith(DOC, 'h1');
    expect(fc.isClean(fc.folderCheck.runFor(DOC)!)).toBe(true);
    expect(fc.folderCheck.commitRequest).toEqual({ path: DOC, message: 'SOP-1 v1: ' });
  });

  it('"Not now" is remembered for the session; "check again" asks again', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: false });
    ask.mockResolvedValue(false);
    await fc.folderCheck.afterSave(DOC, ask, tr);
    await fc.folderCheck.afterSave(DOC, ask, tr);
    expect(ask).toHaveBeenCalledOnce();
    expect(be.runFolderCheck).not.toHaveBeenCalled();
    expect(fc.folderCheck.runFor(DOC)?.refused).toBe(true);
    expect(fc.folderCheck.uncommitted[DOC]).toBe(true);

    // Allowed later via the list: the saved changes are still to commit.
    ask.mockResolvedValue(true);
    be.runFolderCheck.mockResolvedValue(clean);
    await fc.folderCheck.rerun(DOC, ask, tr);
    expect(ask).toHaveBeenCalledTimes(2);
    expect(fc.folderCheck.commitRequest?.path).toBe(DOC);
  });

  it('errors block the commit and stay "not committed"', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: true });
    be.runFolderCheck.mockResolvedValue(withError);
    await fc.folderCheck.afterSave(DOC, ask, tr);
    expect(ask).not.toHaveBeenCalled();
    expect(fc.folderCheck.commitRequest).toBeNull();
    expect(fc.folderCheck.uncommitted[DOC]).toBe(true);
    expect(fc.counts(fc.folderCheck.runFor(DOC)!.result!.findings)).toEqual({ errors: 2, warnings: 1 });
    await fc.folderCheck.requestCommit(DOC);
    expect(fc.folderCheck.commitRequest).toBeNull();
  });

  it('only warnings allow the commit', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: true });
    be.runFolderCheck.mockResolvedValue({ findings: [withError.findings[2]] });
    await fc.folderCheck.afterSave(DOC, ask, tr);
    expect(fc.folderCheck.commitRequest?.path).toBe(DOC);
  });

  it('a failed or not applicable check is never clean', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: true });
    for (const result of [{ findings: [], failed: true, stderr: 'boom' }, { findings: [], error: 'not a folder' }]) {
      be.runFolderCheck.mockResolvedValue(result);
      await fc.folderCheck.afterSave(DOC, ask, tr);
      expect(fc.isClean(fc.folderCheck.runFor(DOC)!)).toBe(false);
      expect(fc.folderCheck.commitRequest).toBeNull();
    }
  });

  it('without commit: true there is no commit dialog and nothing "not committed"', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(false), trusted: true });
    be.runFolderCheck.mockResolvedValue(clean);
    await fc.folderCheck.afterSave(DOC, ask, tr);
    expect(fc.folderCheck.commitRequest).toBeNull();
    expect(fc.folderCheck.uncommitted[DOC]).toBeUndefined();
  });

  it('"check again" alone opens no commit dialog after a commit', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: true });
    be.runFolderCheck.mockResolvedValue(clean);
    await fc.folderCheck.afterSave(DOC, ask, tr);
    expect(await fc.folderCheck.commit(DOC, 'msg', tr)).toBeNull();
    expect(fc.folderCheck.uncommitted[DOC]).toBeUndefined();
    await fc.folderCheck.rerun(DOC, ask, tr);
    expect(fc.folderCheck.commitRequest).toBeNull();
  });

  it('commit errors come back in the user\'s words; "nothing to commit" counts as done', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: true });
    be.runFolderCheck.mockResolvedValue(clean);
    await fc.folderCheck.afterSave(DOC, ask, tr);
    be.commitFolderFile.mockRejectedValueOnce(new Error('git user.name or user.email is not set on this computer'));
    expect(await fc.folderCheck.commit(DOC, 'msg', tr)).toBe('commitNoIdentity');
    expect(fc.folderCheck.uncommitted[DOC]).toBe(true);
    be.commitFolderFile.mockRejectedValueOnce(new Error('git commit: hook says no: exit status 1'));
    expect(await fc.folderCheck.commit(DOC, 'msg', tr)).toContain('hook says no');
    be.commitFolderFile.mockRejectedValueOnce(new Error('the file has no changes to commit'));
    expect(await fc.folderCheck.commit(DOC, 'msg', tr)).toBeNull();
    expect(fc.folderCheck.uncommitted[DOC]).toBeUndefined();
  });

  it('a newer save wins over a slower older run', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: true });
    let finishOld!: (r: unknown) => void;
    be.runFolderCheck.mockImplementationOnce(() => new Promise((r) => (finishOld = r)));
    be.runFolderCheck.mockResolvedValueOnce(clean);
    const old = fc.folderCheck.afterSave(DOC, ask, tr);
    await vi.waitFor(() => expect(finishOld).toBeTypeOf('function'));
    await fc.folderCheck.afterSave(DOC, ask, tr);
    finishOld(withError);
    await old;
    expect(fc.isClean(fc.folderCheck.runFor(DOC)!)).toBe(true);
  });

  it('shows findings inline only for the named, unchanged file, with a line', async () => {
    be.getFolderCheck.mockResolvedValue({ spec: spec(), trusted: true });
    be.runFolderCheck.mockResolvedValue(withError);
    await fc.folderCheck.afterSave(DOC, ask, tr);
    const inline = fc.folderCheck.inline(DOC, false, tr);
    expect(inline.map((d) => [d.line, d.severity])).toEqual([[3, 'error'], [5, 'warning']]);
    expect(inline[0].source).toBe('folderCheckTitle');
    expect(fc.folderCheck.inline(DOC, true, tr)).toEqual([]);
    // The translation's finding has no line: list only, but its tab knows the run.
    expect(fc.folderCheck.inline(OTHER, false, tr)).toEqual([]);
    expect(fc.folderCheck.runFor(OTHER)?.dir).toBe(DIR);
  });

  it('reports an invalid .marksafe.yaml', async () => {
    be.getFolderCheck.mockRejectedValue(new Error('version 9 is not supported'));
    await fc.folderCheck.afterSave(DOC, ask, tr);
    expect(fc.folderCheck.runFor(DOC)?.error).toContain('version 9');
    expect(be.runFolderCheck).not.toHaveBeenCalled();
  });
});
