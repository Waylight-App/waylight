import { fold } from './smart/text';

/**
 * Frecency of search choices (#159): which result was chosen for which
 * query, weighted by frequency and recency (as in Firefox and Alfred).
 * Each entry keeps one value that halves every HALF_LIFE_MS, so recording
 * a choice is "decay to now, add one" and no timestamps pile up.
 */

export interface FrecencyEntry {
  /** Normalised query prefix the choice was made for. */
  key: string;
  /** Id of the chosen result, e.g. "bookmark:12" or "answer:system.cpu". */
  target: string;
  /** Value at time `at`. */
  value: number;
  /** Milliseconds since the epoch. */
  at: number;
  /**
   * The value split by time slot (SLOTS long, see slotOf), decayed like
   * `value` (#480). Missing in entries stored before; they count the same
   * in every slot.
   */
  slots?: number[];
}

/** Six 4-hour slots on weekdays, six on weekends. */
export const SLOTS = 12;
/** How much a pick in the current slot outweighs one at another time. */
export const SLOT_WEIGHT = 2;
/** Number of suggestions on an empty field. */
export const MAX_SUGGESTIONS = 5;

/** The time slot of a moment in the browser's own time zone. */
export function slotOf(at: Date): number {
  const day = at.getDay(); // 0 = Sunday
  const weekend = day === 0 || day === 6;
  return (weekend ? 6 : 0) + Math.floor(at.getHours() / 4);
}

export const HALF_LIFE_MS = 14 * 24 * 3600 * 1000;
export const MAX_ENTRIES = 200;
/** Longest stored query prefix; longer queries are cut. */
export const MAX_KEY = 24;
/** Score multiplier per unit of boost, and the boost cap. */
export const BOOST_FACTOR = 0.5;
export const MAX_BOOST = 10;
/** Entries below this value are forgotten. */
const MIN_VALUE = 0.05;

/** The query as stored: folded, single spaces, cut to MAX_KEY. */
export function normaliseQuery(query: string): string {
  return fold(query).trim().replace(/\s+/g, ' ').slice(0, MAX_KEY);
}

/** Value of an entry at `now`, after exponential decay. */
export function decayed(entry: FrecencyEntry, now: number): number {
  const age = Math.max(0, now - entry.at);
  return entry.value * 0.5 ** (age / HALF_LIFE_MS);
}

/**
 * The boost of a target for a query: the decayed values of the choices
 * whose query starts with what is typed now, or that what is typed now
 * starts with ("pl" and "plex server" both profit from "plex").
 */
export function boost(entries: readonly FrecencyEntry[], query: string, target: string, now: number): number {
  const q = normaliseQuery(query);
  if (!q) return 0;
  let sum = 0;
  for (const e of entries)
    if (e.target === target && (e.key.startsWith(q) || q.startsWith(e.key))) sum += decayed(e, now);
  return Math.min(sum, MAX_BOOST);
}

/** A match score with the boost applied. */
export function boosted(score: number, b: number): number {
  return score * (1 + BOOST_FACTOR * b);
}

/**
 * Records a choice and returns the new entry list: decays the existing
 * entry and adds one, drops faded entries and evicts the least valuable
 * ones beyond MAX_ENTRIES.
 */
export function record(
  entries: readonly FrecencyEntry[],
  query: string,
  target: string,
  now: number,
): FrecencyEntry[] {
  const key = normaliseQuery(query);
  if (!key || !target) return [...entries];
  const old = entries.find((e) => e.key === key && e.target === target);
  const factor = old ? decayed(old, now) / (old.value || 1) : 0;
  const slots =
    old?.slots?.length === SLOTS ? old.slots.map((v) => v * factor) : new Array<number>(SLOTS).fill(0);
  slots[slotOf(new Date(now))] = (slots[slotOf(new Date(now))] ?? 0) + 1;
  const updated: FrecencyEntry = { key, target, value: (old ? decayed(old, now) : 0) + 1, at: now, slots };
  const rest = entries.filter((e) => e !== old);
  return prune([updated, ...rest], now);
}

/** Drops faded entries and keeps the MAX_ENTRIES most valuable ones. */
export function prune(entries: readonly FrecencyEntry[], now: number): FrecencyEntry[] {
  return entries
    .map((e) => ({ e, v: decayed(e, now) }))
    .filter((x) => x.v >= MIN_VALUE)
    .sort((a, b) => b.v - a.v)
    .slice(0, MAX_ENTRIES)
    .map((x) => x.e);
}

/** Validates stored entries (storage may hold anything). */
export function parseEntries(data: unknown): FrecencyEntry[] {
  if (!Array.isArray(data)) return [];
  return data
    .filter(
      (e): e is FrecencyEntry =>
        typeof e === 'object' &&
        e !== null &&
        typeof e.key === 'string' &&
        typeof e.target === 'string' &&
        typeof e.value === 'number' &&
        Number.isFinite(e.value) &&
        typeof e.at === 'number' &&
        Number.isFinite(e.at),
    )
    .map((e) => (validSlots(e.slots) ? e : { key: e.key, target: e.target, value: e.value, at: e.at }));
}

/** Slot data of the expected shape; anything else is dropped, not fixed. */
function validSlots(slots: unknown): slots is number[] {
  return (
    Array.isArray(slots) &&
    slots.length === SLOTS &&
    slots.every((v) => typeof v === 'number' && Number.isFinite(v) && v >= 0)
  );
}

/** A suggested target for the empty search field and its weight. */
export interface Suggestion {
  target: string;
  score: number;
}

/**
 * What to suggest on an empty field (#480): every target's decayed value
 * summed over all queries it was chosen for, weighted by how much of it
 * falls into the current time slot. A target picked mornings on weekdays
 * comes first on a weekday morning. Entries without slot data count as
 * spread evenly over the day.
 */
export function suggest(
  entries: readonly FrecencyEntry[],
  now: number,
  limit = MAX_SUGGESTIONS,
): Suggestion[] {
  const slot = slotOf(new Date(now));
  const scores = new Map<string, number>();
  for (const e of entries) {
    const total = e.slots?.reduce((a, b) => a + b, 0) ?? 0;
    const share = e.slots && total > 0 ? (e.slots[slot] ?? 0) / total : 1 / SLOTS;
    const score = (decayed(e, now) * (1 + SLOT_WEIGHT * SLOTS * share)) / (1 + SLOT_WEIGHT);
    scores.set(e.target, (scores.get(e.target) ?? 0) + score);
  }
  return [...scores]
    .map(([target, score]) => ({ target, score }))
    .filter((s) => s.score >= MIN_VALUE)
    .sort((a, b) => b.score - a.score || a.target.localeCompare(b.target))
    .slice(0, limit);
}
