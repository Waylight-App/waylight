import { describe, expect, it } from 'vitest';
import { addDays, daysBetween, easter, isoWeek, isoWeekday, isValid } from './dates';
import { zoneOf } from './places';
import { offsetFromLocal, parseDate, parseTimeQuery } from './time';

const d = (year: number, month: number, day: number) => ({ year, month, day });
// Monday 16 November 2026, 15:00 local time.
const now = new Date(2026, 10, 16, 15, 0);
const ask = (q: string, order: 'dmy' | 'mdy' = 'dmy') => parseTimeQuery(q, now, order);

describe('calendar arithmetic (#473)', () => {
  it('computes the ISO week around year boundaries', () => {
    expect(isoWeek(d(2026, 11, 16))).toEqual({ week: 47, year: 2026 });
    expect(isoWeek(d(2026, 1, 1))).toEqual({ week: 1, year: 2026 }); // a Thursday
    expect(isoWeek(d(2027, 1, 1))).toEqual({ week: 53, year: 2026 }); // a Friday
    expect(isoWeek(d(2024, 12, 30))).toEqual({ week: 1, year: 2025 }); // a Monday
    expect(isoWeek(d(2021, 1, 3))).toEqual({ week: 53, year: 2020 });
    expect(isoWeek(d(2020, 12, 31))).toEqual({ week: 53, year: 2020 });
  });

  it('knows weekdays, day differences and real dates', () => {
    expect(isoWeekday(d(2026, 11, 16))).toBe(1);
    expect(isoWeekday(d(2026, 11, 22))).toBe(7);
    expect(daysBetween(d(2026, 3, 28), d(2026, 3, 30))).toBe(2); // across the DST change
    expect(addDays(d(2026, 2, 27), 2)).toEqual(d(2026, 3, 1));
    expect(isValid(d(2028, 2, 29))).toBe(true);
    expect(isValid(d(2027, 2, 29))).toBe(false);
    expect(isValid(d(2026, 4, 31))).toBe(false);
  });

  it('computes Easter Sunday', () => {
    expect(easter(2026)).toEqual(d(2026, 4, 5));
    expect(easter(2027)).toEqual(d(2027, 3, 28));
    expect(easter(2038)).toEqual(d(2038, 4, 25));
  });
});

describe('time questions (#473)', () => {
  it('answers time, date and week in German and English', () => {
    for (const q of ['uhrzeit', 'Wie spät ist es?', 'time']) expect(ask(q)).toEqual({ kind: 'now' });
    for (const q of ['datum', 'heute', 'date', 'today']) expect(ask(q)).toEqual({ kind: 'today' });
    for (const q of ['kw', 'KW', 'week']) expect(ask(q)).toEqual({ kind: 'week', week: 47, year: 2026 });
  });

  it('finds the time zone of known places', () => {
    expect(ask('zeit in tokio')).toEqual({ kind: 'zone', place: 'tokio', zone: 'Asia/Tokyo' });
    expect(ask('time in New York')).toMatchObject({ zone: 'America/New_York' });
    expect(ask('uhrzeit in zürich')).toMatchObject({ zone: 'Europe/Zurich' });
    expect(ask('time in atlantis')).toBeNull();
    expect(zoneOf('sao paulo')).toBe('America/Sao_Paulo');
  });

  it('counts the days until named and typed dates', () => {
    expect(ask('tage bis weihnachten')).toEqual({
      kind: 'until',
      target: d(2026, 12, 25),
      days: 39,
      name: 'christmas',
    });
    expect(ask('days until 2027-03-01')).toMatchObject({ target: d(2027, 3, 1), days: 105, name: null });
    expect(ask('tage bis ostern')).toMatchObject({ target: d(2027, 3, 28) }); // Easter 2026 has passed
    expect(ask('wie lange noch bis silvester')).toMatchObject({ days: 45 });
    expect(ask('tage bis 1.3.')).toMatchObject({ target: d(2027, 3, 1) });
    expect(ask('days until 3/1', 'mdy')).toMatchObject({ target: d(2027, 3, 1) });
    expect(ask('tage bis 31.4.')).toBeNull();
  });

  it('shifts today by days and weeks', () => {
    expect(ask('in 90 tagen')).toEqual({ kind: 'shift', date: d(2027, 2, 14), days: 90 });
    expect(ask('90 days ago')).toEqual({ kind: 'shift', date: d(2026, 8, 18), days: -90 });
    expect(ask('vor 2 wochen')).toMatchObject({ date: d(2026, 11, 2), days: -14 });
    expect(ask('in 1 week')).toMatchObject({ days: 7 });
  });

  it('stays silent for everything else', () => {
    for (const q of ['', 'zeiterfassung', 'time machine', 'in 90 minuten', '5 km in mi', 'heute abend kino'])
      expect(ask(q)).toBeNull();
  });

  it('reads dates in both orders', () => {
    const today = d(2026, 11, 16);
    expect(parseDate('03/04/2027', today, 'dmy')).toEqual(d(2027, 4, 3));
    expect(parseDate('03/04/2027', today, 'mdy')).toEqual(d(2027, 3, 4));
    expect(parseDate('29.2.', today, 'dmy')).toEqual(d(2027, 3, 1)); // no 29 February in 2027
  });
});

describe('offsetFromLocal (#473)', () => {
  it('is the difference between the zone and the browser time zone', () => {
    const at = new Date(Date.UTC(2026, 0, 15, 12));
    const local = -at.getTimezoneOffset();
    expect(offsetFromLocal('UTC', at)).toBe(-local);
    expect(offsetFromLocal('Asia/Tokyo', at)).toBe(540 - local);
    expect(offsetFromLocal('Asia/Kolkata', at)).toBe(330 - local);
  });
});
