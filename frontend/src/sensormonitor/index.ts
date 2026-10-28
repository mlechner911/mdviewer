// SPDX-License-Identifier: Elastic-2.0
// Copyright (c) 2026 Michael Lechner. Part of the optional Sensor Monitor
// module, licensed under the Elastic License 2.0: see sensormonitor/LICENSE.

// The optional Sensor Monitor module (frontend part). MarkSafe imports it
// as "$sensormonitor": the Sensor Monitor variant resolves that here, the
// standard build to ../sensormonitor-off (see ../../vite.config.ts), so
// none of this is in a public build.
export { default as Extension } from './SensorMonitor.svelte';
export { default as SettingsSection } from './SettingsSection.svelte';
export { texts } from './i18n';

import licenseText from '../../../sensormonitor/LICENSE?raw';
/** The module's license, shown in the About window of this build. */
export const license: { module: string; name: string; text: string } | null = {
  module: 'Sensor Monitor', name: 'Elastic License 2.0', text: licenseText,
};
