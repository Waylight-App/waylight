import { describe, expect, it } from 'vitest';
import { nextHint } from './hints';

describe('nextHint', () => {
  const go = { focused: false, reducedMotion: false };
  it('rotates and wraps around', () => {
    expect(nextHint(0, 4, go)).toBe(1);
    expect(nextHint(3, 4, go)).toBe(0);
    expect(nextHint(0, 0, go)).toBe(0);
  });
  it('pauses on focus and stays on the first hint for reduced motion', () => {
    expect(nextHint(2, 4, { ...go, focused: true })).toBe(2);
    expect(nextHint(3, 4, { ...go, reducedMotion: true })).toBe(0);
  });
});
