import { getContext, setContext } from 'svelte';

/**
 * The widget kinds whose tile opens the guided setup dialog itself (it takes
 * its turn in the walk below). The others are set up in the widget settings.
 */
export const GUIDED_SETUP_KINDS: readonly string[] = ['waste', 'immich', 'media', 'smarthome', 'proxmox'];

/**
 * The setup walk of the onboarding done page: the widget kinds that still
 * need setting up, one after another. The front of the queue is the widget
 * whose setup is open now; ending its dialog, with or without saving, moves
 * on, so "Later" stays possible at every step.
 */
export class SetupWalk {
  private queue = $state<string[]>([]);
  /** Called once when the last setup is over. */
  private finished: (() => void) | undefined;

  /** The kind whose setup is up, if a walk is running. */
  get current(): string | undefined {
    return this.queue[0];
  }

  get active(): boolean {
    return this.queue.length > 0;
  }

  start(kinds: readonly string[], finished?: () => void): void {
    this.queue = kinds.filter((kind, i) => kinds.indexOf(kind) === i);
    this.finished = finished;
  }

  next(): void {
    this.queue = this.queue.slice(1);
    if (this.queue.length === 0) {
      const done = this.finished;
      this.finished = undefined;
      done?.();
    }
  }

  /** Lets a caller that wants to run something after the walk register it. */
  whenFinished(done: () => void): void {
    this.finished = done;
  }
}

/** The one walk of the app; the wizard starts it and the dashboard follows it. */
export const setupWalk = new SetupWalk();

const kindKey = Symbol('widget-kind');

/** Called by the card around a widget, so its content knows which kind it is. */
export function provideWidgetKind(kind: () => string | undefined): void {
  setContext(kindKey, kind);
}

/**
 * Called by the setup of a widget: runs `open` when the walk reaches this
 * widget and it can show its setup (`ready`). Returns what to call when the
 * setup ends, so the walk moves on.
 */
export function takeSetupTurn(ready: () => boolean, open: () => void): () => void {
  const kind = getContext<(() => string | undefined) | undefined>(kindKey);
  let mine = false;
  $effect(() => {
    if (!mine && ready() && kind?.() === setupWalk.current) {
      mine = true;
      open();
    }
  });
  return () => {
    if (!mine) return;
    mine = false;
    setupWalk.next();
  };
}
