import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/**
 * The plugin is loaded with a plain import(), which does not load CSS files,
 * so component styles are injected by the JavaScript itself (ADR 0021).
 */
/** @type {import('@sveltejs/vite-plugin-svelte').SvelteConfig} */
export default {
  preprocess: vitePreprocess(),
  compilerOptions: { runes: true, css: 'injected' },
};
