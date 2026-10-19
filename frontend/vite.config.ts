import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'
import wails from '@wailsio/runtime/plugins/vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte(), wails('./bindings')],
  resolve: {
    // ../packages/mlcext-editor imports CodeMirror; one instance for the app
    // and the package (two copies break CodeMirror's instanceof checks).
    dedupe: ['@codemirror/state', '@codemirror/view', '@codemirror/language', '@codemirror/lint', '@codemirror/autocomplete'],
  },
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 34116,
    strictPort: true,
    fs: { allow: ['..'] },
  }
})
