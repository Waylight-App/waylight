import type { Visibility } from '../../../shared/sharing/sharing';

/** What a meter measures. */
export const KINDS = ['electricity', 'gas', 'water', 'heat', 'other'] as const;
export type Kind = (typeof KINDS)[number];

/** The counter value on one day. */
export interface Reading {
  /** YYYY-MM-DD. */
  day: string;
  value: number;
}

/** A household meter with its newest reading (#689). */
export interface Meter {
  id: string;
  name: string;
  kind: Kind;
  unit: string;
  last: Reading | null;
  visibility: Visibility;
  mine: boolean;
}

/** Consumption of one year; null where the readings do not cover it. */
export interface Consumption {
  year: number;
  /** Twelve entries, January first. */
  months: (number | null)[];
  total: number | null;
  /** True when the total covers the whole year. */
  complete: boolean;
}

/** A meter with its readings, newest first, and the consumption of one year. */
export interface MeterDetail {
  meter: Meter;
  readings: Reading[];
  consumption: Consumption;
  /** Calendar years with readings, the latest first. */
  years: number[];
}

/** What the dialog edits. */
export interface MeterInput {
  name: string;
  kind: Kind;
  unit: string;
}

/** What the quick entry form edits; value is null while the field is empty. */
export interface ReadingInput {
  day: string;
  value: number | null;
}

/** Message keys of validateMeter and validateReading. */
export type MeterError =
  | 'meters.error.name'
  | 'meters.error.unit'
  | 'meters.error.day'
  | 'meters.error.value'
  | 'meters.error.lower'
  | 'meters.error.higher';

const toKind = (s: string): Kind => ((KINDS as readonly string[]).includes(s) ? (s as Kind) : 'other');

/** Narrows the kind the server sends; an unknown one reads as "other". */
export const parseKind = toKind;

/** Validates the meter dialog and returns a message key or null. */
export function validateMeter(input: MeterInput): MeterError | null {
  const name = input.name.trim();
  if (!name || [...name].length > 60) return 'meters.error.name';
  const unit = input.unit.trim();
  if (!unit || [...unit].length > 12) return 'meters.error.unit';
  return null;
}

/** Today as YYYY-MM-DD in the browser's time zone. */
export function today(now: Date = new Date()): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${now.getFullYear()}-${p(now.getMonth() + 1)}-${p(now.getDate())}`;
}

/**
 * Validates a reading against the earlier ones, like the server does: a
 * counter only counts up, so a reading may not be lower than an earlier one
 * or higher than a later one. A reading on a day that has one replaces it.
 */
export function validateReading(
  input: ReadingInput,
  existing: readonly Reading[],
  now: string = today(),
): MeterError | null {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(input.day) || input.day > now) return 'meters.error.day';
  const value = input.value;
  if (value === null || !Number.isFinite(value) || value < 0) return 'meters.error.value';
  for (const r of existing) {
    if (r.day < input.day && r.value > value) return 'meters.error.lower';
    if (r.day > input.day && r.value < value) return 'meters.error.higher';
  }
  return null;
}

/** Years to offer in the year selector: those with readings and the current one. */
export function yearChoices(years: readonly number[], current: number): number[] {
  return [...new Set([current, ...years])].sort((a, b) => b - a);
}
