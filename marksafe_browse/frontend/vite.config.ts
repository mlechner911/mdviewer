import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

export default defineConfig({
  // NOTE: spread the array svelte() returns — svelte-check only finds the
  // plugin by name in a flat plugins list, otherwise all diagnostics break.
  plugins: [...svelte({
    compilerOptions: {
      generate: 'dom',
    },
  })],
  define: {
    'process.env.NODE_ENV': JSON.stringify('production'),
    'global.process': 'undefined',
    'globalThis.process': 'undefined',
  },
  resolve: {
    conditions: ['browser'],
  },
  build: {
    outDir: '../internal/browse/frontend/dist',
    emptyOutDir: true,
    rollupOptions: {
      // Content hashes: every build gets unique filenames, so browsers
      // never serve a stale bundle from cache. The Go server discovers
      // the current names from the embedded assets dir at startup.
      output: {
        entryFileNames: 'assets/index-[hash].js',
        chunkFileNames: 'assets/chunk-[hash].js',
        assetFileNames: 'assets/[name]-[hash][extname]',
      },
    },
  },
  server: {
    port: 5173,
  },
});
