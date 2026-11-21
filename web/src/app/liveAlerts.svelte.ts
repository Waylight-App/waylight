import type { LiveNotification } from '../features/notifications/domain/live';

/**
 * Live notifications waiting to be dismissed (#583). A tiny module, so the
 * shell shows the toasts without loading the stream code on every page.
 */
export const liveAlerts = $state<{ list: (LiveNotification & { key: number })[] }>({ list: [] });

/** At most this many toasts are shown; older ones make room. */
const MAX_TOASTS = 4;
let next = 0;

/** Adds a toast; info and warning disappear by themselves after a while. */
export function addLiveAlert(n: LiveNotification): void {
  const key = ++next;
  liveAlerts.list = [...liveAlerts.list, { ...n, key }].slice(-MAX_TOASTS);
  if (n.severity !== 'critical') setTimeout(() => dismissLiveAlert(key), 15_000);
}

/** Removes one toast. */
export function dismissLiveAlert(key: number): void {
  liveAlerts.list = liveAlerts.list.filter((a) => a.key !== key);
}
