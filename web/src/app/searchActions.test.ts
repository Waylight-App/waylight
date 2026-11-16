import { describe, expect, it } from 'vitest';
import { matchActions, matches, type SearchAction } from './searchActions';

const action = (id: string, keywords: string[]): SearchAction => ({
  id,
  keywords,
  title: () => id,
  detail: () => '',
  run: () => Promise.resolve(''),
});

describe('search actions (#453)', () => {
  const office = action('office', ['office', 'büro']);
  const home = action('home', ['home', 'homeoffice']);

  it('matches a keyword or its beginning, ignoring case and spaces', () => {
    expect(matches(office, 'office')).toBe(true);
    expect(matches(office, '  OFF ')).toBe(true);
    expect(matches(office, 'Büro')).toBe(true);
    expect(matches(home, 'homeoff')).toBe(true);
  });

  it('needs two letters and never matches longer text', () => {
    expect(matches(office, 'o')).toBe(false);
    expect(matches(office, '')).toBe(false);
    expect(matches(office, 'office party')).toBe(false);
  });

  it('keeps the registration order and limits the list', () => {
    const all = [office, home, action('ho2', ['hotel']), action('ho3', ['hobby'])];
    expect(matchActions(all, 'ho').map((a) => a.id)).toEqual(['home', 'ho2', 'ho3']);
    expect(matchActions(all, 'ho', 2)).toHaveLength(2);
  });
});
