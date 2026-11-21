import type { SavedPlace } from '../../../shared/places/places';

/** A widget that takes its location from a place. */
export interface PlaceUse {
  widgetId: string;
  tabId: string;
  kind: string;
}

/** A saved place with the widgets on the account's own tabs that use it (#591). */
export interface Place extends SavedPlace {
  usedBy: PlaceUse[];
}

/** Maximum name length; mirrors the backend. */
export const MAX_NAME_LENGTH = 80;

/** A search result to save as a place. */
export interface PlaceCandidate {
  name: string;
  label: string;
  lat: number;
  lon: number;
}

/** Validates a place name and returns an error message key or null. */
export function validateName(name: string): 'places.name_empty' | 'places.name_too_long' | null {
  const n = name.trim();
  if (n === '') return 'places.name_empty';
  if ([...n].length > MAX_NAME_LENGTH) return 'places.name_too_long';
  return null;
}

/**
 * Describes where a place is used, e.g. "Weather (Home), Fuel prices": one
 * entry per widget with its tab, so the delete question can name them.
 */
export function usageList(
  uses: readonly PlaceUse[],
  widgetName: (kind: string) => string,
  tabName: (tabId: string) => string | undefined,
): string[] {
  return uses.map((u) => {
    const tab = tabName(u.tabId);
    return tab ? `${widgetName(u.kind)} (${tab})` : widgetName(u.kind);
  });
}
