import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '../../shared/api-client/http';
import { branding, loadBranding, PRODUCT_NAME } from './branding.svelte';

/** A tiny stand-in for the document parts applyBranding touches. */
function fakeDocument() {
  const icon = { href: '/favicon.svg', type: 'image/svg+xml' };
  const styles: { id: string; textContent: string; remove: () => void }[] = [];
  const doc = {
    title: '',
    querySelector: () => icon,
    getElementById: (id: string) => styles.find((s) => s.id === id) ?? null,
    createElement: () => {
      const s = { id: '', textContent: '', remove: () => styles.splice(styles.indexOf(s), 1) };
      return s;
    },
    head: { append: (s: (typeof styles)[number]) => styles.push(s) },
  };
  return { doc, icon, styles };
}

describe('branding', () => {
  it('keeps the product branding when the endpoint fails', async () => {
    const { doc } = fakeDocument();
    vi.stubGlobal('document', doc);
    const http = { get: vi.fn().mockRejectedValue(new Error('404')) } as unknown as HttpClient;
    expect((await loadBranding(http)).name).toBe(PRODUCT_NAME);
    expect(doc.title).toBe(PRODUCT_NAME);
    vi.unstubAllGlobals();
  });

  it('applies name, favicon and only valid colours', async () => {
    const { doc, icon, styles } = fakeDocument();
    vi.stubGlobal('document', doc);
    const http = {
      get: vi.fn().mockResolvedValue({
        name: 'Acme',
        custom: true,
        accent: '#112233',
        secondary: 'red;}body{display:none',
        hidePoweredBy: true,
        favicon: '/api/branding/favicon?v=2',
      }),
    } as unknown as HttpClient;
    await loadBranding(http);
    expect(branding().name).toBe('Acme');
    expect(doc.title).toBe('Acme');
    expect(icon.href).toBe('/api/branding/favicon?v=2');
    expect(styles[0]?.textContent).toBe(':root{--color-accent:#112233;}');
    vi.unstubAllGlobals();
  });
});
