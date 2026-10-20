<script lang="ts">
  /**
   * How the Sensor Monitor reads this document (MLC Markdown Extension,
   * non-standard): ID, kind, version, title, event kinds, fields — taken
   * from mlcext.Analyze, the same code the server uses. Editing aid only:
   * not printed, not exported.
   */
  let { info } = $props<{
    info: {
      refNo: string; kindLabel: string; version: string; title: string; description: string;
      eventKinds: string[] | null; eventKindsSet: boolean; fields: number; prefillFields: number; fieldErrors: number;
    };
  }>();
</script>

<div class="mlc-docinfo">
  <div class="mlc-docinfo-title">MLC-Erweiterung · so liest der Sensor Monitor dieses Dokument</div>
  <dl>
    <dt>Kennung</dt>
    {#if info.refNo}<dd>{info.refNo}</dd>{:else}<dd class="missing">fehlt</dd>{/if}
    <dt>Art</dt>
    {#if info.kindLabel}<dd>{info.kindLabel}</dd>{:else if info.refNo}<dd class="missing">unbekannt (Präfix SOP-, VA-, FB- oder INFO-)</dd>{:else}<dd class="missing">fehlt</dd>{/if}
    <dt>Version</dt>
    {#if info.version}<dd>{info.version}</dd>{:else}<dd class="missing">fehlt</dd>{/if}
    {#if info.title}<dt>Titel</dt><dd>{info.title}</dd>{/if}
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
</div>

<style>
  .mlc-docinfo {
    border: 1px solid var(--mlc-md-border);
    border-left: 3px solid #c94c35;
    border-radius: var(--mlc-md-radius);
    background: var(--mlc-md-surface);
    margin-bottom: 1.5rem;
    padding: 0.6rem 1rem;
    font-size: 0.8rem;
  }
  .mlc-docinfo-title {
    font-family: var(--mlc-md-font-mono);
    font-size: 0.68rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--mlc-md-muted);
    margin-bottom: 0.4rem;
  }
  dl { display: grid; grid-template-columns: max-content 1fr; gap: 0.15rem 1rem; margin: 0; }
  dt { color: var(--mlc-md-muted); font-weight: 500; }
  dd { margin: 0; }
  .missing { color: var(--mlc-md-alert-warning); font-style: italic; }
  .error { color: var(--mlc-md-alert-caution); font-weight: 600; }
  .tag { display: inline-block; padding: 0.05rem 0.45rem; margin: 0 0.25rem 0.1rem 0; font-family: var(--mlc-md-font-mono); font-size: 0.72rem; border-radius: 9999px; background: var(--mlc-md-border); }
  @media print { .mlc-docinfo { display: none; } }
</style>
