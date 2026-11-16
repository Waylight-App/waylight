import type { ToolEnv } from './detect';

/** A point in time as Unix timestamp and date (#477). */
export interface TimestampResult {
  kind: 'timestamp';
  /** Seconds since 1970-01-01 UTC. */
  seconds: number;
  date: Date;
  /** Which way the user asked: from a timestamp, or to one. */
  direction: 'from-unix' | 'to-unix';
}

const TO_UNIX = /^(.+?)\s+(?:to|in|als|nach)\s+unix$/i;

/**
 * "1700000000" (seconds) or "1700000000000" (milliseconds) gives the date;
 * "unix" or "unix now" the current timestamp; "2026-11-16 12:00 to unix" or
 * "16.11.2026 12:00 in unix" the timestamp of a date, read in local time
 * unless it ends in "Z" or "UTC".
 */
export function timestamp(query: string, env: ToolEnv): TimestampResult | null {
  const q = query.trim();
  if (/^\d{9,13}$/.test(q)) {
    const ms = q.length >= 12 ? Number(q) : Number(q) * 1000;
    return result(new Date(ms), 'from-unix');
  }
  if (/^unix(?:\s+\S+)?$/i.test(q)) return result(env.now(), 'to-unix');
  const text = TO_UNIX.exec(q)?.[1];
  const date = text === undefined ? null : parseDate(text);
  return date ? result(date, 'to-unix') : null;
}

function result(date: Date, direction: TimestampResult['direction']): TimestampResult | null {
  const ms = date.getTime();
  if (!Number.isFinite(ms)) return null;
  return { kind: 'timestamp', seconds: Math.floor(ms / 1000), date, direction };
}

/** ISO-like "2026-11-16[ T]12:00[:30][Z|UTC]" or German "16.11.2026 12:00". */
export function parseDate(text: string): Date | null {
  const t = text.trim();
  const iso = /^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{1,2}):(\d{2})(?::(\d{2}))?)?\s*(z|utc)?$/i.exec(t);
  const de = /^(\d{1,2})\.(\d{1,2})\.(\d{4})(?:,?\s+(\d{1,2}):(\d{2})(?::(\d{2}))?)?\s*(utc)?$/i.exec(t);
  const m = iso ?? de;
  if (!m) return null;
  const [y, mo, d] = iso ? [m[1], m[2], m[3]] : [m[3], m[2], m[1]];
  const parts = [y, mo, d, m[4] ?? '0', m[5] ?? '0', m[6] ?? '0'].map(Number) as [
    number,
    number,
    number,
    number,
    number,
    number,
  ];
  const [year, month, day, hour, minute, second] = parts;
  if (month < 1 || month > 12 || day < 1 || day > 31 || hour > 23 || minute > 59 || second > 59) return null;
  const date = m[7]
    ? new Date(Date.UTC(year, month - 1, day, hour, minute, second))
    : new Date(year, month - 1, day, hour, minute, second);
  // Rejects dates like 31.02., which Date would roll over into March.
  const check = m[7] ? date.getUTCDate() : date.getDate();
  return check === day ? date : null;
}
