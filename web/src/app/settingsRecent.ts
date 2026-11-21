import { pushRecent } from './settingsNav';

/**
 * "Recently changed" on the settings start page (#600) is kept in this
 * browser only: it is a convenience, not account data, and needs no API.
 */
const KEY = 'waylight.settings.recent';

/** The section ids changed last, newest first; empty when storage is blocked. */
export function loadRecent(storage?: Pick<Storage, 'getItem'>): string[] {
  try {
    // Reading localStorage itself throws where site data is blocked.
    const raw: unknown = JSON.parse((storage ?? globalThis.localStorage)?.getItem(KEY) ?? '[]');
    return Array.isArray(raw) ? raw.filter((x): x is string => typeof x === 'string') : [];
  } catch {
    return [];
  }
}

/** Records a change in section id and returns the new list. */
export function rememberChange(id: string, storage?: Pick<Storage, 'getItem' | 'setItem'>): string[] {
  const next = pushRecent(loadRecent(storage), id);
  try {
    (storage ?? globalThis.localStorage)?.setItem(KEY, JSON.stringify(next));
  } catch {
    // Private windows may refuse storage; the list then lives for this visit.
  }
  return next;
}
