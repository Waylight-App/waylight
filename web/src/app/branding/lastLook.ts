import type { ThemeMode } from '../../shared/theme/theme';

/**
 * The look of the account that last signed in on this browser (#526), so
 * the sign-in page shows it instead of the instance default. Only a gallery
 * theme's CSS and the light/dark/system choice are kept: no name, no account
 * id, and never someone's own custom CSS.
 */
export interface LastLook {
  css: string;
  theme: ThemeMode;
}

const KEY = 'waylight:sign-in-look';
/** Marker the server writes at the start of a gallery theme's CSS. */
const THEME_MARKER = '/* waylight theme: ';
/** Gallery themes are a few hundred bytes; anything far larger is not one. */
const MAX_CSS = 16 << 10;
const MODES: readonly ThemeMode[] = ['system', 'light', 'dark'];

/**
 * Whether css is a gallery theme as the server writes it. Checked again on
 * read, so an entry someone edited in the browser cannot load anything.
 */
export function isGalleryThemeCss(css: string): boolean {
  return (
    css.startsWith(THEME_MARKER) && css.length <= MAX_CSS && !/url\s*\(|@import|expression\s*\(/i.test(css)
  );
}

/**
 * Keeps the look after the account's appearance has loaded. An account
 * without a gallery theme clears it, so the sign-in page falls back to the
 * instance default instead of showing someone else's theme.
 */
export function rememberLook(look: LastLook, storage: Storage | undefined = globalThis.localStorage): void {
  try {
    if (isGalleryThemeCss(look.css))
      storage?.setItem(KEY, JSON.stringify({ css: look.css, theme: look.theme }));
    else storage?.removeItem(KEY);
  } catch {
    // Storage full or blocked (private mode): the default look stays.
  }
}

/** The kept look, or null when there is none or it does not pass the checks. */
export function lastLook(storage: Storage | undefined = globalThis.localStorage): LastLook | null {
  try {
    const raw = storage?.getItem(KEY);
    if (!raw) return null;
    const v = JSON.parse(raw) as Partial<LastLook>;
    if (typeof v.css !== 'string' || !isGalleryThemeCss(v.css)) return null;
    const theme = MODES.includes(v.theme as ThemeMode) ? (v.theme as ThemeMode) : 'system';
    return { css: v.css, theme };
  } catch {
    return null;
  }
}
