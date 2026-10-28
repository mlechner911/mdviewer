<!--
  SPDX-License-Identifier: Elastic-2.0
  Copyright (c) 2026 Michael Lechner. Part of the optional Sensor Monitor
  module, licensed under the Elastic License 2.0: see sensormonitor/LICENSE.
-->
<script lang="ts">
  /**
   * The Sensor Monitor module in the frontend: sign-in, open, save as
   * draft, comparison with the approved version and the watermark
   * suggestion. MarkSafe mounts it only in the Sensor Monitor variant (see
   * index.ts and ../../vite.config.ts) and talks to it through props and
   * suggestWatermark(); everything else is in here.
   *
   * Off until an address is set (File › Settings…): then nothing of it is
   * shown.
   */
  import { onMount, untrack } from 'svelte';
  import { t } from '../i18n';
  import type { EffectiveTheme_t } from '../lib/constants';
  import { showToast } from '../lib/stores';
  import { EventsOn } from '../lib/wails';
  import ConfirmModal from '../components/ConfirmModal.svelte';
  import LoginModal from './LoginModal.svelte';
  import OpenModal from './OpenModal.svelte';
  import DraftModal from './DraftModal.svelte';
  import * as api from './api';

  export type StatusItem = { text: string; tone?: 'ok' | 'warn'; title?: string };

  let {
    markdown,
    doc = null,
    hasErrors = false,
    filename = '',
    theme = 'dark',
    onOpenTab,
    onStatus,
    onMenuChanged,
  } = $props<{
    /** Text of the active tab. */
    markdown: string;
    /** How mlcext reads the active document. */
    doc?: { refNo: string; version: string; title: string } | null;
    /** The local check reports errors (the server would refuse). */
    hasErrors?: boolean;
    /** File name of the active tab, "" when unsaved. */
    filename?: string;
    theme?: EffectiveTheme_t;
    onOpenTab: (title: string, text: string) => void;
    /** Entries for the status bar. */
    onStatus: (items: StatusItem[]) => void;
    /** Address or sign-in changed: rebuild the native menu. */
    onMenuChanged: () => void;
  }>();

  let status = $state<api.Status>({ address: '', signedIn: false, email: '', expires: '' });
  async function refresh() {
    status = await api.status();
    onMenuChanged(); // the "Sensor Monitor" submenu follows address and sign-in
  }

  // Actions need a sign-in: without one, sign in first and run the action
  // afterwards.
  let showLogin = $state(false);
  let showOpen = $state(false);
  let showDraft = $state(false);
  let afterLogin: (() => void) | null = null;
  function requireSignIn(action: () => void) {
    if (status.signedIn) { action(); return; }
    afterLogin = action;
    showLogin = true;
  }
  function loginDone() {
    showLogin = false;
    const next = afterLogin;
    afterLogin = null;
    void refresh().then(() => next?.());
  }

  function opened(text: string, d: api.SensorDocument) {
    showOpen = false;
    const title = [d.ref_no, d.version && `v${d.version}`].filter(Boolean).join(' ') || d.filename || $t('untitled');
    onOpenTab(title, text);
    showToast($t('smOpened', title), 3000, 'info');
  }

  function draftSaved(d: api.SavedDraft) {
    showDraft = false;
    const n = d.warnings?.length ?? 0;
    showToast($t('smDraftSaved', d.doc_no, d.version || '–') + (n ? ' · ' + $t('smDraftWarnings', n) : ''), n ? 6000 : 3500, 'info');
    comparison = null; // compare again
    scheduleCompare(0);
  }

  // Comparison (signed in, document with doc_id): does the working copy
  // match the approved version? Shown in the status bar.
  let comparison = $state<{ ref: string; content: string; result: api.Comparison } | null>(null);
  let compareTimer: ReturnType<typeof setTimeout> | undefined;
  const ref = $derived(status.signedIn ? (doc?.refNo ?? '') : '');
  function scheduleCompare(delay = 1500) {
    clearTimeout(compareTimer);
    compareTimer = setTimeout(async () => {
      const r = ref;
      const content = markdown;
      if (!r) { comparison = null; return; }
      const res = await api.compare(r, content);
      if (res.comparison && r === ref) comparison = { ref: r, content, result: res.comparison };
      else if (res.error) comparison = null;
    }, delay);
  }
  $effect(() => {
    // Again when the id, the text or the sign-in changes.
    void ref; void markdown;
    untrack(() => scheduleCompare());
  });

  // Status bar: the comparison and until when signed in.
  $effect(() => {
    const items: StatusItem[] = [];
    const c = comparison;
    if (c && c.ref === ref) {
      const r = c.result;
      if (r.matches_approved && r.approved) items.push({ text: $t('smCmpApproved', c.ref, r.approved.version), tone: 'ok' });
      else if (r.matches_draft && r.draft) items.push({ text: $t('smCmpDraft', c.ref, r.draft.version), tone: 'warn' });
      else if (r.approved) items.push({ text: $t('smCmpDiffers', c.ref, r.approved.version), tone: 'warn' });
      else if (r.draft) items.push({ text: `${c.ref}: ${$t('smDraft')} ${r.draft.version}`, tone: 'warn' });
      else items.push({ text: $t('smCmpNew', c.ref), tone: 'warn' });
    }
    if (status.signedIn) {
      const until = status.expires ? new Date(status.expires).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';
      items.push({ text: until ? $t('smStatus', until) : $t('smSignedIn'), tone: 'ok', title: status.email });
    }
    untrack(() => onStatus(items));
  });

  // Watermark suggestion before printing or exporting a PDF: ENTWURF for a
  // document that does not match the approved version. Only a suggestion
  // (MarkSafe is not only for the Sensor Monitor), and never when the user
  // set a watermark. "" keeps the setting.
  let wmAsk = $state<((yes: boolean) => void) | null>(null);
  export async function suggestWatermark(ownWatermark: string): Promise<string> {
    const c = comparison;
    if (ownWatermark || !c || c.ref !== ref || c.content !== markdown || c.result.matches_approved) return '';
    const yes = await new Promise<boolean>((resolve) => (wmAsk = resolve));
    return yes ? 'ENTWURF' : '';
  }
  function answerWatermark(yes: boolean) {
    wmAsk?.(yes);
    wmAsk = null;
  }

  onMount(() => {
    const offs = [
      EventsOn('menu-sm-login', () => (showLogin = true)),
      EventsOn('menu-sm-open', () => requireSignIn(() => (showOpen = true))),
      EventsOn('menu-sm-save-draft', () => requireSignIn(() => (showDraft = true))),
      EventsOn('sensormonitor-login', (r: any) => {
        if (r?.ok) showToast($t('smSignedIn'));
        else if (r?.error === 'signed-out') showToast($t('smSignedOut'));
        void refresh();
      }),
    ];
    const changed = () => void refresh();
    window.addEventListener('sensormonitor-changed', changed);
    // The sign-in runs out after an hour (no refresh): then menu and status
    // bar go back to "not signed in".
    const timer = setInterval(() => {
      if (status.expires && Date.parse(status.expires) <= Date.now()) void refresh();
    }, 30_000);
    void refresh();
    return () => {
      offs.forEach((off) => off());
      window.removeEventListener('sensormonitor-changed', changed);
      clearInterval(timer);
      clearTimeout(compareTimer);
    };
  });
