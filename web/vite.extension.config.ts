import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

/**
 * Builds the app for the browser extension's local mode (ADR 0019). The
 * output goes to dist-extension/; extension/scripts/app.mjs copies it into
 * the extension before packaging. Unlike the server build it is not gzipped
 * (the browser reads the files from the extension package) and uses relative
 * paths, because extension pages live under a chrome-extension:// or
 * moz-extension:// origin.
 */
export default defineConfig({
  plugins: [svelte()],
  base: './',
  publicDir: false,
  build: {
    outDir: 'dist-extension',
    emptyOutDir: true,
    sourcemap: false,
    rollupOptions: { input: 'extension.html' },
  },
});
