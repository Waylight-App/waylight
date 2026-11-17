import { describe, expect, it } from 'vitest';
import { isGalleryThemeCss, lastLook, rememberLook } from './lastLook';

/** A Map-backed Storage for the tests. */
function memoryStorage(): Storage {
  const m = new Map<string, string>();
  return {
    get length() {
      return m.size;
    },
    clear: () => m.clear(),
    getItem: (k) => m.get(k) ?? null,
    key: (i) => [...m.keys()][i] ?? null,
    removeItem: (k) => void m.delete(k),
    setItem: (k, v) => void m.set(k, v),
  };
}

const amber =
  '/* waylight theme: amber */\n:root:root,\n:root:root[data-theme] {\n  --color-bg: #fff3d6;\n}\n';

describe('the sign-in page keeps the last look (#526)', () => {
  it('keeps a gallery theme and the scheme', () => {
    const s = memoryStorage();
    rememberLook({ css: amber, theme: 'dark' }, s);
    expect(lastLook(s)).toEqual({ css: amber, theme: 'dark' });
  });

  it('forgets it for own CSS or no theme', () => {
    const s = memoryStorage();
    rememberLook({ css: amber, theme: 'light' }, s);
    rememberLook({ css: 'body { color: red }', theme: 'light' }, s);
    expect(lastLook(s)).toBeNull();
  });

  it('refuses entries that were changed to load something', () => {
    expect(isGalleryThemeCss(amber)).toBe(true);
    expect(isGalleryThemeCss(amber + 'body{background:url(https://evil.example/x)}')).toBe(false);
    expect(isGalleryThemeCss('/* waylight theme: x */ @import "https://evil.example/a.css";')).toBe(false);
    expect(isGalleryThemeCss(amber + 'x'.repeat(17 << 10))).toBe(false);
    const s = memoryStorage();
    s.setItem('waylight:sign-in-look', JSON.stringify({ css: 'body{}', theme: 'dark' }));
    expect(lastLook(s)).toBeNull();
    s.setItem('waylight:sign-in-look', '{not json');
    expect(lastLook(s)).toBeNull();
  });

  it('falls back to the system scheme for an unknown one', () => {
    const s = memoryStorage();
    s.setItem('waylight:sign-in-look', JSON.stringify({ css: amber, theme: 'sepia' }));
    expect(lastLook(s)?.theme).toBe('system');
  });
});
