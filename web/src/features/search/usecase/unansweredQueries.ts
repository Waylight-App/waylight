import {
  addUnanswered,
  groupUnanswered,
  isRecordable,
  type UnansweredEntry,
  type UnansweredGroup,
} from '../domain/unanswered';

/** What the store keeps: the switch and the list. */
export interface UnansweredState {
  enabled: boolean;
  entries: UnansweredEntry[];
}

/** Output port: where the list lives. Only ever the browser (#481). */
export interface UnansweredStore {
  load(): UnansweredState;
  save(state: UnansweredState): void;
}

/**
 * Remembers queries without an answer (#481). Off by default; the caller
 * decides when a query counts as unanswered (it went to the web search and
 * matched nothing). Switching off deletes the list.
 */
export class UnansweredQueries {
  private state: UnansweredState;

  constructor(
    private readonly store: UnansweredStore,
    private readonly now: () => number = Date.now,
  ) {
    this.state = store.load();
  }

  get enabled(): boolean {
    return this.state.enabled;
  }

  get size(): number {
    return this.state.entries.length;
  }

  /** Records a query that ended in the web search without any match. */
  record(query: string): void {
    if (!this.state.enabled || !isRecordable(query)) return;
    this.update({ ...this.state, entries: addUnanswered(this.state.entries, query, this.now()) });
  }

  /** The list grouped by shared words, most asked first. */
  groups(): UnansweredGroup[] {
    return groupUnanswered(this.state.entries);
  }

  setEnabled(enabled: boolean): void {
    this.update({ enabled, entries: enabled ? this.state.entries : [] });
  }

  clear(): void {
    this.update({ ...this.state, entries: [] });
  }

  private update(next: UnansweredState) {
    this.state = next;
    this.store.save(next);
  }
}
