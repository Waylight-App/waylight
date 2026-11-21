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
  /** A line by time of day after the greeting (#512). */
  greetingMoment: boolean;
  /** What the launcher lists (#189). */
  launcher: 'both' | 'services' | 'bookmarks';
  /** What the start page shows (#281): everyday topics, or also the home network. */
  profile: StartProfile;
  /** Lists settings for developers and power users (#602); off by default. */
  showAdvanced?: boolean;
  /** Scales the whole interface, text and spacing (#681). */
  textSize?: TextSize;
  /** High-contrast colours of the light or dark scheme (#681). */
  highContrast?: boolean;
  /** One column of compact views without edit controls (#681). */
  simpleView?: boolean;
}

/** Text sizes of the accessibility options (#681). */
export const TEXT_SIZES = ['normal', 'large', 'larger'] as const;
export type TextSize = (typeof TEXT_SIZES)[number];

/** Parses a text size from storage or the API; unknown values are normal. */
export function parseTextSize(value: string | undefined): TextSize {
  return TEXT_SIZES.find((s) => s === value) ?? 'normal';
}

/**
 * Data attributes of the document root for the accessibility options
 * (#681); tokens.css and base.css style them. An attribute that is not
 * needed is undefined, so it is removed.
 */
export function accessibilityAttributes(a: Appearance): Record<string, string | undefined> {
  const size = parseTextSize(a.textSize);
  return {
    'data-text-size': size === 'normal' ? undefined : size,
    'data-contrast': a.highContrast ? 'high' : undefined,
    'data-simple': a.simpleView ? 'true' : undefined,
  };
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
  greetingMoment: false,
  launcher: 'both',
  profile: 'homelab',
};

/**
 * CSS custom properties derived from the preferences. Empty values are
 * omitted so the theme defaults apply.
 */
export function cssVariables(a: Appearance): Record<string, string> {
  const vars: Record<string, string> = {};
  // High contrast has its own checked colours; an own accent colour or a
  // picture behind translucent cards would lower the contrast again (#681).
  if (a.highContrast) return vars;
  if (/^#[0-9a-f]{6}$/i.test(a.accentColor)) vars['--color-accent'] = a.accentColor;
  const bg = backgroundCss(a.backgroundImage);
  if (bg) {
    vars['--background-image'] = bg;
    // Slightly translucent cards keep the background visible.
    vars['--widget-opacity'] = '0.94';
  }
  return vars;
}
