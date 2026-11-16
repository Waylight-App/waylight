import { describe, expect, it } from 'vitest';
import { calcSource, timeSource } from './searchSources';

describe('calcSource (#160)', () => {
  const direct = (q: string) => calcSource().direct?.(q) ?? null;

  it('answers arithmetic with a copyable plain value', () => {
    const a = direct('1000*1000+0.5');
    expect(a).toMatchObject({ kind: 'metric', intent: 'calc.expression', title: '1000*1000+0.5 =' });
    // No thousands separators in the clipboard value.
    expect(a?.copy).toMatch(/^1000000[.,]5$/);
    expect(direct('1/3')?.copy).toMatch(/^0[.,]3333333333$/);
  });

  it('rounds conversions to six significant digits', () => {
    const a = direct('5 km in mi');
    expect(a).toMatchObject({ intent: 'calc.conversion', value: expect.stringMatching(/^3[.,]10686 mi$/) });
    expect(a?.copy).toMatch(/^3[.,]10686$/);
    expect(direct('30 °C in F')).toMatchObject({ value: '86°F' });
  });

  it('stays silent for other queries', () => {
    expect(direct('wetter morgen')).toBeNull();
    expect(direct('42')).toBeNull();
  });
});

describe('timeSource (#473)', () => {
  // Monday 16 November 2026, 15:04 local time.
  const now = new Date(2026, 10, 16, 15, 4);
  const direct = (q: string) => timeSource(() => now).direct?.(q) ?? null;

  it('answers time, week and date questions with a copyable value', () => {
    // The clock follows the locale: 15:04 in German, 03:04 PM in English.
    expect(direct('uhrzeit')).toMatchObject({ kind: 'metric', intent: 'time.now' });
    expect(direct('uhrzeit')?.copy).toMatch(/^(15:04|03:04\s?PM)$/);
    expect(direct('kw')).toMatchObject({ intent: 'time.week', value: '47', copy: '47' });
    expect(direct('tage bis weihnachten')).toMatchObject({ intent: 'time.until', copy: '39' });
    expect(direct('in 90 tagen')?.copy).toMatch(/2027/);
  });

  it('shows the time of another zone', () => {
    const a = direct('time in utc');
    expect(a).toMatchObject({ intent: 'time.zone' });
    expect(a?.copy).toMatch(/^\d{2}:\d{2}(\s?[AP]M)?$/);
  });

  it('stays silent for other queries', () => {
    expect(direct('wetter morgen')).toBeNull();
    expect(direct('5 km in mi')).toBeNull();
  });
});
