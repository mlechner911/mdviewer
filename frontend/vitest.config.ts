// Unit tests of the frontend logic (src/**/*.test.ts). Same aliases as the
// app (vite.config.ts), without the Wails plugin: tests mock the backend.
import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import viteConfig from './vite.config.ts';

export default defineConfig({
  plugins: [svelte()],
  resolve: viteConfig.resolve,
  test: { include: ['src/**/*.test.ts'] },
});
