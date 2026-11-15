import { t } from '../../../shared/i18n';
/** A to-do item. `due` is a date string (YYYY-MM-DD) or ''. */
export interface Todo {
  id: string;
  title: string;
  due: string;
  done: boolean;
  overdue: boolean;
  /** Recurrence (#68); kind '' = one-off. */
  repeat: Repeat;
  /** Reminder (#68); at '' = none. */
  reminder: Reminder;
  /** Stored version (#80); 0 = only on this device so far (offline). */
  version: number;
  /** Who else may see it (ADR 0018); empty for external lists. */
  visibility?: string;
  /** The viewer owns the to-do and may share it. */
  mine?: boolean;
}

/** How a to-do recurs: fixed schedule or N days after completion. */
export type RepeatKind = '' | 'daily' | 'weekly' | 'monthly' | 'after';

export interface Repeat {
  kind: RepeatKind;
  /** Interval: every N days/weeks/months, or N days after completion. */
  every: number;
  /** Weekdays of weekly rules (0 = Sunday). */
  days: number[];
}

/** A push at a time of day on the due date or some days before. */
export interface Reminder {
  /** "HH:MM" or '' for none. */
  at: string;
  daysBefore: number;
}

export const NO_REPEAT: Repeat = { kind: '', every: 1, days: [] };
export const NO_REMINDER: Reminder = { at: '', daysBefore: 0 };

/** Limits; mirror the backend. */
export const MAX_REPEAT_EVERY = 365;
export const MAX_REMINDER_DAYS_BEFORE = 30;

/** Validates recurrence and reminder; returns an error message or null. */
export function validateSchedule(due: string, repeat: Repeat, reminder: Reminder): string | null {
  if (repeat.kind !== '') {
    if (!Number.isInteger(repeat.every) || repeat.every < 1 || repeat.every > MAX_REPEAT_EVERY)
      return t('todos.repeat_invalid_every', { max: MAX_REPEAT_EVERY });
    if (repeat.kind === 'weekly' && repeat.days.length === 0) return t('todos.repeat_needs_days');
  }
  if (reminder.at !== '') {
    if (due === '') return t('todos.reminder_needs_due');
    if (!/^([01]\d|2[0-3]):[0-5]\d$/.test(reminder.at)) return t('todos.reminder_invalid_time');
    if (
      !Number.isInteger(reminder.daysBefore) ||
      reminder.daysBefore < 0 ||
      reminder.daysBefore > MAX_REMINDER_DAYS_BEFORE
    )
      return t('todos.reminder_invalid_days', { max: MAX_REMINDER_DAYS_BEFORE });
  }
  return null;
}

/** Short description of a recurrence, e.g. "Every 2 weeks on Mon, Thu". */
export function describeRepeat(r: Repeat, dayName: (d: number) => string): string {
  const n = r.every;
  switch (r.kind) {
    case 'daily':
      return n === 1 ? t('todos.repeat_daily') : t('todos.repeat_every_days', { n });
    case 'weekly': {
      const days = [1, 2, 3, 4, 5, 6, 0]
        .filter((d) => r.days.includes(d))
        .map(dayName)
        .join(', ');
      return n === 1 ? t('todos.repeat_weekly_on', { days }) : t('todos.repeat_every_weeks_on', { n, days });
    }
    case 'monthly':
      return n === 1 ? t('todos.repeat_monthly') : t('todos.repeat_every_months', { n });
    case 'after':
      return t('todos.repeat_after', { n });
    default:
      return '';
  }
}

/** Short description of a reminder, e.g. "08:00, 1 day before". */
export function describeReminder(r: Reminder): string {
  if (r.at === '') return '';
  if (r.daysBefore === 0) return t('todos.reminder_on_day', { at: r.at });
  return t('todos.reminder_days_before', { at: r.at, n: r.daysBefore });
}

/** Configuration of the to-do widget. */
export interface TodosConfig {
  showDone: boolean;
  /** List shown: '' = the built-in list, else a connected list's ID (#169). */
  source: string;
}

/** Reads a stored widget configuration. */
export function parseTodosConfig(raw: Record<string, unknown>): TodosConfig {
  const source =
    typeof raw['source'] === 'string' && /^[A-Za-z0-9_-]{1,64}$/.test(raw['source']) ? raw['source'] : '';
  return { showDone: raw['showDone'] === true, source: source === 'local' ? '' : source };
}

/** What a list can store; the UI hides the rest (#169). */
export interface Capabilities {
  ordering: boolean;
  repeat: boolean;
  reminders: boolean;
}

/** Everything the built-in list supports. */
export const LOCAL_CAPABILITIES: Capabilities = { ordering: true, repeat: true, reminders: true };

/** Kinds of connected lists: CalDAV (#169), Todoist (#170) and the services you sign in to (#171). */
export type SourceKind = 'caldav' | 'todoist' | 'microsoft' | 'google' | 'ticktick';
export const SOURCE_KINDS: readonly SourceKind[] = ['caldav', 'todoist', 'microsoft', 'google', 'ticktick'];

/** Kinds connected by signing in with your own app registration (OAuth). */
export const isOAuthKind = (k: SourceKind): boolean =>
  k === 'microsoft' || k === 'google' || k === 'ticktick';

/** Where to register an app for each OAuth kind (shown in the settings). */
export const OAUTH_CONSOLES: Partial<Record<SourceKind, string>> = {
  microsoft: 'https://entra.microsoft.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade',
  google: 'https://console.cloud.google.com/apis/credentials',
  ticktick: 'https://developer.ticktick.com/manage',
};

/** A connected external to-do list; the password or token is never sent back. */
export interface TodoSource {
  id: string;
  name: string;
  kind: SourceKind;
  url: string;
  username: string;
  /** Todoist: the project ID. */
  listId: string;
  capabilities: Capabilities;
}

/** A task list found on a server: a CalDAV collection (url) or a Todoist project (listId). */
export interface DiscoveredList {
  url: string;
  listId: string;
  name: string;
}

/** Capabilities of the list a widget shows; unknown lists support nothing extra. */
export function capabilitiesOf(source: string, sources: readonly TodoSource[]): Capabilities {
  if (source === '') return LOCAL_CAPABILITIES;
  return (
    sources.find((s) => s.id === source)?.capabilities ?? { ordering: false, repeat: false, reminders: false }
  );
}

/**
 * Formats a due date relative to today ("Today", "Tomorrow", weekday within
 * a week, otherwise the date).
 */
export function formatDue(due: string, today: Date, locale?: string): string {
  if (!due) return '';
  const [y, m, d] = due.split('-').map(Number);
  if (!y || !m || !d) return due;
  const date = new Date(y, m - 1, d);
  const start = new Date(today.getFullYear(), today.getMonth(), today.getDate());
  const days = Math.round((date.getTime() - start.getTime()) / 86_400_000);
  if (days === 0) return t('time.today');
  if (days === 1) return t('time.tomorrow');
  if (days === -1) return t('time.yesterday');
  if (days > 1 && days < 7) return date.toLocaleDateString(locale, { weekday: 'long' });
  return date.toLocaleDateString(locale, { day: 'numeric', month: 'short' });
}
