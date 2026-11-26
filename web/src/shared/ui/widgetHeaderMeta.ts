import { getContext, onDestroy, setContext } from 'svelte';
import { createSubscriber } from 'svelte/reactivity';

/**
 * Short meta text a widget shows at the right of its header, such as
 * "3 due today" (#809). The card draws the header, the widget knows the
 * text, so the widget hands it up through this channel.
 */
export type HeaderMeta = { text: string; tone?: 'neutral' | 'warning' | 'danger' };

/** Reactive holder of one card's meta; kept apart from the context so it can be tested on its own. */
export interface HeaderMetaHolder {
  readonly meta: HeaderMeta | null;
  set(meta: HeaderMeta | null): void;
}

export function createHeaderMeta(): HeaderMetaHolder {
  let current: HeaderMeta | null = null;
  let changed = () => {};
  const subscribe = createSubscriber((update) => {
    changed = update;
    return () => (changed = () => {});
  });
  return {
    get meta() {
      subscribe();
      return current;
    },
    set(meta) {
      const next = meta && meta.text ? meta : null;
      if (next?.text === current?.text && next?.tone === current?.tone) return;
      current = next;
      changed();
    },
  };
}

const key = Symbol('widget-header-meta');

/** Called by the card; it passes `meta` on to its WidgetHeader. */
export function provideHeaderMeta(): { readonly meta: HeaderMeta | null } {
  const holder = createHeaderMeta();
  setContext(key, holder);
  return holder;
}

/**
 * Called by a widget during its setup. Outside a card (detail panel, tests)
 * the setter does nothing, so widgets work the same everywhere. The meta is
 * cleared when the widget goes away, so a swapped view leaves no stale text.
 */
export function useHeaderMeta(): (meta: HeaderMeta | null) => void {
  const holder = getContext<HeaderMetaHolder | undefined>(key);
  if (!holder) return () => {};
  onDestroy(() => holder.set(null));
  return (meta) => holder.set(meta);
}
