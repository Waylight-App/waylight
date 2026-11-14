/**
 * A small channel between widgets and the dashboard: a widget asks to show
 * another widget (e.g. the briefing opens the to-do widget, #52) without
 * knowing the dashboard. The dashboard switches to a tab with a matching
 * widget and scrolls it into view.
 */
const EVENT = 'focus-widget';
const bus = new EventTarget();

/** Asks the dashboard to show a widget of one of the kinds (first match wins). */
export function requestWidgetFocus(kinds: readonly string[]): void {
  bus.dispatchEvent(new CustomEvent<string[]>(EVENT, { detail: [...kinds] }));
}

/** Subscribes to focus requests; returns the unsubscribe function. */
export function onWidgetFocusRequest(handler: (kinds: string[]) => void): () => void {
  const listener = (e: Event) => handler((e as CustomEvent<string[]>).detail);
  bus.addEventListener(EVENT, listener);
  return () => bus.removeEventListener(EVENT, listener);
}

/**
 * Picks the widget to show: a matching widget on the active tab first,
 * otherwise the first match on another tab, kinds in order of preference.
 */
export function pickWidget<W extends { id: string; kind: string }, T extends { id: string; widgets: W[] }>(
  tabs: readonly T[],
  activeId: string,
  kinds: readonly string[],
): { tabId: string; widgetId: string } | null {
  const ordered = [...tabs].sort((a, b) => (a.id === activeId ? -1 : b.id === activeId ? 1 : 0));
  for (const kind of kinds) {
    for (const tab of ordered) {
      const w = tab.widgets.find((x) => x.kind === kind);
      if (w) return { tabId: tab.id, widgetId: w.id };
    }
  }
  return null;
}

/**
 * Scrolls a widget into view, outlines it for 1.5 s and, if asked, moves the
 * keyboard focus to it (after adding or duplicating it, #273; the briefing's
 * links, #52). Reduced motion scrolls at once.
 */
export function highlightWidget(id: string, focus = false): void {
  const el = document.querySelector<HTMLElement>(`[data-widget-id="${CSS.escape(id)}"]`);
  if (!el) return;
  const reduce = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false;
  el.scrollIntoView({ behavior: reduce ? 'auto' : 'smooth', block: 'center' });
  // One outline at a time: a newer request replaces an older one.
  for (const other of document.querySelectorAll('.flash')) other.classList.remove('flash');
  el.classList.add('flash');
  setTimeout(() => el.classList.remove('flash'), 1500);
  if (focus) {
    if (!el.hasAttribute('tabindex')) el.setAttribute('tabindex', '-1');
    el.focus({ preventScroll: true });
  }
}
