import { describe, expect, it } from 'vitest';
import { tabIcon } from './tabIcon';

describe('tabIcon', () => {
  it('knows familiar tab names in English and German', () => {
    expect(tabIcon('Home')).toBe('home');
    expect(tabIcon('Zuhause')).toBe('home');
    expect(tabIcon('Server')).toBe('server');
    expect(tabIcon('Kalender')).toBe('calendar');
    expect(tabIcon('Out and about')).toBe('pin');
  });

  it('falls back to a neutral icon', () => {
    expect(tabIcon('Zebras')).toBe('bookmark');
  });
});
