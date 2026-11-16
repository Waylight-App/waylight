import { fold } from '../smart/text';
import { addDays, civil, type CivilDate, daysBetween, easter, isoWeek, isValid } from './dates';
import { zoneOf } from './places';

/**
 * Time and date answers of the search (#473): "uhrzeit", "kw", "time in
 * tokyo", "tage bis weihnachten", "in 90 tagen". The query is read as a
 * whole, like the calculator, and anything else yields null so the normal
 * search takes over. The current moment comes from the caller.
 */
export type TimeQuery =
  | { kind: 'now' }
  | { kind: 'today' }
  | { kind: 'week'; week: number; year: number }
  | { kind: 'zone'; place: string; zone: string }
  | { kind: 'until'; target: CivilDate; days: number; name: NamedDay | null }
  | { kind: 'shift'; date: CivilDate; days: number };

export type NamedDay = 'christmas_eve' | 'christmas' | 'new_years_eve' | 'new_year' | 'easter';

/** Order of day and month in a date with slashes; dots are always day first. */
export type DateOrder = 'dmy' | 'mdy';

const NOW = new Set([
  'uhrzeit',
  'zeit',
  'wie spat',
  'wie spat ist es',
  'wieviel uhr',
  'wie viel uhr',
  'wieviel uhr ist es',
  'wie viel uhr ist es',
  'time',
  'what time is it',
  'current time',
  'local time',
]);

const TODAY = new Set([
  'datum',
  'heute',
  'welcher tag',
  'welcher tag ist heute',
  'welches datum',
  'den wievielten haben wir',
  'date',
  'today',
  'todays date',
  "today's date",
  'what day is it',
  'what is the date',
]);

const WEEK = new Set([
  'kw',
  'kalenderwoche',
  'welche kw',
  'welche kalenderwoche',
  'week',
  'week number',
  'calendar week',
  'which week',
  'what week is it',
]);

const NAMED: Record<string, NamedDay> = {
  heiligabend: 'christmas_eve',
  'christmas eve': 'christmas_eve',
  weihnachten: 'christmas',
  christmas: 'christmas',
  xmas: 'christmas',
  silvester: 'new_years_eve',
  'new years eve': 'new_years_eve',
  "new year's eve": 'new_years_eve',
  neujahr: 'new_year',
  'new year': 'new_year',
  'new years day': 'new_year',
  "new year's day": 'new_year',
  ostern: 'easter',
  easter: 'easter',
};

function namedDay(name: NamedDay, year: number): CivilDate {
  switch (name) {
    case 'christmas_eve':
      return { year, month: 12, day: 24 };
    case 'christmas':
      return { year, month: 12, day: 25 };
    case 'new_years_eve':
      return { year, month: 12, day: 31 };
    case 'new_year':
      return { year, month: 1, day: 1 };
    case 'easter':
      return easter(year);
  }
}

/** The next time a yearly day comes, today included. */
function next(today: CivilDate, inYear: (year: number) => CivilDate): CivilDate {
  const date = inYear(today.year);
  return daysBetween(today, date) >= 0 ? date : inYear(today.year + 1);
}

/** A capture group as text; groups that did not take part are "". */
const group = (m: RegExpExecArray, i: number): string => m[i] ?? '';
/** A capture group as a number; groups that did not take part are 0. */
const num = (m: RegExpExecArray, i: number): number => Number(group(m, i) || 0);

/**
 * A date as typed: ISO "2027-03-01", "1.3.2027", "1.3." or "1/3/2027"
 * (order from the UI language). Without a year it means the next such day.
 */
export function parseDate(text: string, today: CivilDate, order: DateOrder): CivilDate | null {
  let m = /^(\d{4})-(\d{1,2})-(\d{1,2})$/.exec(text);
  if (m) return checked({ year: num(m, 1), month: num(m, 2), day: num(m, 3) });
  m = /^(\d{1,2})\.(\d{1,2})\.(\d{4})?$/.exec(text);
  if (m) return withYear(num(m, 2), num(m, 1), m[3], today);
  m = /^(\d{1,2})\/(\d{1,2})(?:\/(\d{4}))?$/.exec(text);
  if (m) {
    const [first, second] = [num(m, 1), num(m, 2)];
    return order === 'dmy' ? withYear(second, first, m[3], today) : withYear(first, second, m[3], today);
  }
  return null;
}

