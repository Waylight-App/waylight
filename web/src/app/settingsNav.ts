import type { MessageKey } from '../shared/i18n';

/** Topics of the settings navigation (#261), in display order. */
export type SettingsGroup =
  'general' | 'content' | 'integrations' | 'data' | 'account' | 'organisation' | 'about';

export const SETTINGS_GROUPS: readonly { id: SettingsGroup; label: MessageKey }[] = [
  { id: 'general', label: 'settings.group.general' },
  { id: 'content', label: 'settings.group.content' },
  { id: 'integrations', label: 'settings.group.integrations' },
  { id: 'data', label: 'settings.group.data' },
  { id: 'account', label: 'settings.group.account' },
  // Sections of the ee/ plugin (ADR 0021); empty without it, so not shown.
  { id: 'organisation', label: 'settings.group.organisation' },
  { id: 'about', label: 'settings.group.about' },
];

/** Sections without a group are integrations (most new modules are). */
export const DEFAULT_GROUP: SettingsGroup = 'integrations';

export interface NavGroup<T> {
  id: SettingsGroup;
  label: MessageKey;
  sections: T[];
}

/**
 * Sorts sections into their groups, in group order; empty groups are left
 * out. Within a group the given order is kept, except integrations, which
 * are sorted by their (translated) name because there are many.
 */
export function groupSections<T extends { group?: SettingsGroup }>(
  sections: readonly T[],
  name: (s: T) => string,
  locale?: string,
): NavGroup<T>[] {
  return SETTINGS_GROUPS.map((g) => {
    const own = sections.filter((s) => (s.group ?? DEFAULT_GROUP) === g.id);
    if (g.id === 'integrations') own.sort((a, b) => name(a).localeCompare(name(b), locale));
    return { ...g, sections: own };
  }).filter((g) => g.sections.length > 0);
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
