import type { CommandSpec } from '../features/search/domain/commands';
import { type MessageKey, t } from '../shared/i18n';
import { type CatalogSection, OWN_SECTIONS, type SectionViewer, visibleSections } from './settingsCatalog';
import { registeredSections } from './settingsSections';

/**
 * The app commands of the search field (#476) as data. The search feature
 * matches and ranks them; the shell, as the composition root, supplies the
 * handlers. Commands that change something run directly only when they are
 * harmless and undoable (theme, tab); everything else opens a dialog.
 */
export const STATIC_COMMANDS: CommandSpec[] = [
  {
    id: 'theme.dark',
    terms: { de: ['dunkel', 'dunkelmodus', 'dunkles design'], en: ['dark', 'dark mode', 'night mode'] },
    role: 'member',
    handler: 'theme',
    arg: 'dark',
  },
  {
    id: 'theme.light',
    terms: { de: ['hell', 'hellmodus', 'helles design'], en: ['light', 'light mode', 'day mode'] },
    role: 'member',
    handler: 'theme',
    arg: 'light',
  },
  {
    id: 'theme.system',
    terms: { de: ['design automatisch', 'wie system'], en: ['system theme', 'auto theme'] },
    role: 'member',
    handler: 'theme',
    arg: 'system',
  },
  {
    // "theme dracula": the gallery is part of Appearance.
    id: 'theme.gallery',
    terms: { de: ['design', 'farbschema', 'aussehen'], en: ['theme', 'themes', 'appearance'] },
    role: 'member',
    handler: 'settings',
    arg: 'appearance',
    rest: true,
  },
  {
    id: 'widget.add',
    terms: { de: ['widget hinzufügen', 'neues widget'], en: ['add widget', 'new widget'] },
    role: 'member',
    handler: 'addWidget',
    rest: true,
  },
  {
    id: 'dashboard.edit',
    terms: {
      de: ['bearbeiten', 'anpassen', 'startseite bearbeiten'],
      en: ['edit', 'customise', 'customize'],
    },
    role: 'member',
    handler: 'edit',
    shortcut: 'E',
  },
  {
    id: 'settings.open',
    terms: { de: ['einstellungen'], en: ['settings', 'preferences'] },
    role: 'everyone',
    handler: 'settings',
    weight: 1.1,
  },
  {
    id: 'shortcuts.help',
    terms: { de: ['tastenkürzel', 'tastatur'], en: ['keyboard shortcuts', 'shortcuts', 'hotkeys'] },
    role: 'everyone',
    handler: 'shortcuts',
    shortcut: '?',
  },
  {
    id: 'tour.start',
    terms: { de: ['tour starten', 'rundgang', 'einführung'], en: ['start tour', 'tour', 'introduction'] },
    role: 'member',
    handler: 'tour',
  },
  {
    id: 'session.logout',
    terms: { de: ['abmelden', 'ausloggen'], en: ['log out', 'logout', 'sign out'] },
    role: 'everyone',
    handler: 'logout',
  },
];

const SETTINGS_WORDS = ['einstellungen', 'einstellung', 'settings', 'setting'];
const TAB_WORDS = ['tab', 'reiter', 'seite', 'page'];

/**
 * A deep link for every settings section the account may open, built from
 * the settings navigation (#261) so the list is not kept twice.
 */
export function settingsCommands(sections: readonly CatalogSection[]): CommandSpec[] {
  return sections.map((s) => {
    const name = s.title?.() ?? t(s.label);
    return {
      id: `settings.${s.id}`,
      terms: { de: ['einstellungen', name], en: ['settings', s.id.replace(/-/g, ' ')] },
      role: 'everyone',
      handler: 'settings',
      arg: s.id,
      requires: SETTINGS_WORDS,
    };
  });
}

/** "tab arbeit" switches to the tab; harmless, so it runs directly. */
export function tabCommands(tabs: readonly { id: string; name: string }[]): CommandSpec[] {
  return tabs.map((tab) => ({
    id: `tab.${tab.id}`,
    terms: { de: ['tab', tab.name], en: ['tab'] },
    role: 'everyone',
    handler: 'tab',
    arg: tab.id,
    requires: TAB_WORDS,
  }));
}

const TITLES: Record<string, MessageKey> = {
  'theme.dark': 'commands.theme_dark',
  'theme.light': 'commands.theme_light',
  'theme.system': 'commands.theme_system',
  'theme.gallery': 'commands.theme_gallery',
  'widget.add': 'commands.widget_add',
  'dashboard.edit': 'commands.dashboard_edit',
  'settings.open': 'commands.settings',
  'shortcuts.help': 'commands.shortcuts_help',
  'tour.start': 'commands.tour_start',
  'session.logout': 'commands.session_logout',
};

/** Title of a command row, in the current language. */
export function commandTitle(
  spec: CommandSpec,
  names: { section: (id: string) => string; tab: (id: string) => string },
): string {
  const fixed = TITLES[spec.id];
  if (fixed) return t(fixed);
  if (spec.handler === 'tab') return t('commands.tab', { name: names.tab(spec.arg ?? '') });
  return t('commands.settings_section', { name: names.section(spec.arg ?? '') });
}

/**
 * What the shell, as the composition root, gives the search to run app
 * commands. A missing handler means the command is not offered, e.g. no
 * theme commands while the organisation locks the theme (#305).
 */
export interface CommandHandlers {
  theme?: (mode: 'light' | 'dark' | 'system') => void;
  settings: (section?: string) => void;
  addWidget?: (query: string) => void;
  edit?: () => void;
  shortcuts: () => void;
  tour?: () => void;
  logout: () => void;
  tab: (id: string) => void;
  tabs: () => readonly { id: string; name: string }[];
  /** Who looks at the settings, to offer only the sections they may open. */
  sectionViewer: () => SectionViewer;
}

/** The commands the handlers support, with titles and a runner. */
export function buildCommands(h: CommandHandlers) {
  const sections = () => visibleSections([...OWN_SECTIONS, ...registeredSections], h.sectionViewer());
  const supported: Record<string, boolean> = {
    theme: !!h.theme,
    addWidget: !!h.addWidget,
    edit: !!h.edit,
    tour: !!h.tour,
  };
  const specs = (): CommandSpec[] => [
    ...STATIC_COMMANDS.filter((c) => supported[c.handler] ?? true),
    ...settingsCommands(sections()),
    ...tabCommands(h.tabs()),
  ];
  const names = {
    section: (id: string) => {
      const s = sections().find((x) => x.id === id);
      return s ? (s.title?.() ?? t(s.label)) : id;
    },
    tab: (id: string) => h.tabs().find((tab) => tab.id === id)?.name ?? id,
  };
  const run = (spec: CommandSpec, rest: string) => {
    switch (spec.handler) {
      case 'theme':
        return h.theme?.(spec.arg as 'light' | 'dark' | 'system');
      case 'settings':
        return h.settings(spec.arg);
      case 'addWidget':
        return h.addWidget?.(rest);
      case 'edit':
        return h.edit?.();
      case 'shortcuts':
        return h.shortcuts();
      case 'tour':
        return h.tour?.();
      case 'logout':
        return h.logout();
      case 'tab':
        return h.tab(spec.arg ?? '');
    }
  };
  return { specs, title: (spec: CommandSpec) => commandTitle(spec, names), run };
}
