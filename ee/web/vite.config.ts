import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vitest/config';

/**
 * The ee/ frontend as one ES module (ADR 0021). The server embeds dist/ in
 * builds with the tag `ee` and serves it under /ee/; the web app imports
 * /ee/plugin.js when GET /api/plugins lists it. Svelte is bundled in, so the
 * plugin does not depend on the core's runtime build.
 */
export default defineConfig({
  plugins: [svelte()],
  // Lazy chunks are loaded relative to plugin.js.
  base: './',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
    lib: {
      entry: 'src/plugin.ts',
      formats: ['es'],
      fileName: () => 'plugin.js',
    },
  },
  test: {
    include: ['src/**/*.test.ts'],
    environment: 'node',
  },
});
