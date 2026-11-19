import { describe, expect, it } from 'vitest';
import { bars, byDay, currentHour, parseConfig, type PriceHour } from './electricity';

const h = (iso: string, v: number, cheapest = false): PriceHour => ({
  start: new Date(iso),
  centPerKWh: v,
  cheapest,
});

describe('electricity prices (#288)', () => {
  it('reads the zone of a widget', () => {
    expect(parseConfig({})).toEqual({ zone: 'DE-LU' });
    expect(parseConfig({ zone: 'AT' })).toEqual({ zone: 'AT' });
    expect(parseConfig({ zone: 'FR' })).toEqual({ zone: 'DE-LU' });
  });

  it('groups by day and finds the current hour', () => {
    const hours = [h('2026-11-14T22:00:00', 10), h('2026-11-14T23:00:00', 9), h('2026-11-15T00:00:00', 8)];
    const now = new Date('2026-11-14T23:30:00');
    expect(byDay(hours, now).map((d) => [d.when, d.hours.length])).toEqual([
      ['today', 2],
      ['tomorrow', 1],
    ]);
    // Shortly after midnight yesterday's hours are dropped, not shown as today.
    expect(byDay(hours, new Date('2026-11-15T00:30:00')).map((d) => [d.when, d.hours.length])).toEqual([
      ['today', 1],
    ]);
    expect(currentHour(hours, new Date('2026-11-14T23:30:00'))?.centPerKWh).toBe(9);
    expect(currentHour(hours, new Date('2026-11-15T03:00:00'))).toBeUndefined();
  });

  it('scales bars from zero, negative prices below the base line', () => {
    expect(bars([h('2026-11-14T00:00:00', 10), h('2026-11-14T01:00:00', 5)])).toEqual({
      heights: [100, 50],
      zero: 100,
    });
    const neg = bars([h('2026-11-14T00:00:00', 6), h('2026-11-14T01:00:00', -2)]);
    expect(neg.zero).toBe(75);
    expect(neg.heights).toEqual([75, 25]);
  });
});