function checked(d: CivilDate): CivilDate | null {
  return isValid(d) ? d : null;
}

function withYear(month: number, day: number, year: string | undefined, today: CivilDate): CivilDate | null {
  if (year) return checked({ year: +year, month, day });
  if (!isValid({ year: 2000, month, day })) return null; // 2000 is a leap year, so 29.2. passes
  return next(today, (y) =>
    isValid({ year: y, month, day }) ? { year: y, month, day } : { year: y, month: 3, day: 1 },
  );
}

const UNTIL =
  /^(?:(?:noch\s+)?(?:wie viele\s+|wieviele\s+)?tage\s+bis|wie lange\s+(?:noch\s+)?bis|(?:how many\s+)?days\s+(?:until|till|to)|how long\s+(?:until|till))\s+(.+)$/;
const AHEAD = /^in\s+(\d{1,5})\s+(tag|tagen|tage|days?|woche|wochen|weeks?)$/;
const AGO = /^(?:vor\s+(\d{1,5})\s+(tag|tagen|tage|woche|wochen)|(\d{1,5})\s+(days?|weeks?)\s+ago)$/;
const ZONE = /^(?:(?:uhr)?zeit|wie spat(?: ist es)?|time|what time is it|current time)\s+in\s+(.+)$/;

const daysOf = (n: number, unit: string) => (/^w/.test(unit) ? n * 7 : n);

/** Reads a time or date question, or returns null if the query is something else. */
export function parseTimeQuery(query: string, now: Date, order: DateOrder): TimeQuery | null {
  const q = fold(query)
    .trim()
    .replace(/[?!]+$/, '')
    .replace(/\s+/g, ' ');
  if (q.length < 2 || q.length > 60) return null;
  const today = civil(now);

  if (NOW.has(q)) return { kind: 'now' };
  if (TODAY.has(q)) return { kind: 'today' };
  if (WEEK.has(q)) return { kind: 'week', ...isoWeek(today) };

  let m = ZONE.exec(q);
  if (m) {
    const place = group(m, 1);
    const zone = zoneOf(place);
    return zone ? { kind: 'zone', place, zone } : null;
  }

  m = UNTIL.exec(q);
  if (m) {
    const what = group(m, 1).replace(/^(?:zum|zur|dem|den|the)\s+/, '');
    const name = NAMED[what] ?? null;
    const target = name ? next(today, (y) => namedDay(name, y)) : parseDate(what, today, order);
    return target ? { kind: 'until', target, days: daysBetween(today, target), name } : null;
  }

  m = AHEAD.exec(q);
  if (m) {
    const days = daysOf(num(m, 1), group(m, 2));
    return { kind: 'shift', date: addDays(today, days), days };
  }
  m = AGO.exec(q);
  if (m) {
    // German "vor 3 tagen" fills groups 1–2, English "3 days ago" groups 3–4.
    const days = m[1] ? -daysOf(num(m, 1), group(m, 2)) : -daysOf(num(m, 3), group(m, 4));
    return { kind: 'shift', date: addDays(today, days), days };
  }
  return null;
}

/**
 * Minutes a zone is ahead of the browser's own time at a moment, e.g. 480
 * for Tokyo seen from Berlin in winter. Uses the zone's offset as Intl
 * reports it ("GMT+09:00"), so no zone data is bundled.
 */
export function offsetFromLocal(zone: string, at: Date): number {
  const name =
    new Intl.DateTimeFormat('en-US', { timeZone: zone, timeZoneName: 'longOffset' })
      .formatToParts(at)
      .find((p) => p.type === 'timeZoneName')?.value ?? 'GMT';
  const m = /GMT([+-])(\d{2}):?(\d{2})?/.exec(name);
  const zoneOffset = m ? (group(m, 1) === '-' ? -1 : 1) * (num(m, 2) * 60 + num(m, 3)) : 0;
  return zoneOffset + at.getTimezoneOffset();
}
