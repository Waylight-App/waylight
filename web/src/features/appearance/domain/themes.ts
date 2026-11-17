/** Light or dark: a theme's single scheme, or one of its variants. */
export type ThemeSide = 'light' | 'dark';

/** A theme of the gallery (#111): design tokens only. */
export interface ThemeInfo {
  id: string;
  name: string;
  author: string;
  /** The scheme of a single-scheme theme; "" for a theme with variants. */
  scheme: ThemeSide | '';
  /** Tokens for every scheme; in a theme with variants, the shared ones. */
  tokens: Record<string, string>;
  /** Light and dark tokens of a theme that has both (#505). */
  variants?: Record<ThemeSide, Record<string, string>>;
  builtIn: boolean;
}

/** The theme gallery. */
export interface ThemeGallery {
  themes: ThemeInfo[];
  /** Id of the applied theme, "" for own or no custom CSS. */
  applied: string;
  canRestore: boolean;
}

/** Largest theme file accepted for import. */
export const MAX_THEME_BYTES = 16 << 10;

/**
 * Font keywords a theme may use for --font-sans and --font-display, and the
 * stacks they stand for. Mirrors FontStacks in
 * internal/modules/appearance/domain/theme.go; a Go test keeps both equal.
 */
export const FONT_STACKS: Record<string, string> = {
  system: "system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif",
  figtree: "'Figtree Variable', Figtree, system-ui, sans-serif",
  fraunces: "'Fraunces Variable', Fraunces, Georgia, serif",
  serif: "Georgia, 'Times New Roman', serif",
};

const safeValue = /^(#[0-9a-f]{3}|#[0-9a-f]{6}|#[0-9a-f]{8}|\d{1,2}px)$/i;

/** The sides a gallery card previews: both for a theme with variants. */
export function themeSides(theme: Pick<ThemeInfo, 'scheme' | 'variants'>): ThemeSide[] {
  if (theme.variants) return ['light', 'dark'];
  return [theme.scheme === 'dark' ? 'dark' : 'light'];
}

/** The tokens in effect for one side: shared tokens, the variant on top. */
export function resolvedTokens(
  theme: Pick<ThemeInfo, 'tokens' | 'variants'>,
  side: ThemeSide,
): Record<string, string> {
  return { ...theme.tokens, ...(theme.variants?.[side] ?? {}) };
}

/**
 * Inline style that previews a theme on its card: the tokens as custom
 * properties, scoped to the card. Values are checked again here even though
 * the server validated them, so nothing but colours, radii and the known font
 * stacks gets through; shadows and gradients are left out of the preview.
 */
export function previewStyle(
  theme: Pick<ThemeInfo, 'tokens' | 'variants'>,
  side: ThemeSide = 'light',
): string {
  return Object.entries(resolvedTokens(theme, side))
    .filter(([k]) => /^--[a-z0-9-]+$/.test(k))
    .map(([k, v]): [string, string | undefined] => [
      k,
      k.startsWith('--font-') ? FONT_STACKS[v] : safeValue.test(v) ? v : undefined,
    ])
    .filter((e): e is [string, string] => e[1] !== undefined)
    .map(([k, v]) => `${k}: ${v}`)
    .join('; ');
}
