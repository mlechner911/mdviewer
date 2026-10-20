<script lang="ts">
  /**
   * How the Sensor Monitor reads this document (MLC Markdown Extension,
   * non-standard): ID, kind, version, title, event kinds, fields — taken
   * from mlcext.Analyze, the same code the server uses. Title, id and
   * version are already in the document header (<MlcMarkdown>), so this
   * is a collapsed part below it. Editing aid only: not printed, not
   * exported.
   */
  let { info } = $props<{
    info: {
      refNo: string; kindLabel: string; version: string; title: string; description: string;
      eventKinds: string[] | null; eventKindsSet: boolean; fields: number; prefillFields: number; fieldErrors: number;
      headerInBody?: boolean;
    };
  }>();

  const problems = $derived(
    (!info.refNo ? 1 : 0) + (!info.version ? 1 : 0) + (info.refNo && !info.kindLabel ? 1 : 0) + (info.fieldErrors ? 1 : 0) + (info.headerInBody ? 1 : 0),
  );
</script>

<details class="mlc-docinfo" open={problems > 0}>
  <summary>
    MLC-Erweiterung · so liest der Sensor Monitor dieses Dokument
    {#if problems}<span class="error">· {problems} {problems === 1 ? 'Hinweis' : 'Hinweise'}</span>{/if}
  </summary>
  <dl>
    <dt>Kennung</dt>
    {#if info.refNo}<dd>{info.refNo}</dd>{:else}<dd class="missing">fehlt (<code>doc_id</code> im Front Matter)</dd>{/if}
    <dt>Art</dt>
    {#if info.kindLabel}<dd>{info.kindLabel}</dd>{:else if info.refNo}<dd class="missing">unbekannt (Präfix SOP-, VA-, FB- oder INFO-)</dd>{:else}<dd class="missing">fehlt</dd>{/if}
    <dt>Version</dt>
    {#if info.version}<dd>{info.version}</dd>{:else}<dd class="missing">fehlt (<code>version</code> im Front Matter)</dd>{/if}
    {#if info.headerInBody}
      <dt>Kopfzeile</dt>
      <dd class="missing">steht noch im Text (alte Form); Kennung und Version gehören ins Front Matter</dd>
    {/if}
    {#if info.description}<dt>Beschreibung</dt><dd>{info.description}</dd>{/if}
    {#if info.eventKindsSet}
      <dt>Gilt bei</dt>
      <dd>
        {#if info.eventKinds?.length}
          {#each info.eventKinds as k}<span class="tag">{k}</span>{/each}
        {:else}gilt allgemein{/if}
      </dd>
    {/if}
    {#if info.fields || info.fieldErrors}
      <dt>Formularfelder</dt>
      <dd>
        {info.fields}{#if info.prefillFields}, davon {info.prefillFields} aus einem Ereignis vorausgefüllt{/if}{#if info.fieldErrors}, <span class="error">{info.fieldErrors} fehlerhaft</span>{/if}
      </dd>
    {/if}
  </dl>
</details>

<style>
  .mlc-docinfo {
    border: 1px solid var(--mlc-md-border);
    border-left: 3px solid #c94c35;
    border-radius: var(--mlc-md-radius);
    background: var(--mlc-md-surface);
    margin-bottom: 1.5rem;
    padding: 0.4rem 1rem;
    font-size: 0.8rem;
  }
  summary {
    cursor: pointer;
    font-family: var(--mlc-md-font-mono);
    font-size: 0.68rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--mlc-md-muted);
  }
  dl { display: grid; grid-template-columns: max-content 1fr; gap: 0.15rem 1rem; margin: 0.4rem 0 0.2rem; }
  dt { color: var(--mlc-md-muted); font-weight: 500; }
  dd { margin: 0; }
  code { font-size: 0.95em; }
  .missing { color: var(--mlc-md-alert-warning); font-style: italic; }
  .error { color: var(--mlc-md-alert-caution); font-weight: 600; }
  .tag { display: inline-block; padding: 0.05rem 0.45rem; margin: 0 0.25rem 0.1rem 0; font-family: var(--mlc-md-font-mono); font-size: 0.72rem; border-radius: 9999px; background: var(--mlc-md-border); }
  @media print { .mlc-docinfo { display: none; } }
</style>
