// The texts of the shared dialogs (mlcext-editor-svelte) in MarkSafe's
// interface language, from its own translations (German, English, Spanish,
// French); "%s" stays for the dialog to fill.
import { derived } from 'svelte/store';
import { DIALOG_LABELS } from 'mlcext-editor-svelte';
import { t } from '../i18n';

export const dialogLabels = derived(t, ($t) => Object.fromEntries(Object.keys(DIALOG_LABELS.en).map((k) => [k, $t(k)])));
