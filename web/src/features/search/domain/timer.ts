import { fold } from './smart/text';

/**
 * Timers and alarms from the search field (#474): "timer 10 min",
 * "timer 1h30 pizza", "wecker 17:30", "alarm 6:45 aufstehen". Parsing only;
 * scheduling lives behind the TimerScheduler port.
 */
export interface TimerRequest {
  kind: 'timer' | 'alarm';
  /** When it runs out, in milliseconds since the epoch. */
  end: number;
  /** Optional text after the time, e.g. "pizza". */
  label: string;
  /** An alarm whose time already passed today rings tomorrow. */
  tomorrow: boolean;
}

/** A running timer as stored in the browser. */
export interface Timer extends TimerRequest {
  id: string;
}

const TIMER_WORDS = new Set(['timer', 'countdown', 'kurzzeitwecker', 'eieruhr']);
const ALARM_WORDS = new Set(['alarm', 'wecker', 'weck', 'wake']);

/** Longest timer accepted: one day. */
export const MAX_DURATION_MS = 24 * 3600 * 1000;

const UNIT_MS: Record<string, number> = {
  s: 1000,
  sec: 1000,
  sek: 1000,
  sekunde: 1000,
  sekunden: 1000,
  second: 1000,
  seconds: 1000,
  m: 60_000,
  min: 60_000,
  mins: 60_000,
  minute: 60_000,
  minuten: 60_000,
  minutes: 60_000,
  h: 3_600_000,
  std: 3_600_000,
  stunde: 3_600_000,
  stunden: 3_600_000,
  hour: 3_600_000,
  hours: 3_600_000,
};

/**
 * Reads a duration from the start of the words: "10 min", "10min", "1h30",
 * "1h 30m", "90" (minutes), "1,5 h". Returns the duration and how many
 * words it used, or null.
 */
export function parseDuration(words: readonly string[]): { ms: number; used: number } | null {
  let ms = 0;
  let used = 0;
  let any = false;
  while (used < words.length) {
    const word = words[used] ?? '';
    // "1h30", "1h30m", "2h", "45min", "90s"
    const compact = /^(\d+(?:[.,]\d+)?)([a-z]+)(\d+)?([a-z]*)$/.exec(word);
    if (compact) {
      const unit = UNIT_MS[compact[2] ?? ''];
      if (!unit) break;
      ms += Number((compact[1] ?? '').replace(',', '.')) * unit;
      if (compact[3]) {
        // "1h30" → the trailing number uses the next smaller unit.
        const next = UNIT_MS[compact[4] ?? ''] ?? (unit === 3_600_000 ? 60_000 : 1000);
        ms += Number(compact[3]) * next;
      }
      used++;
      any = true;
      continue;
    }
    const num = /^\d+(?:[.,]\d+)?$/.test(word) ? Number(word.replace(',', '.')) : null;
    if (num === null) break;
    const unit = UNIT_MS[words[used + 1] ?? ''];
    if (unit) {
      ms += num * unit;
      used += 2;
    } else {
      // A bare number is minutes, but only as the last part of the duration.
      if (any) break;
      ms += num * 60_000;
      used++;
    }
    any = true;
  }
  return any && ms > 0 ? { ms: Math.round(ms), used } : null;
}

/** "17:30", "6:45", "6.45", "17 uhr", "17h" → hours and minutes, or null. */
export function parseClock(words: readonly string[]): { h: number; m: number; used: number } | null {
  const first = words[0] ?? '';
  let m = /^(\d{1,2})[:.](\d{2})(?:uhr)?$/.exec(first);
  if (m) return valid(Number(m[1]), Number(m[2]), words[1] === 'uhr' ? 2 : 1);
  m = /^(\d{1,2})(?:uhr|h)?$/.exec(first);
  if (m && (first.endsWith('uhr') || first.endsWith('h') || words[1] === 'uhr'))
    return valid(Number(m[1]), 0, words[1] === 'uhr' ? 2 : 1);
  return null;
}

function valid(h: number, m: number, used: number) {
  return h < 24 && m < 60 ? { h, m, used } : null;
}

/**
 * Parses a timer or alarm request; null when the text is not one. An
 * alarm time that already passed today is set for tomorrow.
 */
export function parseTimer(query: string, now: Date): TimerRequest | null {
  const words = fold(query).trim().split(/\s+/).filter(Boolean);
  const head = words[0] ?? '';
  const rest = words.slice(1);
  if (TIMER_WORDS.has(head)) {
    const d = parseDuration(rest);
    if (!d || d.ms > MAX_DURATION_MS) return null;
    return { kind: 'timer', end: now.getTime() + d.ms, label: label(query, 1 + d.used), tomorrow: false };
  }
  if (ALARM_WORDS.has(head)) {
    const c = parseClock(rest.filter((w) => w !== 'um' && w !== 'at'));
    if (!c) return null;
    const skipped = rest.findIndex((w) => w !== 'um' && w !== 'at');
    const at = new Date(now);
    at.setHours(c.h, c.m, 0, 0);
    const tomorrow = at.getTime() <= now.getTime();
    if (tomorrow) at.setDate(at.getDate() + 1);
    return { kind: 'alarm', end: at.getTime(), label: label(query, 1 + skipped + c.used), tomorrow };
  }
  return null;
}

/** The words after the first `skip` words, as typed (not folded). */
function label(query: string, skip: number): string {
  return query.trim().split(/\s+/).slice(skip).join(' ');
}

/** Whether the text asks for the list of running timers. */
export function asksForTimers(query: string): boolean {
  const q = fold(query).trim();
  return TIMER_WORDS.has(q) || ALARM_WORDS.has(q) || q === 'timers';
}

/** Remaining time as "1:05:09" or "4:59". */
export function formatRemaining(ms: number): string {
  const total = Math.max(0, Math.ceil(ms / 1000));
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  const pad = (n: number) => String(n).padStart(2, '0');
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`;
}

/** Timers that have not run out, soonest first. */
export function running(timers: readonly Timer[], now: number): Timer[] {
  return timers.filter((t) => t.end > now).sort((a, b) => a.end - b.end);
}
