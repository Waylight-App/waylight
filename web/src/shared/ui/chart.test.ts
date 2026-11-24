import { describe, expect, it } from 'vitest';
import { areaPath, niceMax, smoothPath } from './chart';

describe('smoothPath', () => {
  it('is empty without points', () => expect(smoothPath([])).toBe(''));

  it('starts at the first point and ends at the last', () => {
    const d = smoothPath([
      [0, 10],
      [50, 2],
      [100, 8],
    ]);
    expect(d.startsWith('M0,10C')).toBe(true);
    expect(d.endsWith('100,8')).toBe(true);
  });

  it('keeps control points between the heights of their segment', () => {
    const d = smoothPath([
      [0, 0],
      [10, 40],
      [20, 0],
    ]);
    const ys = [...d.matchAll(/C[\d.-]+,([\d.-]+) [\d.-]+,([\d.-]+)/g)].flatMap((m) => [
      Number(m[1]),
      Number(m[2]),
    ]);
    expect(ys.length).toBe(4);
    for (const y of ys) {
      expect(y).toBeGreaterThanOrEqual(0);
      expect(y).toBeLessThanOrEqual(40);
    }
  });
});

describe('areaPath', () => {
  it('closes the line down to the baseline', () => {
    const d = areaPath(
      [
        [0, 5],
        [10, 3],
      ],
      20,
    );
    expect(d.endsWith('L10,20L0,20Z')).toBe(true);
  });
});

describe('niceMax', () => {
  it('rounds up to 1, 2, 5 or 10 times a power of ten', () => {
    expect(niceMax(47)).toBe(50);
    expect(niceMax(96)).toBe(100);
    expect(niceMax(230)).toBe(500);
    expect(niceMax(0)).toBe(1);
  });
});
