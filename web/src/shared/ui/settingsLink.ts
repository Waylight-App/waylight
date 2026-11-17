import { getContext, setContext } from 'svelte';

/**
 * Lets a widget open the settings at one section (#512), e.g. the calendar
 * widget without a calendar opens Settings → Calendar. The shell provides
 * it; where nothing provides it (the browser extension's local mode, tests)
 * widgets simply show no button.
 */
export type OpenSettings = (section: string) => void;

const key = Symbol('open-settings');

/** Called once by the shell during its setup. */
export function provideOpenSettings(open: OpenSettings): void {
  setContext(key, open);
}

/** Called by a component during its setup; undefined without a provider. */
export function useOpenSettings(): OpenSettings | undefined {
  return getContext<OpenSettings | undefined>(key);
}
