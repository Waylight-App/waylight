import { parseUnanswered } from '../domain/unanswered';
import type { UnansweredState, UnansweredStore } from '../usecase/unansweredQueries';

const KEY = 'waylight.search-unanswered';

/**
 * Keeps the queries without an answer in localStorage (#481), never on the
 * server. Off unless the stored state says otherwise. Without storage
 * (private windows, blocked site data) nothing is remembered.
 */
export class LocalUnansweredStore implements UnansweredStore {
  constructor(private readonly storage: () => Storage | undefined = () => globalThis.localStorage) {}

  load(): UnansweredState {
    try {
      const raw = this.storage()?.getItem(KEY);
      if (!raw) return { enabled: false, entries: [] };
      const data: unknown = JSON.parse(raw);
      const obj = typeof data === 'object' && data !== null ? (data as Record<string, unknown>) : {};
      const enabled = obj['enabled'] === true;
      return { enabled, entries: enabled ? parseUnanswered(obj['entries']) : [] };
    } catch {
      return { enabled: false, entries: [] };
    }
  }

  save(state: UnansweredState): void {
    try {
      const storage = this.storage();
      // Off means nothing is kept, not even an empty list.
      if (!state.enabled) storage?.removeItem(KEY);
      else storage?.setItem(KEY, JSON.stringify(state));
    } catch {
      // Storage full or blocked: the list simply is not kept.
    }
  }
}
