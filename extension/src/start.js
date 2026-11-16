// What a new tab opens (ADR 0019): the user's server, the app bundled for
// the local mode, or a page that lets them choose. Kept free of browser
// APIs so it can be tested with node --test.

/** Path of the bundled app inside the extension. */
export const LOCAL_APP = 'app/extension.html';

/**
 * @param {{ base: string, mode: string }} settings stored settings
 * @returns {'server' | 'local' | 'choose'}
 */
export function startTarget({ base, mode }) {
  if (mode === 'local') return 'local';
  // Installs from before the local mode only stored an address.
  if (base) return 'server';
  return 'choose';
}
