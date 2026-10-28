// "$sensormonitor" in the standard build: the module is not built in (see
// ../sensormonitor/index.ts and ../../vite.config.ts).
import type { Component } from 'svelte';

export const Extension: Component<any> | null = null;
export const SettingsSection: Component<any> | null = null;
export const texts: Record<string, Record<string, string>> = {};
