import type { Timer } from '../features/search/domain/timer';

/**
 * Timers that ran out and wait to be dismissed (#474). Kept in a tiny module
 * so the shell can show the toast without loading the timer code on every
 * page; app/timers.ts fills it.
 */
export const timerAlerts = $state<{ list: Timer[] }>({ list: [] });

/** Whether this browser has stored timers, so the shell loads them after a reload. */
export function hasStoredTimers(): boolean {
  try {
    const raw = localStorage.getItem('waylight.timers');
    return raw !== null && raw !== '[]';
  } catch {
    return false;
  }
}
