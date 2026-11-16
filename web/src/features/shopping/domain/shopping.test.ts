import { describe, expect, it } from 'vitest';
import { exportUrl, resolveCurrent, type ShoppingList, suggestListUrl, withDone } from './shopping';

const list = (id: string, mine: boolean): ShoppingList => ({
  id,
  name: id,
  visibility: 'private',
  mine,
  open: 0,
});

describe('resolveCurrent', () => {
  it('keeps a saved choice that still exists', () => {
    expect(resolveCurrent({ lists: [list('a', true), list('b', true)], caldav: false, current: 'b' })).toBe(
      'b',
    );
    expect(resolveCurrent({ lists: [list('a', true)], caldav: true, current: 'caldav' })).toBe('caldav');
  });
  it('falls back to the first own list, then CalDAV, then a shared list', () => {
    expect(resolveCurrent({ lists: [list('s', false), list('o', true)], caldav: true, current: '' })).toBe(
      'o',
    );
    expect(resolveCurrent({ lists: [list('s', false)], caldav: true, current: 'gone' })).toBe('caldav');
    expect(resolveCurrent({ lists: [list('s', false)], caldav: false, current: 'caldav' })).toBe('s');
  });
  it('is empty when there is nothing to show', () => {
    expect(resolveCurrent({ lists: [], caldav: false, current: '' })).toBe('');
  });
});

describe('shopping domain', () => {
  it('builds export addresses', () => {
    expect(exportUrl('a b', 'ics')).toBe('/api/shopping/lists/a%20b/export?format=ics');
  });
  it('suggests a list next to the calendar', () => {
    expect(suggestListUrl('https://c/remote.php/dav/calendars/me/personal/')).toBe(
      'https://c/remote.php/dav/calendars/me/shopping/',
    );
    expect(suggestListUrl('nonsense')).toBe('');
  });
  it('updates done immutably', () => {
    const items = [{ id: 'a', summary: 'Milk', done: false }];
    const next = withDone(items, 'a', true);
    expect(next[0]?.done).toBe(true);
    expect(items[0]?.done).toBe(false);
  });
});
