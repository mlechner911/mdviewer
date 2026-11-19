import { fileURLToPath } from 'node:url';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vitest/config';

// CodeMirror is a peer dependency: the copy of the app that runs it
// (frontend/), as in mlcext-editor. mlcext-editor comes as source.
const frontend = (p: string) => fileURLToPath(new URL(`../../frontend/node_modules/${p}`, import.meta.url));

export default defineConfig({
  plugins: [svelte()],
  resolve: {
    conditions: ['browser'],
    alias: [
      { find: /^mlcext-editor$/, replacement: fileURLToPath(new URL('../mlcext-editor/src/index.ts', import.meta.url)) },
      { find: /^@codemirror\/(.*)$/, replacement: frontend('@codemirror/$1') },
      { find: /^@lezer\/(.*)$/, replacement: frontend('@lezer/$1') },
      { find: /^codemirror$/, replacement: frontend('codemirror') },
    ],
  },
  test: { environment: 'jsdom' },
});
