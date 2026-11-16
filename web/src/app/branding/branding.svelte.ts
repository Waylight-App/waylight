import type { HttpClient } from '../../shared/api-client/http';

/**
 * The instance's branding (#305) as GET /api/branding returns it: public, in
 * every build. Without the licensed feature it is the product's own, so the
 * shell can always use it. Image fields are versioned URLs; '' means the
 * product's image.
 */
export interface Branding {
  name: string;
  /** True when an organisation set its own branding. */
  custom: boolean;
  accent?: string;
  secondary?: string;
  hidePoweredBy: boolean;
  logo?: string;
  logoDark?: string;
  favicon?: string;
  /** The organisation's theme ('' = none); the sign-in page uses it. */
  defaultTheme?: '' | 'system' | 'light' | 'dark';
}

export const PRODUCT_NAME = 'trisse-home';
const defaultBranding: Branding = { name: PRODUCT_NAME, custom: false, hidePoweredBy: false };

const state = $state<{ current: Branding }>({ current: defaultBranding });

/** The branding the shell shows; reactive. */
export function branding(): Branding {
  return state.current;
}

const hex = /^#[0-9a-f]{6}$/i;

/**
 * Applies the branding to the document: title, favicon and the accent
 * colours as base values. A person's own accent colour (appearance) is set
 * inline on <html> and therefore still wins.
 */
export function applyBranding(b: Branding, doc: Document = document): void {
  doc.title = b.name || PRODUCT_NAME;
  const icon = doc.querySelector<HTMLLinkElement>('link[rel="icon"]');
  if (icon) {
    icon.href = b.favicon || '/favicon.svg';
    icon.type = b.favicon ? '' : 'image/svg+xml';
  }
  let style = doc.getElementById('branding-colours');
  const rules = [
    b.accent && hex.test(b.accent) ? `--color-accent:${b.accent};` : '',
    b.secondary && hex.test(b.secondary) ? `--color-secondary:${b.secondary};` : '',
  ].join('');
  if (!rules) {
    style?.remove();
    return;
  }
  if (!style) {
    style = doc.createElement('style');
    style.id = 'branding-colours';
    doc.head.append(style);
  }
  // Only validated hex colours reach the style sheet.
  style.textContent = `:root{${rules}}`;
}

/** Loads and applies the branding; on any error the product's own stays. */
export async function loadBranding(http: HttpClient): Promise<Branding> {
  try {
    const b = await http.get<Branding>('/api/branding');
    state.current = { ...defaultBranding, ...b, name: b.name || PRODUCT_NAME };
  } catch {
    state.current = defaultBranding;
  }
  applyBranding(state.current);
  return state.current;
}
