// The optional Sensor Monitor module (frontend part). MarkSafe imports it
// as "$sensormonitor": the Sensor Monitor variant resolves that here, the
// standard build to ../sensormonitor-off (see ../../vite.config.ts), so
// none of this is in a public build.
export { default as Extension } from './SensorMonitor.svelte';
export { default as SettingsSection } from './SettingsSection.svelte';
export { texts } from './i18n';
