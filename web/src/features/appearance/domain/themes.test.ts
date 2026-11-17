import { describe, expect, it, vi } from 'vitest';
import { FONT_STACKS, previewStyle, resolvedTokens, themeSides } from './themes';

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
