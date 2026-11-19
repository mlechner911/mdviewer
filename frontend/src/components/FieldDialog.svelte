<script lang="ts">
  /**
   * Form field of the MLC Markdown Extension (non-standard) as a dialog: the
   * shared <MlcFieldDialog> (packages/mlcext-editor-svelte), writing the
   * field through mlcext.FormatField over Wails.
   */
  import { MlcFieldDialog } from 'mlcext-editor-svelte';
  import { locale } from '../i18n';
  import { dialogLabels } from '../lib/dialogLabels';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as backend from '../lib/backend';

  let { initial = null, usedNames = [], prefill = [], properties = [], theme = 'dark', onApply, onClose }: {
    /** The field being edited; null inserts a new one. */
    initial?: backend.MLCFieldSpec | null;
    /** Names of the other fields in the document (duplicate warning). */
    usedNames?: string[];
    /** Field names the Sensor Monitor fills from an event. */
    prefill?: { name: string; label: string }[];
    /** Properties of the document's schema, with their description. */
    properties?: { name: string; label: string }[];
    theme?: EffectiveTheme_t;
    onApply: (text: string) => void;
    onClose: () => void;
  } = $props();
</script>

<MlcFieldDialog
  {initial}
  format={(spec) => backend.mlcFormatField(spec as backend.MLCFieldSpec)}
  {usedNames}
  {prefill}
  {properties}
  theme={theme === 'light' ? 'light' : 'dark'}
  locale={$locale}
  labels={$dialogLabels}
  {onApply}
  {onClose}
/>
