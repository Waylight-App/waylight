import { describe, expect, it } from 'vitest';
import { DEFAULT_LIMIT, parseConfig, toPayload, whenKey } from './birthdays';

describe('birthdays domain', () => {
  it('parses the widget limit with a safe default', () => {
    expect(parseConfig({ limit: 8 })).toEqual({ limit: 8 });
    expect(parseConfig({ limit: 0 })).toEqual({ limit: DEFAULT_LIMIT });
    expect(parseConfig({ limit: '3' })).toEqual({ limit: DEFAULT_LIMIT });
  });

  it('builds the payload and drops an unknown year', () => {
    expect(toPayload({ name: ' Anna  Maria ', date: '1990-02-28', yearKnown: true })).toEqual({
      name: 'Anna Maria',
      month: 2,
      day: 28,
      year: 1990,
    });
    expect(toPayload({ name: 'Ben', date: '2000-12-01', yearKnown: false })).toMatchObject({ year: 0 });
    expect(toPayload({ name: '', date: '2000-12-01', yearKnown: true })).toBe('birthdays.invalid_name');
    expect(toPayload({ name: 'Ben', date: '', yearKnown: true })).toBe('birthdays.invalid_date');
  });

  it('labels the distance', () => {
    expect(whenKey(0).key).toBe('birthdays.today');
    expect(whenKey(1).key).toBe('birthdays.tomorrow');
    expect(whenKey(12)).toEqual({ key: 'birthdays.in_days', n: 12 });
  });
});
