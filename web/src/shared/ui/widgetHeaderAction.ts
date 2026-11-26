import { getContext, onDestroy, setContext } from 'svelte';
import { createSubscriber } from 'svelte/reactivity';
import type { IconName } from './icons/paths';

/**
 * The one add action a widget offers in its header, such as "New note"
 * (#809). The card draws it as a 32 px icon button next to its own controls.
 */
export type HeaderAction = { icon: IconName; label: string; onclick: () => void };

/** Reactive holder of one card's action; kept apart from the context so it can be tested on its own. */
export interface HeaderActionHolder {
  readonly action: HeaderAction | null;
  set(action: HeaderAction | null): void;
}

export function createHeaderAction(): HeaderActionHolder {
  let current: HeaderAction | null = null;
  let changed = () => {};
  const subscribe = createSubscriber((update) => {
    changed = update;
    return () => (changed = () => {});
  });
  return {
    get action() {
      subscribe();
      return current;
    },
    set(action) {
      current = action;
      changed();
    },
  };
}

const key = Symbol('widget-header-action');

/** Called by the card; it draws `action` in its header. */
export function provideHeaderAction(): { readonly action: HeaderAction | null } {
  const holder = createHeaderAction();
  setContext(key, holder);
  return holder;
}

/**
 * Called by a widget during its setup. `inHeader` is false outside a card
 * (detail panel, tests): the widget then shows the button in its own body,
 * so the action is never lost. The action is cleared when the widget goes away.
 */
export function useHeaderAction(): {
  inHeader: boolean;
  set: (action: HeaderAction | null) => void;
} {
  const holder = getContext<HeaderActionHolder | undefined>(key);
  if (!holder) return { inHeader: false, set: () => {} };
  onDestroy(() => holder.set(null));
  return { inHeader: true, set: (action) => holder.set(action) };
}
