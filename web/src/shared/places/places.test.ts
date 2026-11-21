import { describe, expect, it } from 'vitest';
import { placeAt, placeIdOf, type SavedPlace } from './places';

const home: SavedPlace = { id: 'p1', name: 'Home', label: '', lat: 53.1435, lon: 8.2146 };

describe('saved places', () => {
  it('reads the place id from widget settings', () => {
    expect(placeIdOf({ placeId: 'p1' })).toBe('p1');
    expect(placeIdOf({ placeId: 3 })).toBe('');
    expect(placeIdOf({})).toBe('');
  });

  it('finds a place at nearly the same coordinates', () => {
    expect(placeAt([home], 53.1438, 8.2149)).toBe(home);
    expect(placeAt([home], 53.2, 8.2146)).toBeUndefined();
  });
});
