<script lang="ts">
  /**
   * Preview renders the active document with the shared component
   * <MlcMarkdown> (mlc-markdown), so a document looks the same in MarkSafe,
   * the Sensor Monitor and the MLC Terminal. MarkSafe adds its security
   * model through the component's hooks: local images only from
   * whitelisted folders (served by /local-resource, which checks again),
   * external images and links only for whitelisted domains — an image is
   * not requested before that check.
   */
  import 'katex/dist/katex.min.css';
  import 'mlc-markdown/styles.css';
  import hljsLight from 'highlight.js/styles/github.css?raw';
  import hljsDark from 'highlight.js/styles/github-dark.css?raw';
  import { MlcMarkdown } from 'mlc-markdown-svelte';
  import DocInfoCard from './DocInfoCard.svelte';
  import { createMdHooks, highlight, IMAGE_PLACEHOLDER, loadKatex, resolveUrl } from '../lib/mdHooks';
  import * as backend from '../lib/backend';
  import type { Theme } from '../themes';
  import { untrack } from 'svelte';
  import { exampleValues } from '../../../packages/mlcext-editor/src/schemaHelp';
  import { locale, t } from '../i18n';

  let {
    source = '',
    theme,
    fontSize = 100,
    centerImages = false,
    currentFilePath = null,
    imagesRevision = 0,
    analysis = null,
    onsecurity_request,
    onopen_file,
    onscrollline,
    onrendered,
  }: {
    source: string;
    theme: Theme;
    fontSize?: number;
    /** View › Center Images: images alone in their paragraph are centred. */
    centerImages?: boolean;
    currentFilePath?: string | null;
    /** Raised when a folder or domain was allowed: images are checked again. */
    imagesRevision?: number;
    /** mlcext.Analyze of the document: the "how the Sensor Monitor reads it" card. */
    analysis?: backend.MLCAnalysis | null;
    onsecurity_request?: (detail: { type: 'path' | 'url'; resource: string }) => void;
    onopen_file?: (detail: { path: string }) => void;
    onscrollline?: (line: number) => void;
    onrendered?: (el: HTMLElement) => void;
  } = $props();

  let container = $state<HTMLElement>();
  let view = $state<{ scrollToLine(line: number): void }>();
  const mlcDoc = $derived(!!analysis?.detected);

  // Example values from the schema: only what is shown here changes, never
  // the document, the PDF or the file.
  let examplesOn = $state(false);
  let filled = $state<{ from: string; text: string } | null>(null);
  const examples = $derived(exampleValues(analysis?.schema, $locale));
  const hasExamples = $derived(Object.keys(examples).length > 0);
  const showExamples = $derived(examplesOn && hasExamples);
  $effect(() => {
    const md = source;
    const values = examples;
    if (!showExamples) return;
    untrack(() => {
      void backend.mlcFillForm(md, values).then((text) => {
        if (md === source) filled = { from: md, text };
      });
    });
  });
  const shownSource = $derived(showExamples && filled?.from === source ? filled.text : source);

  const { checkImages, onNavigate, loadMermaid } = createMdHooks({
    filePath: () => currentFilePath,
    mermaid: () => ({ theme: theme.mermaidTheme, themeVariables: theme.mermaidVars || {} }),
    onsecurity: (detail) => onsecurity_request?.(detail),
    onopenfile: (detail) => onopen_file?.(detail),
  });

  async function rendered(el: HTMLElement) {
    await checkImages(el);
    onrendered?.(el);
  }

  /** The preview shows example values, not what the document says. */
  export function showingExamples(): boolean {
    return showExamples;
  }

  /** Scroll so that source line `line` is at the top. */
  export function scrollToLine(line: number) {
    view?.scrollToLine(line);
  }

  /**
   * getExportHTML returns the rendered document (diagrams, formulas, images)
   * as a self-contained element: local images become data: URIs, each
   * diagram keeps its Mermaid source in an inert <template>, formulas keep
   * their TeX in KaTeX's MathML annotation.
   */
  export async function getExportHTML(): Promise<string> {
    const root = container?.querySelector('.mlc-md');
    if (!root) return '';
    const clone = root.cloneNode(true) as HTMLElement;
    clone.removeAttribute('style');
    // Editing aids, not content.
    clone.querySelectorAll('.mlc-docinfo, .mlc-md-frontmatter').forEach((el) => el.remove());
    for (const img of Array.from(clone.querySelectorAll('img'))) {
      const src = img.getAttribute('src') ?? '';
      if (src.startsWith(IMAGE_PLACEHOLDER)) {
        img.remove(); // not allowed, never loaded
        continue;
      }
      if (!src.startsWith('/local-resource')) continue;
      try {
        const blob = await (await fetch(src)).blob();
        img.src = await new Promise<string>((resolve, reject) => {
          const reader = new FileReader();
          reader.onload = () => resolve(reader.result as string);
          reader.onerror = () => reject(reader.error);
          reader.readAsDataURL(blob);
        });
      } catch (err) {
        console.error('Export: image not embedded:', src, err);
      }
    }
    return clone.outerHTML;
  }

  /** The highlight.js stylesheet of the active theme (also for the export). */
  export function codeCSS(): string {
    return theme.id === 'dark' ? hljsDark : hljsLight;
  }
</script>

{@html '<' + 'style' + '>' + (theme.id === 'dark' ? hljsDark : hljsLight) + '</' + 'style' + '>'}

{#if hasExamples}
  <label class="flex items-center gap-2 px-8 py-1 text-xs {theme.containerClass} {showExamples ? 'text-amber-500' : 'opacity-70'}">
    <input type="checkbox" bind:checked={examplesOn} />
    {$t('previewExamples')}
  </label>
{/if}
<div bind:this={container} class="preview-container flex-1 overflow-auto p-8 {theme.containerClass}" style="--preview-scale: {fontSize / 100};">
  {#key `${theme.id}:${imagesRevision}`}
    <MlcMarkdown
      bind:this={view}
      source={shownSource}
      locale={$locale}
      theme={theme.id}
      frontMatter={mlcDoc ? 'hide' : 'box'}
      debounce={150}
      {centerImages}
      rawHtml="sanitize"
      {highlight}
      {resolveUrl}
      {onNavigate}
      {loadKatex}
      {loadMermaid}
      scrollContainer={container}
      onScrollLine={(line: number) => onscrollline?.(line)}
      onRendered={rendered}
    >
      {#snippet header()}
        {#if mlcDoc && analysis}<DocInfoCard info={analysis.document} />{/if}
      {/snippet}
    </MlcMarkdown>
  {/key}
</div>

<style>
  .preview-container :global(.mlc-md) { font-size: calc(1rem * var(--preview-scale, 1)); }

</style>
