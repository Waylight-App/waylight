import { describe, expect, it } from 'vitest';
import { DEFAULT_ALBUM, DEFAULT_INTERVAL, move, nextIndex, parseConfig, usedPercent } from './photos';

describe('photo frame domain', () => {
  it('parses the config with defaults', () => {
    expect(parseConfig({})).toEqual({ album: DEFAULT_ALBUM, interval: DEFAULT_INTERVAL, fit: 'contain' });
    expect(parseConfig({ album: ' Summer ', interval: 60, fit: 'cover' })).toEqual({
      album: 'Summer',
      interval: 60,
      fit: 'cover',
    });
    expect(parseConfig({ interval: 1 }).interval).toBe(DEFAULT_INTERVAL);
  });

  it('wraps around', () => {
    expect(nextIndex(2, 3)).toBe(0);
    expect(nextIndex(0, 0)).toBe(0);
  });

  it('moves within bounds', () => {
    expect(move(['a', 'b', 'c'], 2, -1)).toEqual(['a', 'c', 'b']);
    expect(move(['a'], 0, 1)).toBeNull();
  });

  it('computes the storage share', () => {
    expect(usedPercent({ usedBytes: 75, quotaBytes: 300 })).toBe(25);
    expect(usedPercent({ usedBytes: 1, quotaBytes: 0 })).toBe(0);
  });
});