</script>

{#if showLogin}
  <LoginModal {theme} onDone={loginDone} onClose={() => { showLogin = false; afterLogin = null; }} />
{/if}
{#if showOpen}
  <OpenModal
    initialRef={doc?.refNo ?? ''}
    {theme}
    onOpen={opened}
    onSignIn={() => { showOpen = false; requireSignIn(() => (showOpen = true)); }}
    onClose={() => (showOpen = false)}
  />
{/if}
{#if showDraft}
  <DraftModal
    {markdown}
    filename={filename || (doc?.refNo ? `${doc.refNo}.md` : '')}
    refNo={doc?.refNo ?? ''}
    version={doc?.version ?? ''}
    title={doc?.title ?? ''}
    {hasErrors}
    {theme}
    onSaved={draftSaved}
    onSignIn={() => { showDraft = false; requireSignIn(() => (showDraft = true)); }}
    onClose={() => (showDraft = false)}
  />
{/if}
{#if wmAsk}
  <ConfirmModal
    title={$t('smWmTitle')}
    message={$t('smWmMessage')}
    confirmLabel={$t('smWmYes')}
    cancelLabel={$t('smWmNo')}
    confirmClass="bg-blue-600 hover:bg-blue-500"
    {theme}
    onConfirm={() => answerWatermark(true)}
    onCancel={() => answerWatermark(false)}
  />
{/if}
