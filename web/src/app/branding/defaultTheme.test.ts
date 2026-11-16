import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '../../shared/api-client/http';
import { loadDefaultTheme, showDefaultTheme } from './defaultTheme';

/** Just enough of a document for showDefaultTheme. */
function fakeDocument() {
  const styles: { dataset: Record<string, string>; textContent: string; remove: () => void }[] = [];
  const doc = {
    createElement: () => {
      const s = { dataset: {}, textContent: '', remove: () => styles.splice(styles.indexOf(s), 1) };
      return s;
    },
    head: { append: (s: (typeof styles)[number]) => styles.push(s) },
  };
  return { doc: doc as unknown as Document, styles };
}

describe('default theme on the sign-in page (#507)', () => {
  it('adds the theme CSS and takes it out again', () => {
    const { doc, styles } = fakeDocument();
    const remove = showDefaultTheme({ id: 'amber', css: ':root{--color-bg:#fff3d6}', scheme: '' }, doc);
    expect(styles).toHaveLength(1);
    expect(styles[0]?.textContent).toBe(':root{--color-bg:#fff3d6}');
    expect(styles[0]?.dataset['source']).toBe('default-theme');
    remove();
    expect(styles).toHaveLength(0);
  });

  it('adds nothing for the built-in look', () => {
    const { doc, styles } = fakeDocument();
    showDefaultTheme({ id: '', css: '', scheme: '' }, doc);
    expect(styles).toHaveLength(0);
  });

  it('keeps the built-in look when the server does not answer', async () => {
    const http = { get: vi.fn().mockRejectedValue(new Error('offline')) } as unknown as HttpClient;
    await expect(loadDefaultTheme(http)).resolves.toEqual({ id: '', css: '', scheme: '' });
  });
});
