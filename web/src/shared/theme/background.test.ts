import { describe, expect, it } from 'vitest';
import { backgroundCss, BUILTIN_BACKGROUNDS, isValidBackground } from './background';

describe('backgrounds', () => {
  it('accepts the four forms the server accepts', () => {
    for (const v of ['', 'https://img.example/a.jpg', 'builtin:dusk', '/api/appearance/backgrounds/abc'])
      expect(isValidBackground(v)).toBe(true);
    for (const v of [
      'ftp://x',
      'https://x/a.jpg) red',
      '/etc/passwd',
      'builtin:Dusk!',
      'javascript:alert(1)',
    ])
      expect(isValidBackground(v)).toBe(false);
  });

  it('turns values into CSS', () => {
    expect(backgroundCss('')).toBeNull();
    expect(backgroundCss('builtin:dusk')).toBe(BUILTIN_BACKGROUNDS[0]?.css);
    expect(backgroundCss('builtin:unknown')).toBeNull();
    expect(backgroundCss('/api/appearance/backgrounds/a1')).toBe('url("/api/appearance/backgrounds/a1")');
    expect(backgroundCss('https://x/a.jpg')).toBe('url("https://x/a.jpg")');
  });
});
