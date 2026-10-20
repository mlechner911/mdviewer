<script lang="ts">
  // Overview of the documents of the MLC Markdown Extension (non-standard)
  // in the root: RefNo, kind, title, version and check results, read by the
  // shared checker mlcext. Status and approval live in the Sensor Monitor
  // only and are deliberately not shown.
  import { createEventDispatcher, onMount } from 'svelte';
  import type { MlcDocument } from './browseClient';

  let { docs }: { docs: MlcDocument[] } = $props();

  const dispatch = createEventDispatcher();
  let q = $state('');
  let kind = $state('');
  let input = $state<HTMLInputElement | null>(null);

  const kinds = $derived(
    [...new Map(docs.filter((d) => d.kind).map((d) => [d.kind, d.kindLabel])).entries()],
  );
  const shown = $derived(
    docs.filter((d) => {
      if (kind && d.kind !== kind) return false;
      const s = q.trim().toLowerCase();
      if (!s) return true;
      return [d.refNo, d.title, d.path].some((v) => v.toLowerCase().includes(s));
    }),
  );

  onMount(() => input?.focus());

  function onWindowKey(e: KeyboardEvent) {
    if (e.key === 'Escape') dispatch('close');
  }
</script>

<svelte:window onkeydown={onWindowKey} />

<div class="docs-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) dispatch('close'); }}>
  <div class="docs-modal" role="dialog" aria-modal="true" aria-label="Dokumente">
    <div class="docs-head">
      <strong>Dokumente</strong>
      <span class="docs-note">MLC-Erweiterung · so liest der Sensor Monitor sie · Status und Freigabe stehen nur im Sensor Monitor</span>
    </div>
    <div class="docs-filter">
      <input bind:this={input} bind:value={q} placeholder="Kennung, Titel oder Datei …" aria-label="Filter" spellcheck="false" />
      <select bind:value={kind} aria-label="Art">
        <option value="">Alle Arten</option>
        {#each kinds as [k, label]}
          <option value={k}>{label}</option>
        {/each}
      </select>
    </div>
    <div class="docs-table-wrap">
      <table class="docs-table">
        <thead>
          <tr><th>Kennung</th><th>Art</th><th>Titel</th><th>Version</th><th>Felder</th><th>Prüfung</th></tr>
        </thead>
        <tbody>
          {#each shown as d}
            <tr onclick={() => dispatch('open', { path: d.path })} title={d.path}>
              <td class="mono">{d.refNo || '—'}</td>
              <td>{d.kindLabel || '—'}</td>
              <td>{d.title}</td>
              <td class="mono">{d.version || '—'}</td>
              <td class="num">{d.fields || ''}</td>
              <td>
                {#if d.errors}<span class="bad">{d.errors} Fehler</span>{/if}
                {#if d.warnings}<span class="warn">{d.warnings} Warnung{d.warnings > 1 ? 'en' : ''}</span>{/if}
                {#if !d.errors && !d.warnings}<span class="ok">in Ordnung</span>{/if}
              </td>
            </tr>
          {:else}
            <tr><td colspan="6" class="empty">Keine Dokumente.</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
    <div class="docs-footer">{shown.length} von {docs.length} · <kbd>Esc</kbd> schließen</div>
  </div>
</div>

<style>
  .docs-overlay {
    position: fixed; inset: 0; z-index: 50;
    display: flex; justify-content: center; align-items: flex-start;
    padding: 8vh 1rem 1rem; background: rgba(15, 23, 42, 0.55);
  }
  .docs-modal {
    width: min(980px, 96vw); max-height: 84vh; display: flex; flex-direction: column;
    background: var(--surface); border: 1px solid var(--border); border-radius: 4px;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.35); overflow: hidden; color: var(--text);
  }
  .docs-head { display: flex; gap: 0.75rem; align-items: baseline; padding: 0.75rem 1rem 0.25rem; }
  .docs-note { font-size: 0.72rem; color: var(--muted); }
  .docs-filter { display: flex; gap: 0.5rem; padding: 0.5rem 1rem; }
  .docs-filter input, .docs-filter select {
    padding: 0.4rem 0.6rem; border: 1px solid var(--border); border-radius: 4px;
    background: var(--bg); color: var(--text); font: inherit; font-size: 0.85rem;
  }
  .docs-filter input { flex: 1; }
  .docs-table-wrap { overflow: auto; }
  .docs-table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
  .docs-table th {
    position: sticky; top: 0; text-align: left; padding: 0.4rem 0.6rem;
    background: var(--table-head-bg, var(--surface)); border-bottom: 1px solid var(--border);
    font-family: var(--font-mono); font-size: 0.7rem; text-transform: uppercase; letter-spacing: 0.04em;
  }
  .docs-table td { padding: 0.35rem 0.6rem; border-bottom: 1px solid var(--border); vertical-align: top; }
  .docs-table tbody tr { cursor: pointer; }
  .docs-table tbody tr:hover td { background: var(--surface-hover); }
  .mono { font-family: var(--font-mono); white-space: nowrap; }
  .num { text-align: right; font-family: var(--font-mono); }
  .bad { color: #dc2626; font-weight: 600; margin-right: 0.5rem; }
  .warn { color: #d97706; font-weight: 600; }
  .ok { color: var(--muted); }
  .empty { text-align: center; color: var(--muted); padding: 1.5rem; }
  .docs-footer { padding: 0.5rem 1rem; font-size: 0.72rem; color: var(--muted); border-top: 1px solid var(--border); }
  kbd { font-family: var(--font-mono); font-size: 0.7rem; }
</style>
