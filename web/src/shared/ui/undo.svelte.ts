import { announce } from '../a11y/live.svelte';
import { t } from '../i18n';

/**
 * Undo after a quick action in a widget (#867): ticking off or deleting an
 * item. The widget hides or changes the item at once and hands the server
 * call to the queue, which sends it when the toast goes away. Undo cancels
 * the call instead of trying to rebuild the item, so the item keeps its id,
 * history and position. One action waits at a time; a new one sends the
 * previous one first.
 */
export interface UndoableAction {
  /** What happened, e.g. "Milk removed". */
  text: string;
  /** Sends the change; a failure reverts it (the widget shows the error). */
  commit: () => Promise<unknown>;
  /** Shows the item as it was. */
  revert: () => void;
}

/** As long as the undo toast of edit mode. */
export const UNDO_MS = 8000;

export class UndoQueue {
  /** The toast on screen; key changes with every action so the toast is announced again. */
  current = $state<{ text: string; key: number } | null>(null);
  #pending: UndoableAction | null = null;
  #timer: ReturnType<typeof setTimeout> | undefined;
  #key = 0;

  constructor(private readonly ms = UNDO_MS) {}

  offer(action: UndoableAction): void {
    void this.flush();
    this.#pending = action;
    this.current = { text: action.text, key: ++this.#key };
    announce(action.text, { repeat: true });
    this.#timer = setTimeout(() => void this.flush(), this.ms);
  }

  /** Sends the waiting action now, e.g. before the page goes away or a bulk action. */
  flush(): Promise<void> {
    const action = this.#take();
    if (!action) return Promise.resolve();
    return action.commit().then(
      () => undefined,
      () => action.revert(),
    );
  }

  /** Takes the waiting action back; reports whether there was one. */
  undo(): boolean {
    const action = this.#take();
    if (!action) return false;
    action.revert();
    announce(t('undo.restored'), { repeat: true });
    return true;
  }

  #take(): UndoableAction | null {
    clearTimeout(this.#timer);
    const action = this.#pending;
    this.#pending = null;
    this.current = null;
    return action;
  }
}

/** The one queue of the app; the shell draws its toast. */
export const undoQueue = new UndoQueue();

/**
 * Offers a quick action on one item: id stays in pending (which the widget
 * applies when it draws its list) until send has run or the action is
 * taken back. send updates the widget's own list on success and rethrows
 * a failure after showing it. Keeping the change as an overlay, not in the
 * list itself, means a poll meanwhile cannot bring the item back.
 */
export function offerUndo(
  pending: Set<string>,
  id: string,
  text: string,
  send: () => Promise<unknown>,
  queue: UndoQueue = undoQueue,
): void {
  pending.add(id);
  queue.offer({
    text,
    commit: () => send().finally(() => pending.delete(id)),
    revert: () => pending.delete(id),
  });
}

if (typeof window !== 'undefined') {
  // A change the user made is sent before the page is closed or hidden,
  // rather than silently dropped.
  window.addEventListener('pagehide', () => void undoQueue.flush());
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'hidden') void undoQueue.flush();
  });
}
