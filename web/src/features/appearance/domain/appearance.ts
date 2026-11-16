import { backgroundCss } from '../../../shared/theme/background';
import type { ThemeMode } from '../../../shared/theme/theme';

/** Visual preferences of the start page. */
export interface Appearance {
  theme: ThemeMode;
  /**
   * The organisation keeps members on its theme and accent colour (#305);
   * the server ignores both on save, so the settings show them read-only.
   */
  themeLocked?: boolean;
  accentColor: string;
  backgroundImage: string;
  customCss: string;
  /** UI language; "" follows the browser. */
  language: '' | 'en' | 'de';
  /** Name in the greeting of the header (#189); may be empty. */
  greetingName: string;
  /** What the launcher lists (#189). */
  launcher: 'both' | 'services' | 'bookmarks';
  /** What the start page shows (#281): everyday topics, or also the home network. */
  profile: StartProfile;
}

/** Start-page profiles (#281). */
export type StartProfile = 'everyday' | 'homelab';

/** Widget categories that the "everyday" profile hides (#281). */
export const HOMELAB_CATEGORIES: readonly string[] = ['server', 'network'];

/** Whether a widget category is shown for a profile. */
export const categoryShown = (profile: StartProfile, category: string): boolean =>
  profile === 'homelab' || !HOMELAB_CATEGORIES.includes(category);

/** Defaults before the preferences are loaded. */
export const defaultAppearance: Appearance = {
  theme: 'system',
  accentColor: '',
  backgroundImage: '',
  customCss: '',
  language: '',
  greetingName: '',
  launcher: 'both',
  profile: 'homelab',
};

/**
 * CSS custom properties derived from the preferences. Empty values are
 * omitted so the theme defaults apply.
 */
export function cssVariables(a: Appearance): Record<string, string> {
  const vars: Record<string, string> = {};
  if (/^#[0-9a-f]{6}$/i.test(a.accentColor)) vars['--color-accent'] = a.accentColor;
  const bg = backgroundCss(a.backgroundImage);
  if (bg) {
    vars['--background-image'] = bg;
    // Slightly translucent cards keep the background visible.
    vars['--widget-opacity'] = '0.94';
  }
  return vars;
}
