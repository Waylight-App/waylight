import { getContext, setContext } from 'svelte';
import type { IconName } from './icons/paths';

/**
 * The card around a widget hands its icon down, so the shared "not set up"
 * state shows the symbol of its widget without every widget passing it on.
 */
const key = Symbol('widget-icon');

/** Called by the card; the getter keeps the icon reactive. */
export function provideWidgetIcon(icon: () => IconName | undefined): void {
  setContext(key, icon);
}

/** The icon of the surrounding widget, if there is one. */
export function useWidgetIcon(): IconName | undefined {
  return getContext<(() => IconName | undefined) | undefined>(key)?.();
}
