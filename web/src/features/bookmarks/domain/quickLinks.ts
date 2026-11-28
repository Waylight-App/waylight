import type { Bookmark } from './bookmark';

/**
 * Quick links widget (#859): a handful of favourite bookmarks as tiles; the
 * digits 1 to 9 open the first nine. Pure rules, the bookmarks come from the
 * bookmarks use cases.
 */

/** Most tiles one widget shows. */
export const MAX_QUICK_LINKS = 12;
/** Links that get a digit shortcut. */
export const SHORTCUT_LINKS = 9;

/** Stored configuration: the chosen bookmarks in the order they were chosen. */
export interface QuickLinksConfig {
  ids: string[];
}

/** Reads the stored configuration; anything unexpected counts as no choice. */
export function parseQuickLinksConfig(raw: Record<string, unknown>): QuickLinksConfig {
  const ids = Array.isArray(raw.ids) ? raw.ids.filter((x): x is string => typeof x === 'string') : [];
  return { ids: [...new Set(ids)].slice(0, MAX_QUICK_LINKS) };
}

/**
 * The tiles to show: the chosen bookmarks in their order, skipping deleted
 * ones. Without a choice the first bookmarks fill the widget, so it is
 * useful right after adding it.
 */
export function quickLinks(bookmarks: readonly Bookmark[], cfg: QuickLinksConfig): Bookmark[] {
  if (cfg.ids.length === 0) return bookmarks.slice(0, MAX_QUICK_LINKS);
  const byId = new Map(bookmarks.map((b) => [b.id, b]));
  return cfg.ids.flatMap((id) => byId.get(id) ?? []);
}

/** Adds or removes a bookmark from the choice; adding stops at the maximum. */
export function toggleQuickLink(ids: readonly string[], id: string): string[] {
  if (ids.includes(id)) return ids.filter((x) => x !== id);
  return ids.length >= MAX_QUICK_LINKS ? [...ids] : [...ids, id];
}

/** Moves a chosen bookmark one place up (-1) or down (+1). */
export function moveQuickLink(ids: readonly string[], id: string, by: -1 | 1): string[] {
  const from = ids.indexOf(id);
  const to = from + by;
  if (from < 0 || to < 0 || to >= ids.length) return [...ids];
  const next = [...ids];
  [next[from], next[to]] = [next[to] as string, next[from] as string];
  return next;
}
