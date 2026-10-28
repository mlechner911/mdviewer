// SPDX-License-Identifier: Elastic-2.0
// Copyright (c) 2026 Michael Lechner. Part of the optional Sensor Monitor
// module, licensed under the Elastic License 2.0: see sensormonitor/LICENSE.

// Refusals of the Sensor Monitor, translated by their fixed code; the
// server's own text is only an English fallback.
import type { SensorError } from './api';

type T = (key: string, ...args: any[]) => string;

const KEYS: Record<string, string> = {
  'not-signed-in': 'smErrNotSignedIn',
  unauthorized: 'smErrNotSignedIn',
  ERR_UNAUTHENTICATED: 'smErrNotSignedIn',
  forbidden: 'smErrServiceForbidden', // mlcauth: service access not granted
  ERR_FORBIDDEN: 'smErrForbidden',
  ERR_DOCUMENT_NOT_FOUND: 'smErrNotFound',
  ERR_DOCUMENT_UNCHANGED: 'smErrUnchanged',
  ERR_DOCUMENT_ID_REQUIRED: 'smErrIdRequired',
  ERR_REASON_REQUIRED: 'smErrReasonRequired',
  ERR_SIGN_PIN_REQUIRED: 'smErrPinRequired',
  ERR_SIGN_PIN_INVALID: 'smErrPinInvalid',
  ERR_SIGN_PIN_MISSING: 'smErrPinMissing',
  ERR_SIGN_PIN_LOCKED: 'smErrPinLocked',
  ERR_SIGN_PIN_AFTER_IDLE: 'smErrPinIdle',
  ERR_DOCUMENT_TOO_LARGE: 'smErrTooLarge',
};

/** The refusal as a sentence for the user. */
export function sensorErrorText(e: SensorError, t: T): string {
  switch (e.code) {
    case 'ERR_DOCUMENT_DRAFT_PENDING':
      return t('smErrDraftPending', e.params?.doc_no || '–');
    case 'ERR_MARKDOWN_INVALID':
      return t('smErrInvalid', e.params?.findings || e.message);
  }
  const key = KEYS[e.code];
  return key ? t(key) : t('smErrOther', e.message || e.code);
}

/** The refusal asks for the signing PIN (show the PIN field). */
export function needsPin(e: SensorError | null): boolean {
  return !!e && ['ERR_SIGN_PIN_REQUIRED', 'ERR_SIGN_PIN_INVALID', 'ERR_SIGN_PIN_AFTER_IDLE'].includes(e.code);
}
