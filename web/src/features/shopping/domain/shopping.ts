/** One entry of the shopping list. */
export interface ShoppingItem {
  id: string;
  summary: string;
  done: boolean;
}

/** Connection settings as shown in the UI (the password is never returned). */
export interface ShoppingSettings {
  configured: boolean;
  listUrl: string;
  username: string;
  hasPassword: boolean;
}

/** Prefill values derived from another CalDAV configuration. */
export interface ShoppingSuggestion {
  listUrl: string;
  username: string;
}

/**
 * Suggests a task list URL next to a calendar collection, e.g.
 * …/calendars/me/personal/ → …/calendars/me/shopping/.
 */
export function suggestListUrl(calendarUrl: string): string {
  const m = /^(.*\/)[^/]+\/?$/.exec(calendarUrl.trim());
  return m ? `${m[1]}shopping/` : '';
}

/** A shopping list kept on Waylight (#471). */
export interface ShoppingList {
  id: string;
  name: string;
  /** "private", "everyone" or "group:<id>" (ADR 0018). */
  visibility: string;
  /** True for the account's own lists; shared lists can be used, not managed. */
  mine: boolean;
  /** Number of open items. */
  open: number;
}

/** What the widget can show: the lists on Waylight and the CalDAV list. */
export interface ShoppingOverview {
  lists: ShoppingList[];
  /** True when a CalDAV task list is connected. */
  caldav: boolean;
  /** The saved choice: a list id, "caldav", or "" when nothing was chosen. */
  current: string;
}

/** Value of the CalDAV list in a choice of lists. */
export const CALDAV = 'caldav';

/** Export formats of a list on Waylight. */
export type ExportFormat = 'text' | 'ics';

/**
 * The list the widget shows. A saved choice wins while it still exists;
 * otherwise the first own list, then a connected CalDAV list (which keeps
 * the widget unchanged for accounts that only ever used CalDAV), then the
 * first list shared with the account. "" means there is nothing to show.
 */
export function resolveCurrent(o: ShoppingOverview): string {
  if (o.current === CALDAV ? o.caldav : o.lists.some((l) => l.id === o.current)) return o.current;
  const own = o.lists.find((l) => l.mine);
  if (own) return own.id;
  if (o.caldav) return CALDAV;
  return o.lists[0]?.id ?? '';
}

/** Download address of a list export; the session cookie authorises it. */
export function exportUrl(listId: string, format: ExportFormat): string {
  return `/api/shopping/lists/${encodeURIComponent(listId)}/export?format=${format}`;
}

/** Applies a done change without mutating the input (optimistic update). */
export function withDone(items: ShoppingItem[], id: string, done: boolean): ShoppingItem[] {
  return items.map((i) => (i.id === id ? { ...i, done } : i));
}
