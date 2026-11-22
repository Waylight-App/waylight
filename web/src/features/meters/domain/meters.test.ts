import { describe, expect, it } from 'vitest';
import { barHeights } from '../../../shared/ui/bars';
import { parseKind, today, validateMeter, validateReading, yearChoices } from './meters';

describe('meters domain', () => {
  it('validates the meter dialog', () => {
    expect(validateMeter({ name: 'Power', kind: 'electricity', unit: 'kWh' })).toBeNull();
    expect(validateMeter({ name: ' ', kind: 'gas', unit: 'm³' })).toBe('meters.error.name');
    expect(validateMeter({ name: 'x'.repeat(61), kind: 'gas', unit: 'm³' })).toBe('meters.error.name');
    expect(validateMeter({ name: 'Gas', kind: 'gas', unit: '' })).toBe('meters.error.unit');
    expect(validateMeter({ name: 'Gas', kind: 'gas', unit: 'x'.repeat(13) })).toBe('meters.error.unit');
  });

  it('checks a reading against the others like the server', () => {
    const existing = [
      { day: '2026-01-01', value: 100 },
      { day: '2026-03-01', value: 200 },
    ];
    const check = (day: string, value: number | null) =>
      validateReading({ day, value }, existing, '2026-11-22');
    expect(check('2026-02-01', 150)).toBeNull();
    expect(check('2026-03-01', 120)).toBeNull(); // replaces the reading of that day
    expect(check('2026-02-01', 99)).toBe('meters.error.lower');
    expect(check('2026-02-01', 201)).toBe('meters.error.higher');
    expect(check('2026-11-23', 300)).toBe('meters.error.day');
    expect(check('', 300)).toBe('meters.error.day');
    expect(check('2026-04-01', null)).toBe('meters.error.value');
    expect(check('2026-04-01', -1)).toBe('meters.error.value');
  });

  it('formats today in local time and offers the current year', () => {
    expect(today(new Date(2026, 0, 5, 23, 30))).toBe('2026-01-05');
    expect(yearChoices([2025, 2024], 2026)).toEqual([2026, 2025, 2024]);
    expect(yearChoices([2026, 2025], 2026)).toEqual([2026, 2025]);
    expect(parseKind('water')).toBe('water');
    expect(parseKind('oil')).toBe('other');
  });

  it('scales the bars from zero', () => {
    expect(barHeights([10, 5, 0])).toEqual({ heights: [100, 50, 0], zero: 100 });
  });
});
