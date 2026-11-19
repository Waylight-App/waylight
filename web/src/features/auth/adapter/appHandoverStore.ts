import type { AppHandover } from '../domain/appHandover';

// The handover waits in the tab's session storage while the person signs in:
// single sign-on leaves the page and comes back, so memory is not enough.
const key = 'waylight.appHandover';
const maxAge = 10 * 60 * 1000;

/** Remembers the handover of this tab. */
export function saveAppHandover(handover: AppHandover, now = Date.now()): void {
  try {
    sessionStorage.setItem(key, JSON.stringify({ ...handover, at: now }));
  } catch {
    // Without storage the handover only works if no redirect is needed.
  }
}

/** The pending handover of this tab, if it is less than ten minutes old. */
export function pendingAppHandover(now = Date.now()): AppHandover | null {
  try {
    const raw = sessionStorage.getItem(key);
    if (!raw) return null;
    const { challenge, state, at } = JSON.parse(raw) as AppHandover & { at: number };
    return now - at < maxAge ? { challenge, state } : null;
  } catch {
    return null;
  }
}

/** Forgets the handover once it was answered. */
export function clearAppHandover(): void {
  try {
    sessionStorage.removeItem(key);
  } catch {
    // Nothing stored.
  }
}
