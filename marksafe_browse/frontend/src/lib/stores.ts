import { writable, derived } from 'svelte/store';
import type { Writable } from 'svelte/store';

// Theme store
export const appTheme: Writable<string> = writable('auto');
export const effectiveTheme: Writable<string> = writable('dark');

// UI state stores
export const isFocusMode: Writable<boolean> = writable(false);
export const isEditorHidden: Writable<boolean> = writable(false);
export const isPrinting: Writable<boolean> = writable(false);
export const splitWidth: Writable<number> = writable(50);
export const fontSize: Writable<number> = writable(100);

// Content stores
export const tabs: Writable<any[]> = writable([]);
export const activeTabIndex: Writable<number> = writable(0);
export const htmlContent: Writable<string> = writable('');
export const isReady: Writable<boolean> = writable(false);

// Drop/Toast stores
export const dropMessage: Writable<string> = writable('');
export const showToast: Writable<boolean> = writable(false);
export const toastType: Writable<string> = writable('info');

// Version store
export const appVersion: Writable<string> = writable('0.0.0');

// Tab interface
export interface Tab {
  id: string;
  title: string;
  path: string | null;
  content: string;
  isDirty: boolean;
}

// Derived: word count from active tab
export const wordCount = derived([tabs, activeTabIndex], ([$tabs, $activeTabIndex]) => {
  const active = $tabs[$activeTabIndex] || null;
  return active ? (active.content?.trim().split(/\s+/).filter(Boolean).length || 0) : 0;
});

export const charCount = derived([tabs, activeTabIndex], ([$tabs, $activeTabIndex]) => {
  const active = $tabs[$activeTabIndex] || null;
  return active ? (active.content?.length || 0) : 0;
});

export const readingTime = derived(wordCount, ($wc) => Math.ceil($wc / 225));
