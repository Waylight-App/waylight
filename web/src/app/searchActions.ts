/**
 * Commands that plugins add to the start page search (#453). The loader
 * fills the registry once per page; the search panel asks it for the actions
 * that match the typed text. Kept free of Svelte so the matching is tested
 * on its own.
 */
export interface SearchAction {
  /** Stable row id, unique across plugins. */
  id: string;
  /** Lower-case keywords; see matches(). */
  keywords: string[];
  title(): string;
  detail(): string;
  /** Runs the action and returns the confirmation to show. */
  run(): Promise<string>;
}

/** Filled by the plugin loader; empty without plugins. */
export const searchActions: SearchAction[] = [];

/** Shortest text that offers an action by a keyword's beginning. */
export const MIN_PREFIX = 2;

/** Whether the typed text offers an action: equal to a keyword or its start. */
export function matches(action: Pick<SearchAction, 'keywords'>, query: string): boolean {
  const q = query.trim().toLocaleLowerCase();
  if (q.length < MIN_PREFIX) return false;
  return action.keywords.some((k) => k === q || k.startsWith(q));
}

/** The actions offered for the typed text, in registration order. */
export function matchActions(actions: readonly SearchAction[], query: string, limit = 3): SearchAction[] {
  return actions.filter((a) => matches(a, query)).slice(0, limit);
}
