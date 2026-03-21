import { writable, derived } from 'svelte/store';
import { APP_THEME } from './constants';
import type { AppTheme_t } from './constants';

// App Theme Store
export const appTheme = writable<AppTheme_t>(APP_THEME.DARK);

// Effective Theme (resolves 'auto' to 'dark' or 'light')
export const effectiveAppTheme = writable<'dark' | 'light'>('dark');

// Layout
export const splitWidth = writable(50);
export const isFocusMode = writable(false);
export const isEditorHidden = writable(false);
export const isPrinting = writable(false);

// Global UI feedback
export const dropMessage = writable<string | null>(null);

export function showToast(message: string, duration = 3000) {
    dropMessage.set(message);
    setTimeout(() => dropMessage.set(null), duration);
}
