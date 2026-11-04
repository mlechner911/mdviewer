// "$sensormonitor" in the standard build: the module, a private submodule
// at ../sensormonitor, is not built in (see ../../vite.config.ts).
import type { Component } from 'svelte';

export const Extension: Component<any> | null = null;
export const SettingsSection: Component<any> | null = null;
export const texts: Record<string, Record<string, string>> = {};
export const license: { module: string; name: string; text: string } | null = null;
