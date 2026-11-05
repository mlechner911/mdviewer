import { fileURLToPath } from 'node:url';

// CodeMirror is a peer dependency: the host's copy, the one the app runs.
// A plain object: this package does not install vitest, `task check` runs it
// from packages/mlc-markdown.
export default {
  resolve: {
    alias: [{ find: /^@codemirror\/(.*)$/, replacement: fileURLToPath(new URL('../../frontend/node_modules/@codemirror/$1', import.meta.url)) }],
  },
  test: { environment: 'jsdom' },
};
