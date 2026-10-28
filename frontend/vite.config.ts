import {defineConfig} from 'vite'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import {svelte} from '@sveltejs/vite-plugin-svelte'
import wails from '@wailsio/runtime/plugins/vite'

// Build variant from .variant (written by the Taskfile, "standard" when
// missing): "sensormonitor" builds the optional Sensor Monitor module in,
// otherwise "$sensormonitor" resolves to an empty stand-in and none of the
// module's code is in the bundle. A file, not an environment variable, so
// that Task's fingerprint of frontend/ sees the change.
const variant = (() => {
  try { return readFileSync(new URL('./.variant', import.meta.url), 'utf8').trim().split(',') } catch { return ['standard'] }
})()
const sensormonitor = variant.includes('sensormonitor')

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte(), wails('./bindings')],
  resolve: {
    // Shared packages from ../packages, used from source. One instance of
    // their dependencies for the app and the packages (two copies break
    // CodeMirror's instanceof checks and Svelte's runtime).
    alias: [
      { find: /^mlc-markdown\/styles\.css$/, replacement: fileURLToPath(new URL('../packages/mlc-markdown/styles/mlc-markdown.css', import.meta.url)) },
      { find: /^mlc-markdown$/, replacement: fileURLToPath(new URL('../packages/mlc-markdown/src/index.ts', import.meta.url)) },
      { find: /^\$sensormonitor$/, replacement: fileURLToPath(new URL(sensormonitor ? './src/sensormonitor/index.ts' : './src/sensormonitor-off/index.ts', import.meta.url)) },
      { find: /^mlc-markdown-svelte$/, replacement: fileURLToPath(new URL('../packages/mlc-markdown-svelte/src/index.js', import.meta.url)) },
    ],
    dedupe: ['@codemirror/state', '@codemirror/view', '@codemirror/language', '@codemirror/lint', '@codemirror/autocomplete', 'svelte', 'marked', 'dompurify', 'marked-footnote'],
  },
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 34116,
    strictPort: true,
    fs: { allow: ['..'] },
  }
})
