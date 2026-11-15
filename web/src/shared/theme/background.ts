/**
 * Background values (#277), shared by the global background and the
 * per-tab background. The server validates the same forms
 * (kernel.ParseBackground):
 *
 * - "" – default (for a tab: the global background)
 * - an http(s) image URL
 * - "/api/appearance/backgrounds/<id>" – a picture uploaded to this server
 * - "builtin:<name>" – a gradient or colour drawn by CSS, no request at all
 */

/** A background that ships with the app. */
export interface BuiltinBackground {
  name: string;
  /** CSS background-image value. */
  css: string;
  /** True for plain colours, false for gradients. */
  plain: boolean;
}

export const BUILTIN_PREFIX = 'builtin:';

/** Built-in backgrounds: CSS only, so they work offline and weigh nothing. */
export const BUILTIN_BACKGROUNDS: readonly BuiltinBackground[] = [
  { name: 'dusk', css: 'linear-gradient(135deg, #2b1055 0%, #7597de 100%)', plain: false },
  { name: 'dawn', css: 'linear-gradient(135deg, #ff9a8b 0%, #ff6a88 55%, #ff99ac 100%)', plain: false },
  { name: 'forest', css: 'linear-gradient(160deg, #134e5e 0%, #71b280 100%)', plain: false },
  { name: 'ocean', css: 'linear-gradient(160deg, #0f2027 0%, #203a43 50%, #2c5364 100%)', plain: false },
  { name: 'aurora', css: 'linear-gradient(135deg, #00c9a7 0%, #845ec2 100%)', plain: false },
  { name: 'sand', css: 'linear-gradient(160deg, #e6dada 0%, #c9b79c 100%)', plain: false },
  { name: 'slate', css: 'linear-gradient(#2f3640, #2f3640)', plain: true },
  { name: 'paper', css: 'linear-gradient(#f4f1ea, #f4f1ea)', plain: true },
];

const uploaded = /^\/api\/appearance\/backgrounds\/[A-Za-z0-9_-]{1,64}$/;
const builtin = /^builtin:[a-z0-9-]{1,32}$/;
// Quotes, parentheses, backslashes and whitespace could break out of url().
const imageUrl = /^https?:\/\/[^\s"'()\\]+$/i;

/** Whether a value is a background the server accepts. */
export function isValidBackground(value: string): boolean {
  const v = value.trim();
  return v === '' || uploaded.test(v) || builtin.test(v) || (imageUrl.test(v) && v.length <= 2048);
}

/** True for a picture uploaded to this server. */
export const isUploadedBackground = (value: string): boolean => uploaded.test(value.trim());

/** The built-in background of a value, if it is one. */
export function builtinBackground(value: string): BuiltinBackground | undefined {
  const v = value.trim();
  if (!v.startsWith(BUILTIN_PREFIX)) return undefined;
  return BUILTIN_BACKGROUNDS.find((b) => b.name === v.slice(BUILTIN_PREFIX.length));
}

/**
 * The CSS background-image for a value, or null for the default. Unknown
 * built-in names (from a newer version) fall back to the default.
 */
export function backgroundCss(value: string): string | null {
  const v = value.trim();
  if (v === '') return null;
  if (v.startsWith(BUILTIN_PREFIX)) return builtinBackground(v)?.css ?? null;
  if (uploaded.test(v) || imageUrl.test(v)) return `url("${v}")`;
  return null;
}
