import { type AttentionItem, sameItems } from '../domain/attention';
import type { AttentionUseCases } from '../usecase/attentionUseCases';

/**
 * The attention list as reactive state, shared by the top bar dot and the
 * page so both show the same entries and a fix shows up in both at once.
 */
export class AttentionState {
  items = $state<AttentionItem[]>([]);
  /** True once the first answer arrived. */
  loaded = $state(false);
  /** True while "Check again" asks the services. */
  checking = $state(false);
  error = $state('');
  #uc: AttentionUseCases | null = null;
  #pending: Promise<void> | null = null;

  /** Connects the state to the server; without it every call is a no-op. */
  attach(uc: AttentionUseCases | null): void {
    this.#uc = uc;
    if (!uc) this.items = [];
  }

  get count(): number {
    return this.items.length;
  }

  /** Loads the list; concurrent calls share one request. */
  refresh(): Promise<void> {
    const uc = this.#uc;
    if (!uc) return Promise.resolve();
    this.#pending ??= this.#load(() => uc.list()).finally(() => (this.#pending = null));
    return this.#pending;
  }

  /** Asks every service again, e.g. after a setting changed or on "Check again". */
  async check(): Promise<void> {
    const uc = this.#uc;
    if (!uc || this.checking) return;
    this.checking = true;
    try {
      await this.#load(() => uc.check());
    } finally {
      this.checking = false;
    }
  }

  async #load(fetch: () => Promise<AttentionItem[]>): Promise<void> {
    try {
      const next = await fetch();
      // Keep the old array when nothing changed, so an open page does not
      // re-render (and lose focus) on every poll.
      if (!sameItems(this.items, next)) this.items = next;
      this.error = '';
    } catch (err) {
      this.error = err instanceof Error ? err.message : String(err);
    } finally {
      this.loaded = true;
    }
  }
}
