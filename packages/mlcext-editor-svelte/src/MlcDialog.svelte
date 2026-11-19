<script lang="ts">
  // Frame of the editor's dialogs: overlay, panel, title, buttons. No
  // Tailwind: colours are CSS variables --mlc-dialog-* with MarkSafe's
  // slate palette as fallback, so a host restyles them from outside.
  import type { Snippet } from 'svelte';

  let {
    title,
    theme = 'light',
    width = '32rem',
    submitLabel,
    cancelLabel,
    submitDisabled = false,
    onSubmit,
    onClose,
    children,
  }: {
    title: string;
    theme?: 'light' | 'dark';
    width?: string;
    submitLabel: string;
    cancelLabel: string;
    submitDisabled?: boolean;
    onSubmit: () => void;
    onClose: () => void;
    children: Snippet;
  } = $props();
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
  class="mlc-dialog-overlay"
  role="dialog"
  aria-modal="true"
  aria-label={title}
  tabindex="-1"
  onkeydown={(e) => {
    if (e.key === 'Escape') onClose();
  }}
>
  <div class="mlc-dialog mlc-dialog-{theme === 'dark' ? 'dark' : 'light'}" style:max-width={width}>
    <form
      onsubmit={(e) => {
        e.preventDefault();
        onSubmit();
      }}
    >
      <h3>{title}</h3>
      {@render children()}
      <div class="mlc-dialog-buttons">
        <button type="button" class="mlc-d-button" onclick={onClose}>{cancelLabel}</button>
        <button type="submit" class="mlc-d-button mlc-d-primary" disabled={submitDisabled}>{submitLabel}</button>
      </div>
    </form>
  </div>
</div>

<style>
  .mlc-dialog-overlay {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    background: var(--mlc-dialog-overlay, rgba(0, 0, 0, 0.6));
    backdrop-filter: blur(4px);
  }
  .mlc-dialog {
    --fallback-bg: #f1f5f9;
    --fallback-fg: #0f172a;
    --fallback-border: #cbd5e1;
    --fallback-button: #e2e8f0;
    --fallback-button-hover: #cbd5e1;
    --fallback-input-bg: #ffffff;
    --fallback-input-border: #d1d5db;
    width: 100%;
    overflow: hidden;
    border: 1px solid var(--mlc-dialog-border, var(--fallback-border));
    border-radius: 0.5rem;
    background: var(--mlc-dialog-bg, var(--fallback-bg));
    color: var(--mlc-dialog-fg, var(--fallback-fg));
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    font-family: var(--mlc-dialog-font, inherit);
  }
  .mlc-dialog-dark {
    --fallback-bg: #1e293b;
    --fallback-fg: #f1f5f9;
    --fallback-border: #334155;
    --fallback-button: #334155;
    --fallback-button-hover: #475569;
    --fallback-input-bg: #1e293b;
    --fallback-input-border: #475569;
  }
  form {
    padding: 1.5rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
  }
  h3 {
    margin: 0 0 1rem;
    font-size: 1.125rem;
    line-height: 1.75rem;
    font-weight: 700;
  }
  .mlc-dialog-buttons {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    margin-top: 1.5rem;
  }

  /* Building blocks of the dialog bodies (MlcFieldDialog, MlcImageDialog). */
  .mlc-dialog :global(.mlc-d-grid) {
    display: grid;
    grid-template-columns: max-content 1fr;
    column-gap: 0.75rem;
    row-gap: 0.5rem;
    align-items: center;
  }
  .mlc-dialog :global(.mlc-d-label) {
    opacity: 0.8;
  }
  .mlc-dialog :global(.mlc-d-note) {
    margin: 0.25rem 0 0;
    font-size: 0.75rem;
    opacity: 0.7;
  }
  .mlc-dialog :global(.mlc-d-warn) {
    margin: 0.25rem 0 0;
    font-size: 0.75rem;
    color: var(--mlc-dialog-warn, #f59e0b);
  }
  .mlc-dialog :global(.mlc-d-error) {
    margin: 0.25rem 0 0;
    font-size: 0.75rem;
    color: var(--mlc-dialog-error, #ef4444);
  }
  .mlc-dialog :global(.mlc-d-heading) {
    margin-top: 1.25rem;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.025em;
    opacity: 0.6;
  }
  .mlc-dialog :global(.mlc-d-row) {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
  }
  .mlc-dialog :global(.mlc-d-check) {
    display: flex;
    align-items: center;
    gap: 0.375rem;
  }
  /* :where keeps this rule weak, so .mlc-d-mono and a host's classes win. */
  .mlc-dialog :global(:where(input:not([type='checkbox']):not([type='radio']), select, textarea, .mlc-d-box)) {
    box-sizing: border-box;
    padding: 0.25rem 0.5rem;
    border: 1px solid var(--mlc-dialog-input-border, var(--fallback-input-border));
    border-radius: 0.25rem;
    background: var(--mlc-dialog-input-bg, var(--fallback-input-bg));
    color: inherit;
    font: inherit;
  }
  .mlc-dialog :global([aria-invalid='true']) {
    border-color: var(--mlc-dialog-error, #ef4444);
  }
  .mlc-dialog :global(.mlc-d-mono) {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
  .mlc-dialog :global(.mlc-d-button) {
    padding: 0.5rem 1rem;
    border: 0;
    border-radius: 0.25rem;
    background: var(--mlc-dialog-button, var(--fallback-button));
    color: inherit;
    font: inherit;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.15s;
  }
  .mlc-dialog :global(.mlc-d-button.mlc-d-small) {
    padding: 0.25rem 0.75rem;
    flex-shrink: 0;
  }
  .mlc-dialog :global(.mlc-d-button:hover) {
    background: var(--mlc-dialog-button-hover, var(--fallback-button-hover));
  }
  .mlc-dialog :global(.mlc-d-primary) {
    background: var(--mlc-dialog-primary, #2563eb);
    color: #ffffff;
  }
  .mlc-dialog :global(.mlc-d-primary:hover) {
    background: var(--mlc-dialog-primary-hover, #3b82f6);
  }
  .mlc-dialog :global(button:disabled) {
    opacity: 0.5;
    cursor: default;
  }
</style>
