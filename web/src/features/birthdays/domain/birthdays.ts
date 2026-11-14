import type { MessageKey } from '../../../shared/i18n';

/** Where an entry comes from (#287); address book entries are read-only here. */
export type BirthdaySource = 'manual' | 'vcard' | 'carddav';

/** A birthday with its next occurrence. */
export interface Birthday {
  id: string;
  name: string;
  month: number;
  day: number;
  /** 0 when the year of birth is unknown. */
  year: number;
  source: BirthdaySource;
  /** Next birthday as YYYY-MM-DD (today included). */
  next: string;
  inDays: number;
  /** Age reached on `next`; 0 when the year is unknown. */
  age: number;
}

export interface BirthdayInput {
  name: string;
  /** YYYY-MM-DD from a date input; the year part is ignored when `yearKnown` is false. */
  date: string;
  yearKnown: boolean;
}

export interface BirthdaySettings {
  remindEnabled: boolean;
  remindAt: string;
  cardDavUrl: string;
  username: string;
  hasPassword: boolean;
}

export interface BirthdaySettingsInput {
  remindEnabled: boolean;
  remindAt: string;
  cardDavUrl: string;
  username: string;
  /** Empty keeps the stored password. */
  password: string;
}

export interface ImportReport {
  added: number;
  duplicates: number;
}

export const MAX_NAME_LENGTH = 100;
/** Entries the widget shows unless configured otherwise. */
export const DEFAULT_LIMIT = 5;
const MAX_LIMIT = 50;

/** Widget configuration: how many birthdays to show. */
export function parseConfig(raw: Record<string, unknown>): { limit: number } {
  const v = raw['limit'];
  const limit = typeof v === 'number' && Number.isInteger(v) && v >= 1 && v <= MAX_LIMIT ? v : DEFAULT_LIMIT;
  return { limit };
}

/** Converts the form into the API payload, or returns the message key of the first problem. */
export function toPayload(
  input: BirthdayInput,
): { name: string; month: number; day: number; year: number } | MessageKey {
  const name = input.name.trim().replace(/\s+/g, ' ');
  if (!name || name.length > MAX_NAME_LENGTH) return 'birthdays.invalid_name';
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(input.date);
  if (!m) return 'birthdays.invalid_date';
  return { name, year: input.yearKnown ? Number(m[1]) : 0, month: Number(m[2]), day: Number(m[3]) };
}

/** "today", "tomorrow" or "in n days" as a message key with its parameters. */
export function whenKey(inDays: number): { key: MessageKey; n?: number } {
  if (inDays <= 0) return { key: 'birthdays.today' };
  if (inDays === 1) return { key: 'birthdays.tomorrow' };
  return { key: 'birthdays.in_days', n: inDays };
}

/** Birthdays within this many days are highlighted. */
export const SOON_DAYS = 7;
