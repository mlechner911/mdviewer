import { writable } from 'svelte/store';
import { APP_THEME } from './constants';
import type { AppTheme_t, EffectiveTheme_t } from './constants';

const savedTheme = (typeof localStorage !== 'undefined' ? localStorage.getItem('marksafe_theme') : null) as AppTheme_t | null;
const initialTheme: AppTheme_t = savedTheme && ['dark', 'light', 'auto'].includes(savedTheme) ? savedTheme : APP_THEME.AUTO;

// App Theme Store (dark, light, auto)
export const appTheme = writable<AppTheme_t>(initialTheme);

appTheme.subscribe(val => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('marksafe_theme', val);
  }
});

// Effective Theme (resolves 'auto' to 'dark' or 'light')
export const effectiveAppTheme = writable<EffectiveTheme_t>('dark');

// Layout
export const splitWidth = writable(50);
export const isFocusMode = writable(false);
export const isEditorHidden = writable(false);
export const isPrinting = writable(false);
export const menuVisible = writable(false);

// Global UI feedback
export const dropMessage = writable<string | null>(null);
export const toastType = writable<'info' | 'error'>('info');

export function showToast(message: string, duration = 3000, type: 'info' | 'error' = 'info') {
  toastType.set(type);
  dropMessage.set(message);
  setTimeout(() => dropMessage.set(null), duration);
}

// Application version (fetched from Go backend on startup, falls back to default)
export const appVersion = writable('1.3.1');
