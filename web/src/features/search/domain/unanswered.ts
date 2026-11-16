import { fold, tokenize } from './smart/text';

/**
 * Queries the search could not answer (#481): an opt-in list in the
 * browser that shows which source is missing next. Only queries that went
 * to the web search without any match are recorded, and the filter below
 * keeps out what could be private. Pure functions only.
 */

export interface UnansweredEntry {
  /** The query as stored: folded, single spaces. */
  query: string;
  count: number;
  /** Milliseconds since the epoch of the last time it was asked. */
  last: number;
}

/** Most queries kept; the oldest go first. */
export const MAX_UNANSWERED = 200;
/** Longest query kept; longer ones are rather pasted text than questions. */
export const MAX_LENGTH = 80;

/** The query as stored. */
export function normalise(query: string): string {
  return fold(query).trim().replace(/\s+/g, ' ');
}

/** "!g cats", "cats !w": a DuckDuckGo-style bang anywhere in the query. */
const BANG = /(^|\s)![\p{L}\p{N}]+/u;
const EMAIL = /\S+@\S+\.\S+/;
const URL_LIKE = /^[a-z][a-z0-9+.-]*:\/\/|^www\./;
/** Six digits in a row (with spaces or dashes): phone, account or card numbers. */
const LONG_NUMBER = /\d(?:[\s-]?\d){5,}/;

/**
 * A word that looks like a password, token or key: long, with letters and
 * digits mixed, or a long hex or base64 run. Real words and plain numbers
 * do not trip it.
 */
function looksSecret(word: string): boolean {
  if (word.length >= 24 && /^[A-Za-z0-9+/=_-]+$/.test(word)) return true;
  if (word.length < 12) return false;
  const classes = [/[a-z]/, /[A-Z]/, /\d/, /[^A-Za-z0-9]/].filter((r) => r.test(word)).length;
  return classes >= 3 || (/\d/.test(word) && /[A-Za-z]/.test(word) && word.length >= 16);
}

/**
 * Whether a query may be remembered. Left out: bangs, anything that looks
 * like a secret, e-mail addresses, web addresses, long numbers and very
 * long text. The raw query is checked, before folding, so case still
 * counts for the secret check.
 */
export function isRecordable(query: string): boolean {
  const q = query.trim();
  if (q.length < 2 || q.length > MAX_LENGTH) return false;
  if (BANG.test(q) || EMAIL.test(q) || URL_LIKE.test(q.toLowerCase()) || LONG_NUMBER.test(q)) return false;
  return !q.split(/\s+/).some(looksSecret);
}

/** Adds a query (or counts it again) and keeps the MAX_UNANSWERED newest. */
export function addUnanswered(
  entries: readonly UnansweredEntry[],
  query: string,
  now: number,
): UnansweredEntry[] {
  if (!isRecordable(query)) return [...entries];
  const q = normalise(query);
  const old = entries.find((e) => e.query === q);
  const updated = { query: q, count: (old?.count ?? 0) + 1, last: now };
  return [updated, ...entries.filter((e) => e !== old)]
    .sort((a, b) => b.last - a.last)
    .slice(0, MAX_UNANSWERED);
}

/** Queries that share a word, for the settings list. */
export interface UnansweredGroup {
  /** The shared word, e.g. "zug" for "zug nach köln" and "zug morgen". */
  word: string;
  /** How often the queries of this group were asked in total. */
  count: number;
  queries: UnansweredEntry[];
}

/**
 * Groups the queries by their first meaningful word (stop words left
 * out), most asked first. A query without such a word forms its own group.
 */
export function groupUnanswered(entries: readonly UnansweredEntry[]): UnansweredGroup[] {
  const groups = new Map<string, UnansweredGroup>();
  for (const e of entries) {
    const word = tokenize(e.query)[0] ?? e.query;
    const g = groups.get(word) ?? { word, count: 0, queries: [] };
    g.count += e.count;
    g.queries.push(e);
    groups.set(word, g);
  }
  for (const g of groups.values()) g.queries.sort((a, b) => b.count - a.count || b.last - a.last);
  return [...groups.values()].sort((a, b) => b.count - a.count || a.word.localeCompare(b.word));
}

/** Validates stored entries (storage may hold anything). */
export function parseUnanswered(data: unknown): UnansweredEntry[] {
  if (!Array.isArray(data)) return [];
  return data.filter(
    (e): e is UnansweredEntry =>
      typeof e === 'object' &&
      e !== null &&
      typeof e.query === 'string' &&
      Number.isInteger(e.count) &&
      e.count > 0 &&
      typeof e.last === 'number' &&
      Number.isFinite(e.last),
  );
}
