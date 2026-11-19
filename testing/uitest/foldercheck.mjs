// UI test of the folder check over CDP (docs/UI-TEST.md): a test build
// (test-marksafe-uitest.exe) has the folder testing/uitest/foldercheck open,
// this script goes through permission, finding, exit code 2 and a clean run.
//
//   node foldercheck.mjs [port] [folder]
//
// port: the debugging port (default 9333). folder: where the test folder
// lies, to switch the stand-in check (mode.txt); default the folder next to
// this script. Run from another machine through an SSH tunnel, set
// UITEST_SSH=user@host and folder to the path on that machine.
// Needs playwright or playwright-core.
import { chromium } from 'playwright-core';
import { execFileSync } from 'node:child_process';
import { writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const port = process.argv[2] ?? '9333';
const folder = process.argv[3] ?? fileURLToPath(new URL('./foldercheck', import.meta.url));
const remote = process.env.UITEST_SSH;

function setMode(mode) {
  if (remote) execFileSync('ssh', ['-o', 'BatchMode=yes', remote, `cmd /c "echo ${mode}> ${folder}\\mode.txt"`], { stdio: 'ignore' });
  else writeFileSync(`${folder}/mode.txt`, mode + '\n');
}
const ok = (what) => console.log('ok  ', what);

setMode('findings');
const browser = await chromium.connectOverCDP(`http://127.0.0.1:${port}`);
const page = browser.contexts()[0].pages()[0];

// A document opened by argument starts in the preview.
const toggle = page.getByTestId('toggle-editor');
if ((await toggle.getAttribute('data-hidden')) === 'true') await toggle.click();
await page.locator('.cm-content').waitFor();
ok('editor');

async function editAndSave(text) {
  await page.locator('.cm-content').click();
  await page.keyboard.press('Control+End');
  await page.keyboard.type(text);
  await page.keyboard.press('Control+s');
}
async function openPanel() {
  if (!(await page.getByTestId('foldercheck-panel').isVisible())) await page.getByTestId('status-foldercheck').click();
}

// 1. First save: MarkSafe asks for permission; allowed, the check finds an error.
await editAndSave('\nErste Änderung.');
await page.getByTestId('foldercheck-permission').waitFor({ timeout: 15000 });
await page.getByTestId('foldercheck-permission-confirm').click();
ok('permission allowed');
await page.getByTestId('status-foldercheck').waitFor({ timeout: 20000 });
await openPanel();
const finding = page.getByTestId('foldercheck-finding').first();
await finding.waitFor({ timeout: 20000 });
if ((await finding.getAttribute('data-code')) !== 'version-unchanged') throw new Error('unexpected finding');
ok('finding version-unchanged');
await page.getByTestId('foldercheck-close').click();

// 2. The check program does not fit the checkout: exit code 2 and its message.
setMode('exit2');
await editAndSave('\nZweite Änderung.');
await page.waitForTimeout(1500);
await openPanel();
const error = page.getByTestId('foldercheck-error');
await error.waitFor({ timeout: 20000 });
if (!(await error.textContent()).includes('Prüfprogramm neu bauen')) throw new Error('exit 2 message missing');
ok('exit code 2 shown');

// 3. Fixed: checking again finds nothing.
setMode('ok');
await page.getByTestId('foldercheck-rerun').click();
await page.getByTestId('foldercheck-ok').waitFor({ timeout: 20000 });
ok('check again: no findings');

await browser.close();
