import { fold } from './smart/text';

/**
 * Own keywords (#478): an account maps a word such as "nas" to something it
 * opens often. Typing exactly that word ranks the target first, and Enter
 * opens it. Stored on the server, so they work on every device.
 */

/** What a keyword opens: a tile, a bookmark or a widget by id/kind, a web address, or a search. */
export type KeywordKind = 'service' | 'bookmark' | 'widget' | 'url' | 'query';

export interface OwnKeyword {
  keyword: string;
  kind: KeywordKind;
  /** Service or bookmark id, widget kind, http(s) address or search text. */
  target: string;
  /** Name of the target when the keyword was made, shown in the search. */
  label: string;
}

/** Something a keyword can point to, offered by the app (tiles, bookmarks, widgets). */
export interface KeywordTarget {
  kind: 'service' | 'bookmark' | 'widget';
  /** Service or bookmark id, or the widget kind. */
  target: string;
  label: string;
}

export const KEYWORD_KINDS: readonly KeywordKind[] = ['service', 'bookmark', 'widget', 'url', 'query'];

/** The form the server stores: trimmed and lower-cased. */
export const normalizeKeyword = (word: string): string => word.trim().toLowerCase();

/**
 * The keyword the query is exactly, if any. Case and accents do not matter
 * ("NAS", "müll" = "mull"); anything more than the keyword does not match,
 * so "nas status" still reaches the normal search.
 */
export function matchKeyword(list: readonly OwnKeyword[], query: string): OwnKeyword | null {
  const q = fold(query.trim());
  if (!q || /\s/.test(q)) return null;
  return list.find((k) => fold(k.keyword) === q) ?? null;
}
