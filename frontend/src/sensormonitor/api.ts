// SPDX-License-Identifier: Elastic-2.0
// Copyright (c) 2026 Michael Lechner. Part of the optional Sensor Monitor
// module, licensed under the Elastic License 2.0: see sensormonitor/LICENSE.

// Calls into the Go service marksafe/sensormonitor.Service by name, so
// this module does not depend on generated bindings (it is built only into
// the Sensor Monitor variant, see ../../vite.config.ts).
import { Call } from '@wailsio/runtime';
import { isWailsReady } from '../lib/backend';

const SERVICE = 'marksafe/sensormonitor.Service.';

function call<T>(method: string, ...args: any[]): Promise<T> {
  return Call.ByName(SERVICE + method, ...args) as Promise<T>;
}

export type Status = { address: string; signedIn: boolean; email: string; expires: string };
export type LoginCode = { userCode: string; url: string; expiresIn: number };
export type SensorError = { code: string; message: string; params: Record<string, string>; signIn: boolean };
export type SensorDocument = {
  id: string; doc_no: string; ref_no: string; version: string; title: string; filename: string; status: string;
  approved_at: string; approved_by_label: string; uploaded_by_label: string; valid_until: string; created_at: string;
};
export type Stand = { ref_no: string; approved: SensorDocument | null; draft: SensorDocument | null };
export type Comparison = Stand & { matches_approved: boolean; matches_draft: boolean };
export type Finding = { line: number; col: number; severity: string; code: string; message: string };
export type SavedDraft = SensorDocument & { warnings: Finding[] | null };

const OFF: Status = { address: '', signedIn: false, email: '', expires: '' };

function failure(err: any): SensorError {
  return { code: '', message: String(err?.message ?? err), params: {}, signIn: false };
}

export async function status(): Promise<Status> {
  if (!isWailsReady()) return OFF;
  try { return (await call<Status>('GetStatus')) ?? OFF; } catch { return OFF; }
}

/** Stores the address ("" turns the connection off); returns it normalized. Throws with the reason. */
export async function setAddress(address: string): Promise<string> {
  if (!isWailsReady()) return '';
  return (await call<string>('SetAddress', address)) ?? '';
}

/** Starts the device login; the result comes as the event "sensormonitor-login". Throws with the reason. */
export async function startLogin(): Promise<LoginCode | null> {
  if (!isWailsReady()) return null;
  return (await call<LoginCode>('StartLogin')) ?? null;
}

export async function cancelLogin(): Promise<void> {
  if (!isWailsReady()) return;
  try { await call('CancelLogin'); } catch { /* ignore */ }
}

export async function lookup(ref: string): Promise<{ stand: Stand | null; error: SensorError | null }> {
  if (!isWailsReady()) return { stand: null, error: failure('Wails not ready') };
  try { return await call('Lookup', ref); } catch (err) { return { stand: null, error: failure(err) }; }
}

export async function source(id: string): Promise<{ text: string; error: SensorError | null }> {
  if (!isWailsReady()) return { text: '', error: failure('Wails not ready') };
  try { return await call('Source', id); } catch (err) { return { text: '', error: failure(err) }; }
}

export async function compare(ref: string, md: string): Promise<{ comparison: Comparison | null; error: SensorError | null }> {
  if (!isWailsReady()) return { comparison: null, error: failure('Wails not ready') };
  try { return await call('Compare', ref, md); } catch (err) { return { comparison: null, error: failure(err) }; }
}

export async function saveDraft(md: string, filename: string, reason: string, pin: string): Promise<{ draft: SavedDraft | null; error: SensorError | null }> {
  if (!isWailsReady()) return { draft: null, error: failure('Wails not ready') };
  try { return await call('SaveDraft', md, filename, reason, pin); } catch (err) { return { draft: null, error: failure(err) }; }
}
