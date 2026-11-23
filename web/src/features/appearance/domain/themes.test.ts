import { describe, expect, it, vi } from 'vitest';
import { FONT_STACKS, filterThemes, previewSide, previewStyle, resolvedTokens, themeSides } from './themes';

describe('theme preview (#111)', () => {
  it('only lets colors and radii through', () => {
    expect(
      previewStyle({
        tokens: {
          '--color-bg': '#2e3440',
          '--radius': '6px',
          '--color-text': 'red; background: url(https://evil.example)',
          color: '#fff',
        },
      }),
    ).toBe('--color-bg: #2e3440; --radius: 6px');
  });

  it('turns font keywords into their stack and drops unknown fonts (#505)', () => {
    expect(previewStyle({ tokens: { '--font-display': 'fraunces', '--font-sans': "url(x.woff2)'" } })).toBe(
      `--font-display: ${FONT_STACKS['fraunces']}`,
    );
  });
});

describe('themes with light and dark variants (#505)', () => {
  const theme = {
    scheme: '' as const,
    tokens: { '--radius': '14px' },
    variants: {
      light: { '--color-bg': '#fff3d6' },
      dark: { '--color-bg': '#131829' },
    },
  };

  it('previews both sides', () => {
    expect(themeSides(theme)).toEqual(['light', 'dark']);
    expect(themeSides({ scheme: 'dark' })).toEqual(['dark']);
  });

  it('resolves the shared tokens with the variant on top', () => {
    expect(resolvedTokens(theme, 'dark')).toEqual({ '--radius': '14px', '--color-bg': '#131829' });
    expect(previewStyle(theme, 'light')).toBe('--radius: 14px; --color-bg: #fff3d6');
  });
});

describe('theme import (#111)', () => {
  it('refuses files above 16 KB before uploading', async () => {
    const { AppearanceUseCases } = await import('../usecase/appearanceUseCases');
    const importTheme = vi.fn().mockResolvedValue({ id: 'x' });
    const uc = new AppearanceUseCases({ importTheme } as unknown as ConstructorParameters<
      typeof AppearanceUseCases
    >[0]);
    await expect(uc.importTheme({ size: (16 << 10) + 1, text: () => Promise.resolve('') })).rejects.toThrow();
    await uc.importTheme({ size: 2, text: () => Promise.resolve('{}') });
    expect(importTheme).toHaveBeenCalledWith('{}');
  });
});

describe('filterThemes and previewSide', () => {
  const light = { scheme: 'light' as const, builtIn: true };
  const dark = { scheme: 'dark' as const, builtIn: true };
  const both = { scheme: '' as const, variants: { light: {}, dark: {} }, builtIn: true };
  const own = { scheme: 'dark' as const, builtIn: false };
  it('filters by scheme and origin', () => {
    const all = [light, dark, both, own];
    expect(filterThemes(all, 'all')).toHaveLength(4);
    expect(filterThemes(all, 'light')).toEqual([light, both]);
    expect(filterThemes(all, 'dark')).toEqual([dark, both, own]);
    expect(filterThemes(all, 'own')).toEqual([own]);
  });
  it('previews a design with both variants in the scheme in use', () => {
    expect(previewSide(both, true)).toBe('dark');
    expect(previewSide(both, false)).toBe('light');
    expect(previewSide(light, true)).toBe('light');
  });
});
