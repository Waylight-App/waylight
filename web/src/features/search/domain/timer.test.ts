import { describe, expect, it } from 'vitest';
import { asksForTimers, formatRemaining, parseDuration, parseTimer, running } from './timer';

// Monday, 16 November 2026, 12:00 local time.
const now = new Date(2026, 10, 16, 12, 0, 0);
const min = 60_000;

describe('timer parsing (#474)', () => {
  it.each([
    ['timer 10 min', 10 * min, ''],
    ['timer 10min', 10 * min, ''],
    ['Timer 25 min Pizza', 25 * min, 'Pizza'],
    ['timer 1h30', 90 * min, ''],
    ['timer 1h 30m', 90 * min, ''],
    ['timer 1,5 stunden Wäsche', 90 * min, 'Wäsche'],
    ['timer 90', 90 * min, ''],
    ['timer 30 s', 30_000, ''],
    ['eieruhr 7 minuten', 7 * min, ''],
  ])('%s', (query, ms, label) => {
    expect(parseTimer(query, now)).toEqual({
      kind: 'timer',
      end: now.getTime() + ms,
      label,
      tomorrow: false,
    });
  });

  it('rejects text that is not a timer', () => {
    expect(parseTimer('timer', now)).toBeNull();
    expect(parseTimer('timer pizza', now)).toBeNull();
    expect(parseTimer('timer 0 min', now)).toBeNull();
    expect(parseTimer('timer 30 h', now)).toBeNull();
    expect(parseTimer('wie lange 10 min', now)).toBeNull();
  });

  it('reads durations word by word', () => {
    expect(parseDuration(['2', 'h', '15', 'min', 'tee'])).toEqual({ ms: 135 * min, used: 4 });
    expect(parseDuration(['tee'])).toBeNull();
  });
});

describe('alarm parsing (#474)', () => {
  it('sets an alarm for later today', () => {
    const a = parseTimer('wecker 17:30', now);
    expect(a?.kind).toBe('alarm');
    expect(new Date(a?.end ?? 0).getHours()).toBe(17);
    expect(new Date(a?.end ?? 0).getDate()).toBe(16);
    expect(a?.tomorrow).toBe(false);
  });

  it('sets a time that already passed for tomorrow', () => {
    const a = parseTimer('alarm 6:45 aufstehen', now);
    expect(new Date(a?.end ?? 0).getDate()).toBe(17);
    expect(new Date(a?.end ?? 0).getMinutes()).toBe(45);
    expect(a?.tomorrow).toBe(true);
    expect(a?.label).toBe('aufstehen');
  });

  it.each(['wecker um 18 uhr', 'wecker 18h', 'alarm at 18.00'])('understands %s', (q) => {
    expect(new Date(parseTimer(q, now)?.end ?? 0).getHours()).toBe(18);
  });

  it('rejects invalid times', () => {
    expect(parseTimer('wecker 25:00', now)).toBeNull();
    expect(parseTimer('wecker morgen', now)).toBeNull();
  });
});

describe('timer helpers (#474)', () => {
  it('formats the remaining time', () => {
    expect(formatRemaining(299_000)).toBe('4:59');
    expect(formatRemaining(3_909_000)).toBe('1:05:09');
    expect(formatRemaining(-5)).toBe('0:00');
  });

  it('lists running timers, soonest first', () => {
    const t = (id: string, end: number) => ({ id, kind: 'timer' as const, end, label: '', tomorrow: false });
    expect(running([t('a', 300), t('b', 50), t('c', 200)], 100).map((x) => x.id)).toEqual(['c', 'a']);
  });

  it('recognises the request for the timer list', () => {
    expect(asksForTimers('timer')).toBe(true);
    expect(asksForTimers('Wecker')).toBe(true);
    expect(asksForTimers('timer 5')).toBe(false);
  });
});
