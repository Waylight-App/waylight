import { getContext, setContext } from 'svelte';
import { createSubscriber } from 'svelte/reactivity';
import type { Freshness } from '../freshness/status';

/**
 * Data freshness of the widget in a card (#94). The wrapper around the
 * widget knows it, the header draws it: an overlay on the content covered
 * the widget's text on hover and focus (#851), so the badge sits in the
 * header row next to the meta text instead.
 */
export type WidgetStatus = {
  freshness: Freshness;
  fetchedAt: number | null;
  lastError: string | null;
  now: number;
  refresh: () => void;
};

/** Reactive holder of one card's status; kept apart from the context so it can be tested on its own. */
export interface WidgetStatusHolder {
  readonly status: WidgetStatus | null;
  set(status: WidgetStatus | null): void;
  /** Edit mode hides the badge: the header row holds the edit controls then. */
  hidden: boolean;
}

export function createWidgetStatus(): WidgetStatusHolder {
  let current: WidgetStatus | null = null;
  let hidden = false;
  let changed = () => {};
  const subscribe = createSubscriber((update) => {
    changed = update;
    return () => (changed = () => {});
  });
  return {
    get status() {
      subscribe();
      return hidden ? null : current;
    },
    set(status) {
      current = status;
      changed();
    },
    get hidden() {
      return hidden;
    },
    set hidden(value) {
      if (value === hidden) return;
      hidden = value;
      changed();
    },
  };
}

const key = Symbol('widget-status');

/** Called by the card; the headers inside it read the status. */
export function provideWidgetStatus(): WidgetStatusHolder {
  const holder = createWidgetStatus();
  setContext(key, holder);
  return holder;
}

/** The card's holder, or undefined outside a card (detail panel, tests). */
export function widgetStatus(): WidgetStatusHolder | undefined {
  return getContext<WidgetStatusHolder | undefined>(key);
}
