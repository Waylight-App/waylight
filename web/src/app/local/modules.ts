/**
 * What the extension's local mode offers (ADR 0019): the modules that work
 * without a server, and the widget kinds they bring. Everything else is
 * hidden, so the widget picker, the settings and the search follow this list
 * without knowing about the local mode.
 */

/** Backend module names that have a local implementation. */
export const localModules: readonly string[] = [
  'dashboard',
  'appearance',
  'bookmarks',
  'notes',
  'todos',
  'countdown',
  'focus',
  'habits',
  'weather',
];

/**
 * Widget kinds of those modules, plus the clock, which is drawn in the
 * browser and has no module of its own. The rain radar of the weather module
 * stays out: its tiles come from a third-party map server.
 */
export const localWidgetKinds: readonly string[] = [
  'clock',
  'bookmarks',
  'quicklinks',
  'notes',
  'todos',
  'countdown',
  'focus',
  'habits',
  'weather',
];

/**
 * Data kind per module, the same values the server reports (ADR 0018). The
 * owner is the only account in local mode, so they only matter for code that
 * treats the kinds differently, and should match the server's.
 */
export const localDataKinds: Readonly<Record<string, string>> = {
  dashboard: 'personal-content',
  appearance: 'personal-settings',
  bookmarks: 'personal-content',
  notes: 'personal-content',
  todos: 'personal-content',
  countdown: 'personal-content',
  focus: 'personal-content',
  habits: 'personal-content',
  weather: 'instance',
};
