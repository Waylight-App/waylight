/**
 * Calendar arithmetic for the time answers of the search (#473). Dates are
 * plain calendar days without a time or zone, so adding days never trips
 * over a daylight saving change. Pure functions only.
 */

/** A calendar day; month is 1–12. */
export interface CivilDate {
  year: number;
  month: number;
  day: number;
}

const DAY_MS = 86_400_000;

/** The calendar day of a moment in the browser's own time zone. */
export function civil(at: Date): CivilDate {
  return { year: at.getFullYear(), month: at.getMonth() + 1, day: at.getDate() };
}

/** Days since 1970-01-01 (UTC arithmetic, so every day has 24 hours). */
function ordinal(d: CivilDate): number {
  return Math.round(Date.UTC(d.year, d.month - 1, d.day) / DAY_MS);
}

function fromOrdinal(n: number): CivilDate {
  const at = new Date(n * DAY_MS);
  return { year: at.getUTCFullYear(), month: at.getUTCMonth() + 1, day: at.getUTCDate() };
}

/** Whether the numbers name a real day, e.g. not 31 April or 29 February 2027. */
export function isValid(d: CivilDate): boolean {
  if (!Number.isInteger(d.year) || d.year < 1 || d.year > 9999) return false;
  const back = fromOrdinal(ordinal(d));
  return back.year === d.year && back.month === d.month && back.day === d.day;
}

export function addDays(d: CivilDate, days: number): CivilDate {
  return fromOrdinal(ordinal(d) + days);
}

/** Whole days from a to b; negative when b lies before a. */
export function daysBetween(a: CivilDate, b: CivilDate): number {
  return ordinal(b) - ordinal(a);
}

/** Day of the week, 1 = Monday … 7 = Sunday (ISO 8601). */
export function isoWeekday(d: CivilDate): number {
  // 1970-01-01 was a Thursday (4).
  return ((((ordinal(d) + 3) % 7) + 7) % 7) + 1;
}

/**
 * ISO 8601 week and the year it belongs to: weeks start on Monday and week
 * 1 is the one with the year's first Thursday, so 29 December can already
 * be week 1 of the next year and 1 January still week 52 or 53.
 */
export function isoWeek(d: CivilDate): { week: number; year: number } {
  const thursday = addDays(d, 4 - isoWeekday(d));
  const jan1 = { year: thursday.year, month: 1, day: 1 };
  return { week: Math.floor(daysBetween(jan1, thursday) / 7) + 1, year: thursday.year };
}

/** Easter Sunday of the Gregorian calendar (anonymous Gregorian algorithm). */
export function easter(year: number): CivilDate {
  const a = year % 19;
  const b = Math.floor(year / 100);
  const c = year % 100;
  const d = Math.floor(b / 4);
  const e = b % 4;
  const f = Math.floor((b + 8) / 25);
  const g = Math.floor((b - f + 1) / 3);
  const h = (19 * a + b - d - g + 15) % 30;
  const i = Math.floor(c / 4);
  const k = c % 4;
  const l = (32 + 2 * e + 2 * i - h - k) % 7;
  const m = Math.floor((a + 11 * h + 22 * l) / 451);
  const month = Math.floor((h + l - 7 * m + 114) / 31);
  const day = ((h + l - 7 * m + 114) % 31) + 1;
  return { year, month, day };
}

/** The day as a Date at local noon, for Intl formatting of the weekday and date. */
export function toDate(d: CivilDate): Date {
  return new Date(d.year, d.month - 1, d.day, 12);
}
