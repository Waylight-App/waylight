import { writeFileSync } from 'node:fs';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import type { Plugin } from 'vite';
import { defineConfig } from 'vitest/config';

/**
 * dist/.gitkeep keeps the directory in git, so `go build -tags ee` finds
 * something to embed without an npm build (ee/web/embed.go). Vite empties the
 * directory, so the file is written again after each build.
 */
function keepDist(): Plugin {
  return {
    name: 'keep-dist',
    apply: 'build',
    closeBundle() {
      writeFileSync('dist/.gitkeep', '');
    },
  };
}

/**
 * The ee/ frontend as one ES module (ADR 0021). The server embeds dist/ in
 * builds with the tag `ee` and serves it under /ee/; the web app imports
 * /ee/plugin.js when GET /api/plugins lists it. Svelte is bundled in, so the
 * plugin does not depend on the core's runtime build.
 */
export default defineConfig({
  plugins: [svelte(), keepDist()],
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
    rolldownOptions: {
      // The server caches assets/* forever and plugin.js never (#441), so
      // lazy chunks carry a content hash and live under assets/.
      output: { chunkFileNames: 'assets/[name]-[hash].js', assetFileNames: 'assets/[name]-[hash][extname]' },
    },
  },
  test: {
    include: ['src/**/*.test.ts'],
    environment: 'node',
  },
});
