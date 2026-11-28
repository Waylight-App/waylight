import { describe, expect, it } from 'vitest';
import { iconPaths } from './paths';

// A path made only of ".01" dots renders as specks at 16 px, so the icon looks
// like an empty button (#852). Dots are fine as accents next to real strokes.
const dotOnly = /^(M[\d.]+ [\d.]+h\.01)+$/;

describe('iconPaths', () => {
  it.each(Object.entries(iconPaths))('%s draws more than dots', (_, paths) => {
    expect(paths.some((d) => !dotOnly.test(d.replace(/\s*(?=M)/g, '')))).toBe(true);
  });
});
