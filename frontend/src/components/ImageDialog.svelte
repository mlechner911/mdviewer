<script lang="ts">
  /**
   * Insert or edit an image: the shared <MlcImageDialog> (packages/
   * mlcext-editor-svelte) with MarkSafe's file choice and image check over
   * Wails. A reference works only inside the document's folder: the PDF
   * reads nothing outside it.
   */
  import { MlcImageDialog } from 'mlcext-editor-svelte';
  import { locale } from '../i18n';
  import { dialogLabels } from '../lib/dialogLabels';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as backend from '../lib/backend';

  let { path = null, dataUri = null, existing = null, docPath = null, preferEmbed = false, theme = 'dark', onApply, onClose }: {
    /** A dropped file; null starts with the file choice. */
    path?: string | null;
    /** A picture from the clipboard. */
    dataUri?: string | null;
    /** An image of the document, ![alt](target): its text and where the alt text lies in it. */
    existing?: { text: string; alt: string; target: string; altFrom: number; altTo: number } | null;
    /** The document's file: references start from its folder. */
    docPath?: string | null;
    /** A form being filled in: the record should stay complete on its own. */
    preferEmbed?: boolean;
    theme?: EffectiveTheme_t;
    onApply: (text: string) => void;
    onClose: () => void;
  } = $props();

  const dir = () => docPath ?? '';
  // svelte-ignore state_referenced_locally
  const chosen = path ? backend.inspectImage(path, dir()) : null;
</script>

<MlcImageDialog
  {chosen}
  {dataUri}
  {existing}
  {preferEmbed}
  choose={async () => {
    const p = await backend.chooseImage(dir());
    return p ? await backend.inspectImage(p, dir()) : null;
  }}
  toMarkdown={(r) => backend.imageMarkdown({ path: r.choice?.path ?? '', dataUri: r.dataUri ?? '', alt: r.alt, embed: r.embed }, dir())}
  theme={theme === 'light' ? 'light' : 'dark'}
  locale={$locale}
  labels={$dialogLabels}
  {onApply}
  {onClose}
/>
