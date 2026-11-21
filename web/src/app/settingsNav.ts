import type { MessageKey } from '../shared/i18n';
import type { IconName } from '../shared/ui/icons/paths';

/**
 * The eight task areas of the settings (#600), in display order. Sections
 * are sorted by what someone wants to do, not by the technology behind
 * them, so an integration and its widget settings are found in one place.
 */
export type SettingsArea =
  'account' | 'appearance' | 'notifications' | 'dashboard' | 'connections' | 'people' | 'data' | 'system';

export interface AreaInfo {
  id: SettingsArea;
  label: MessageKey;
  /** One line on the start page that says what the area holds. */
  hint: MessageKey;
  icon: IconName;
  /** Words that lead to the area in the settings search, English and German (#601). */
  keywords: readonly string[];
}

export const SETTINGS_AREAS: readonly AreaInfo[] = [
  {
    id: 'account',
    label: 'settings.area.account',
    hint: 'settings.area.account_hint',
    icon: 'user',
    keywords: ['profile', 'security', 'profil', 'konto', 'sicherheit'],
  },
  {
    id: 'appearance',
    label: 'settings.area.appearance',
    hint: 'settings.area.appearance_hint',
    icon: 'sun',
    keywords: ['look', 'design', 'aussehen', 'darstellung'],
  },
  {
    id: 'notifications',
    label: 'settings.area.notifications',
    hint: 'settings.area.notifications_hint',
    icon: 'bell',
    keywords: ['push', 'alerts', 'mitteilungen'],
  },
  {
    id: 'dashboard',
    label: 'settings.area.dashboard',
    hint: 'settings.area.dashboard_hint',
    icon: 'grid',
    keywords: ['widgets', 'start page', 'startseite', 'inhalte'],
  },
  {
    id: 'connections',
    label: 'settings.area.connections',
    hint: 'settings.area.connections_hint',
    icon: 'plug',
    keywords: ['integrations', 'services', 'apps', 'integrationen', 'dienste', 'zugänge'],
  },
  {
    id: 'people',
    label: 'settings.area.people',
    hint: 'settings.area.people_hint',
    icon: 'users',
    keywords: ['users', 'family', 'team', 'benutzer', 'familie', 'personen'],
  },
  {
    id: 'data',
    label: 'settings.area.data',
    hint: 'settings.area.data_hint',
    icon: 'archive',
    keywords: ['backup', 'export', 'import', 'sicherung', 'daten'],
  },
  {
    id: 'system',
    label: 'settings.area.system',
    hint: 'settings.area.system_hint',
    icon: 'info',
    keywords: ['about', 'version', 'licence', 'license', 'über', 'lizenz'],
  },
];

/**
 * Sections without an area are integrations: most new modules connect a
 * service, so a feature only names its area when it is something else.
 */
export const DEFAULT_AREA: SettingsArea = 'connections';

const AREA_IDS = new Set<string>(SETTINGS_AREAS.map((a) => a.id));

/** Whether a value names an area; plugins pass plain strings. */
export function isSettingsArea(value: unknown): value is SettingsArea {
  return typeof value === 'string' && AREA_IDS.has(value);
}

/** The area a section belongs to. */
export const areaOf = (s: { area?: SettingsArea }): SettingsArea => s.area ?? DEFAULT_AREA;

export interface NavArea<T> extends AreaInfo {
  sections: T[];
}

/**
 * Sorts sections into their areas, in area order; areas without a visible
 * section are left out, so nobody sees an area they cannot use. Within an
 * area the given order is kept, except connections, which are sorted by
 * their (translated) name because there are many.
 */
export function groupByArea<T extends { area?: SettingsArea }>(
  sections: readonly T[],
  name: (s: T) => string,
  locale?: string,
): NavArea<T>[] {
  return SETTINGS_AREAS.map((a) => {
    const own = sections.filter((s) => areaOf(s) === a.id);
    if (a.id === 'connections') own.sort((x, y) => name(x).localeCompare(name(y), locale));
    return { ...a, sections: own };
  }).filter((a) => a.sections.length > 0);
}

/** Index to focus after a key in a vertical list of count entries, or null. */
export function moveFocus(key: string, index: number, count: number): number | null {
  if (count === 0) return null;
  switch (key) {
    case 'ArrowDown':
      return (index + 1) % count;
    case 'ArrowUp':
      return (index - 1 + count) % count;
    case 'Home':
      return 0;
    case 'End':
      return count - 1;
  }
  return null;
}

/** How many shortcuts "Recently changed" offers. */
export const RECENT_LIMIT = 3;

/** The recent list after a change in section id: newest first, no repeats. */
export function pushRecent(recent: readonly string[], id: string, limit = RECENT_LIMIT): string[] {
  return [id, ...recent.filter((r) => r !== id)].slice(0, limit);
}
