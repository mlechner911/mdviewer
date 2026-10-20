<script lang="ts">
  /**
   * NewFromTemplateModal creates a new document of the MLC Markdown
   * Extension (non-standard) from a built-in template: kind, title, area and
   * RefNo. The RefNo suggestion comes from the files next to the current
   * document; only the Sensor Monitor knows which RefNos are really taken.
   */
  import { t } from '../i18n';
  import { STYLE } from '../lib/constants';
  import type { EffectiveTheme_t } from '../lib/constants';
  import * as backend from '../lib/backend';

  let { show = false, theme = 'dark', docPath = null, onCreate, onCancel } = $props<{
    show?: boolean;
    theme?: EffectiveTheme_t;
    docPath?: string | null;
    onCreate: (content: string, refNo: string) => void;
    onCancel: () => void;
  }>();

  const currentTheme = $derived<EffectiveTheme_t>(theme === 'light' ? 'light' : 'dark');
  const toolbarClass = $derived(STYLE.toolbar[currentTheme]);
  const buttonClass = $derived(STYLE.button[currentTheme]);
  const editorClass = $derived(STYLE.editor[currentTheme]);

  let templates = $state<{ kind: string; prefix: string; label: string }[]>([]);
  let kind = $state('sop');
  let title = $state('');
  let area = $state('QM');
  let refNo = $state('');
  let refNoEdited = $state(false);
  let version = $state('1.0');
  let error = $state('');

  $effect(() => {
    if (show && templates.length === 0) {
      backend.mlcTemplates().then((list) => (templates = list));
    }
  });

  // Keep the RefNo suggestion in step with kind and area until edited.
  $effect(() => {
    const k = kind, a = area, p = docPath, open = show;
    if (!open || refNoEdited) return;
    backend.suggestMLCRefNo(k, a, p ?? '').then((s) => {
      if (!refNoEdited) refNo = s;
    });
  });

  async function create() {
    error = '';
    const res = await backend.newMLCDocument({ kind, title, refNo, version });
    if (res.error) {
      error = res.error;
      return;
    }
    onCreate(res.content, refNo);
    title = '';
    refNoEdited = false;
  }

  function onKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onCancel();
    if (e.key === 'Enter' && (e.target as HTMLElement).tagName === 'INPUT') create();
  }
</script>

{#if show}
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm print:hidden" onkeydown={onKeydown}>
  <div class="max-w-lg w-full rounded-lg shadow-2xl overflow-hidden border {toolbarClass}">
    <div class="p-6 flex flex-col gap-3 text-sm">
      <h3 class="text-lg font-bold">{$t('tplHeading')}</h3>
      <p class="text-xs opacity-70">{$t('tplIntro')}</p>

      <label class="flex flex-col gap-1">
        <span class="text-xs font-semibold opacity-80">{$t('tplKind')}</span>
        <select bind:value={kind} class="px-2 py-1.5 rounded border {editorClass}">
          {#each templates as tpl}
            <option value={tpl.kind}>{tpl.label}</option>
          {/each}
        </select>
      </label>

      <label class="flex flex-col gap-1">
        <span class="text-xs font-semibold opacity-80">{$t('tplDocTitle')}</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input bind:value={title} autofocus class="px-2 py-1.5 rounded border {editorClass}" placeholder="Lagerung kühlkettenpflichtiger Arzneimittel" />
      </label>

      <div class="grid grid-cols-3 gap-3">
        <label class="flex flex-col gap-1">
          <span class="text-xs font-semibold opacity-80">{$t('tplArea')}</span>
          <input bind:value={area} class="px-2 py-1.5 rounded border font-mono uppercase {editorClass}" placeholder="QM" />
        </label>
        <label class="flex flex-col gap-1 col-span-2">
          <span class="text-xs font-semibold opacity-80">{$t('tplRefNo')}</span>
          <input bind:value={refNo} oninput={() => (refNoEdited = true)} class="px-2 py-1.5 rounded border font-mono {editorClass}" />
        </label>
      </div>
      <p class="text-[11px] opacity-60 -mt-1">{$t('tplRefNoHint')}</p>

      <label class="flex flex-col gap-1 w-32">
        <span class="text-xs font-semibold opacity-80">{$t('tplVersion')}</span>
        <input bind:value={version} class="px-2 py-1.5 rounded border font-mono {editorClass}" />
      </label>

      {#if error}
        <div class="p-2 rounded text-xs bg-red-600/15 text-red-500">{error}</div>
      {/if}

      <div class="flex justify-end gap-3 mt-2">
        <button onclick={onCancel} class="px-4 py-2 rounded text-sm font-medium transition-colors {buttonClass}">{$t('tplCancel')}</button>
        <button onclick={create} class="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white transition-colors">{$t('tplCreate')}</button>
      </div>
    </div>
  </div>
</div>
{/if}
